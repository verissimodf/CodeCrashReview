package reviews

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	branchPattern = regexp.MustCompile(`^[A-Za-z0-9._/-]{1,200}$`)
	secretPattern = regexp.MustCompile(`(?i)(api[_-]?key|token|secret|password)(\s*[=:]\s*)[^\s"']{8,}`)
)

type sourceManager struct {
	workspaceRoot string
	localRoots    []string
	maxFiles      int
	maxBytes      int
}

func newSourceManager(workspaceRoot, roots string, maxFiles, maxBytes int) *sourceManager {
	var allowed []string
	for _, root := range strings.Split(roots, ",") {
		if clean := strings.TrimSpace(root); clean != "" {
			allowed = append(allowed, filepath.Clean(clean))
		}
	}
	return &sourceManager{filepath.Clean(workspaceRoot), allowed, maxFiles, maxBytes}
}

func (m *sourceManager) acquire(ctx context.Context, job claimedReview) (Snapshot, error) {
	if !branchPattern.MatchString(job.Branch) || strings.Contains(job.Branch, "..") {
		return Snapshot{}, errors.New("invalid repository branch")
	}
	jobDir := filepath.Join(m.workspaceRoot, job.ID.String())
	if err := os.MkdirAll(jobDir, 0700); err != nil {
		return Snapshot{}, fmt.Errorf("create review workspace: %w", err)
	}
	destination := filepath.Join(jobDir, "source")
	if _, err := os.Stat(destination); err == nil {
		if err := os.RemoveAll(destination); err != nil {
			return Snapshot{}, fmt.Errorf("reset review workspace: %w", err)
		}
	}

	var cloneSource string
	switch job.SourceType {
	case "github":
		canonical, err := validateGitHubURL(job.Source)
		if err != nil {
			return Snapshot{}, err
		}
		cloneSource = canonical
	case "local":
		canonical, err := m.validateLocalPath(job.Source)
		if err != nil {
			return Snapshot{}, err
		}
		cloneSource = canonical
	default:
		return Snapshot{}, errors.New("unsupported repository source")
	}

	args := []string{"-c", "protocol.file.allow=never", "-c", "http.followRedirects=false", "clone", "--no-tags", "--single-branch", "--branch", job.Branch}
	if job.SourceType == "github" {
		args = append(args, "--depth", "1")
	} else {
		args[1] = "protocol.file.allow=always"
		args = append(args, "--no-hardlinks")
	}
	args = append(args, "--", cloneSource, destination)
	if _, err := runGit(ctx, jobDir, args...); err != nil {
		return Snapshot{}, fmt.Errorf("clone repository: %w", err)
	}
	sha, err := runGit(ctx, destination, "rev-parse", "HEAD")
	if err != nil {
		return Snapshot{}, fmt.Errorf("read snapshot revision: %w", err)
	}
	contextText, files, bytesRead, err := m.buildContext(destination, job.FocusPaths)
	if err != nil {
		return Snapshot{}, err
	}
	if files == 0 {
		return Snapshot{}, errors.New("repository snapshot contains no reviewable text files")
	}
	return Snapshot{Dir: destination, SHA: strings.TrimSpace(sha), Context: contextText, FileCount: files, ByteCount: bytesRead}, nil
}

func validateGitHubURL(raw string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme != "https" || !strings.EqualFold(u.Hostname(), "github.com") || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return "", errors.New("GitHub source must be an HTTPS github.com repository URL without credentials or query parameters")
	}
	parts := strings.Split(strings.Trim(strings.TrimSuffix(u.Path, ".git"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" || parts[0] == "." || parts[1] == "." {
		return "", errors.New("GitHub source must identify exactly one owner and repository")
	}
	for _, part := range parts {
		if !regexp.MustCompile(`^[A-Za-z0-9_.-]+$`).MatchString(part) {
			return "", errors.New("GitHub owner or repository contains unsupported characters")
		}
	}
	return "https://github.com/" + parts[0] + "/" + parts[1] + ".git", nil
}

func (m *sourceManager) validateLocalPath(raw string) (string, error) {
	if len(m.localRoots) == 0 {
		return "", errors.New("local repositories are disabled")
	}
	requested := filepath.Clean(strings.TrimSpace(raw))
	if !filepath.IsAbs(requested) {
		requested = filepath.Join(m.localRoots[0], requested)
	}
	resolved, err := filepath.EvalSymlinks(requested)
	if err != nil {
		return "", errors.New("local repository is unavailable")
	}
	for _, root := range m.localRoots {
		resolvedRoot, err := filepath.EvalSymlinks(root)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(resolvedRoot, resolved)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			if info, statErr := os.Stat(filepath.Join(resolved, ".git")); statErr == nil && info.IsDir() {
				return resolved, nil
			}
			return "", errors.New("local source must be a Git repository")
		}
	}
	return "", errors.New("local repository is outside the configured roots")
}

