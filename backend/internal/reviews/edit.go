package reviews

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func editsToPatch(ctx context.Context, repository string, edits []Edit) (string, error) {
	if len(edits) == 0 || len(edits) > 5 {
		return "", errors.New("proposal must contain between one and five edits")
	}
	workspace, err := os.MkdirTemp(filepath.Dir(repository), "proposal-edit-")
	if err != nil {
		return "", errors.New("could not create isolated edit workspace")
	}
	defer os.RemoveAll(workspace)
	copyDir := filepath.Join(workspace, "source")
	if _, err = runGit(ctx, workspace, "-c", "protocol.file.allow=always", "clone", "--quiet", "--no-hardlinks", "--", repository, copyDir); err != nil {
		return "", errors.New("could not isolate repository for patch generation")
	}

	for _, edit := range edits {
		rel := filepath.ToSlash(filepath.Clean(strings.TrimSpace(edit.File)))
		if rel == "" || rel == "." || rel == ".." || rel == ".git" || strings.HasPrefix(rel, "../") || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".git/") {
			return "", errors.New("edit contains an unsafe file path")
		}
		if edit.Old == "" || edit.Old == edit.New || len(edit.Old)+len(edit.New) > 64*1024 {
			return "", errors.New("edit must replace a bounded non-empty exact value")
		}
		path := filepath.Join(copyDir, filepath.FromSlash(rel))
		resolved, resolveErr := filepath.EvalSymlinks(path)
		if resolveErr != nil || resolved != path {
			return "", errors.New("edit target crosses a symbolic link")
		}
		info, statErr := os.Lstat(path)
		if statErr != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("edit target is not a regular tracked file")
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return "", errors.New("edit target could not be read")
		}
		if strings.Count(string(data), edit.Old) != 1 {
			return "", fmt.Errorf("exact old text was not found once in %s", rel)
		}
		updated := strings.Replace(string(data), edit.Old, edit.New, 1)
		if writeErr := os.WriteFile(path, []byte(updated), info.Mode().Perm()); writeErr != nil {
			return "", errors.New("isolated edit could not be written")
		}
	}

	patch, err := runGit(ctx, copyDir, "diff", "--no-ext-diff", "--binary", "--")
	if err != nil {
		return "", errors.New("could not generate deterministic patch")
	}
	patch = strings.TrimSpace(patch)
	if patch == "" {
		return "", errors.New("edits did not produce a patch")
	}
	return patch, nil
}
