package reviews

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestValidateGitHubURL(t *testing.T) {
	valid, err := validateGitHubURL("https://github.com/example/project")
	if err != nil || valid != "https://github.com/example/project.git" {
		t.Fatalf("valid URL rejected: %q %v", valid, err)
	}
	for _, value := range []string{
		"http://github.com/example/project",
		"https://user:token@github.com/example/project",
		"https://github.example.com/example/project",
		"https://github.com/example/project/extra",
		"https://github.com/example/project?token=secret",
	} {
		if _, err := validateGitHubURL(value); err == nil {
			t.Errorf("unsafe URL accepted: %s", value)
		}
	}
}

func TestValidateLocalPathStaysInsideRoot(t *testing.T) {
	root := t.TempDir()
	repo := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0700); err != nil {
		t.Fatal(err)
	}
	manager := newSourceManager(t.TempDir(), root, 10, 1024)
	if got, err := manager.validateLocalPath("repo"); err != nil || got != repo {
		t.Fatalf("allowed local repository rejected: %q %v", got, err)
	}
	outside := t.TempDir()
	if err := os.MkdirAll(filepath.Join(outside, ".git"), 0700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "escape")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.validateLocalPath(link); err == nil {
		t.Fatal("symlink escape was accepted")
	}
}

func TestValidatePatch(t *testing.T) {
	repo := t.TempDir()
	if _, err := runGit(context.Background(), repo, "init"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(repo, "file.txt"), []byte("old\n"), 0600); err != nil {
		t.Fatal(err)
	}
	patch := "diff --git a/file.txt b/file.txt\n--- a/file.txt\n+++ b/file.txt\n@@ -1 +1 @@\n-old\n+new"
	if ok, message := validatePatch(context.Background(), repo, patch); !ok {
		t.Fatalf("valid patch rejected: %s", message)
	}
	unsafe := "diff --git a/../../etc/passwd b/../../etc/passwd\n--- a/../../etc/passwd\n+++ b/../../etc/passwd"
	if ok, _ := validatePatch(context.Background(), repo, unsafe); ok {
		t.Fatal("path traversal patch was accepted")
	}
}