func (m *sourceManager) buildContext(root string, focus []string) (string, int, int64, error) {
	var paths []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if rel != "." && skippedDirectory(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || sensitiveFile(entry.Name()) || !matchesFocus(filepath.ToSlash(rel), focus) {
			return nil
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() || info.Size() > 128*1024 {
			return nil
		}
		paths = append(paths, rel)
		return nil
	})
	if err != nil {
		return "", 0, 0, fmt.Errorf("enumerate repository: %w", err)
	}
	sort.Slice(paths, func(i, j int) bool { return pathPriority(paths[i]) < pathPriority(paths[j]) })

	var out strings.Builder
	var total int64
	count := 0
	for _, rel := range paths {
		if count >= m.maxFiles || total >= int64(m.maxBytes) {
			break
		}
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil || bytes.IndexByte(data, 0) >= 0 {
			continue
		}
		remaining := int64(m.maxBytes) - total
		if int64(len(data)) > remaining {
			continue
		}
		content := secretPattern.ReplaceAllString(string(data), "$1$2[REDACTED]")
		fmt.Fprintf(&out, "\n--- FILE: %s ---\n%s\n--- END FILE ---\n", filepath.ToSlash(rel), content)
		total += int64(len(data))
		count++
	}
	return out.String(), count, total, nil
}

func sensitiveFile(name string) bool {
	name = strings.ToLower(name)
	return strings.HasPrefix(name, ".env") || strings.HasPrefix(name, "id_rsa") || strings.HasPrefix(name, "id_ed25519") || strings.Contains(name, "credentials") || strings.HasSuffix(name, ".pem") || strings.HasSuffix(name, ".key") || strings.HasSuffix(name, ".p12") || strings.HasSuffix(name, ".pfx")
}

func skippedDirectory(name string) bool {
	switch name {
	case ".git", "node_modules", "vendor", "dist", "build", ".next", ".svelte-kit", "coverage", "target", "bin", "obj", ".venv", "venv", "__pycache__":
		return true
	}
	return false
}

func matchesFocus(path string, focus []string) bool {
	if len(focus) == 0 {
		return true
	}
	for _, prefix := range focus {
		clean := strings.Trim(filepath.ToSlash(filepath.Clean(prefix)), "/")
		if clean != "" && clean != "." && (path == clean || strings.HasPrefix(path, clean+"/")) {
			return true
		}
	}
	return false
}

func pathPriority(path string) string {
	base := strings.ToLower(filepath.Base(path))
	priority := "5"
	if strings.HasPrefix(base, "readme") || base == "go.mod" || base == "package.json" || base == "pyproject.toml" || base == "cargo.toml" {
		priority = "0"
	} else if strings.Contains(path, "test") || strings.Contains(path, "spec") {
		priority = "3"
	} else if strings.HasSuffix(base, ".go") || strings.HasSuffix(base, ".py") || strings.HasSuffix(base, ".ts") || strings.HasSuffix(base, ".js") || strings.HasSuffix(base, ".java") || strings.HasSuffix(base, ".rs") {
		priority = "1"
	} else if strings.HasSuffix(base, ".yml") || strings.HasSuffix(base, ".yaml") || strings.HasSuffix(base, ".json") || strings.HasSuffix(base, ".toml") {
		priority = "2"
	}
	return priority + "/" + filepath.ToSlash(path)
}

func runGit(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	tempHome := filepath.Join(dir, ".git-home")
	if err := os.MkdirAll(tempHome, 0700); err != nil {
		return "", err
	}
	cmd.Env = []string{
		"PATH=/usr/local/bin:/usr/bin:/bin",
		"HOME=" + tempHome,
		"GIT_TERMINAL_PROMPT=0",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"LANG=C.UTF-8",
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		message := strings.TrimSpace(stderr.String())
		if len(message) > 500 {
			message = message[:500]
		}
		return "", fmt.Errorf("git command failed: %s", message)
	}
	return stdout.String(), nil
}

func digestText(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
