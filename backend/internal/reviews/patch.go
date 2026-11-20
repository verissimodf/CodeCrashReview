package reviews

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func validatePatch(ctx context.Context, snapshotDir, patch string) (bool, string) {
	patch = strings.TrimSpace(patch)
	if patch == "" {
		return false, "proposal did not include a patch"
	}
	if len(patch) > maxPatchBytes {
		return false, "patch exceeds 256 KiB"
	}
	if err := validatePatchPaths(patch); err != nil {
		return false, err.Error()
	}
	file, err := os.CreateTemp(filepath.Dir(snapshotDir), "proposal-*.patch")
	if err != nil {
		return false, "could not create patch validation file"
	}
	name := file.Name()
	defer os.Remove(name)
	if _, err = file.WriteString(patch + "\n"); err != nil {
		file.Close()
		return false, "could not write patch validation file"
	}
	if err = file.Close(); err != nil {
		return false, "could not close patch validation file"
	}
	if _, err = runGit(ctx, snapshotDir, "apply", "--check", "--recount", "--whitespace=error-all", "--", name); err != nil {
		return false, "git apply check failed"
	}
	return true, ""
}

func validatePatchPaths(patch string) error {
	seenHeader := false
	for _, line := range strings.Split(patch, "\n") {
		if strings.HasPrefix(line, "diff --git ") {
			seenHeader = true
			parts := strings.Fields(line)
			if len(parts) != 4 {
				return errors.New("invalid unified diff header")
			}
			for _, raw := range parts[2:] {
				path := strings.TrimPrefix(strings.TrimPrefix(raw, "a/"), "b/")
				clean := filepath.ToSlash(filepath.Clean(path))
				if filepath.IsAbs(path) || clean == ".." || strings.HasPrefix(clean, "../") || clean == ".git" || strings.HasPrefix(clean, ".git/") {
					return fmt.Errorf("unsafe patch path: %s", path)
				}
			}
		}
	}
	if !seenHeader {
		return errors.New("patch must be a git unified diff")
	}
	return nil
}
