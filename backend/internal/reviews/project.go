package reviews

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"
)

func projectFiles(ctx context.Context, root string) (map[string][]byte, error) {
	tracked, err := runGit(ctx, root, "ls-files", "-z")
	if err != nil {
		return nil, err
	}
	files := map[string][]byte{}
	total := int64(0)
	for _, name := range strings.Split(tracked, "\x00") {
		if name == "" {
			continue
		}
		if !utf8.ValidString(name) || filepath.IsAbs(name) || strings.Contains(name, "\\") || filepath.Clean(name) != name || strings.HasPrefix(name, "../") {
			return nil, errors.New("unsafe snapshot path")
		}
		if sensitiveFile(filepath.Base(name)) {
			continue
		}
		skip := false
		for _, part := range strings.Split(name, "/") {
			if skippedDirectory(part) {
				skip = true
			}
		}
		if skip {
			continue
		}
		path := filepath.Join(root, name)
		resolved, err := filepath.EvalSymlinks(path)
		if err != nil || resolved != path {
			continue
		}
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		if info.Size() > 1024*1024 || total+info.Size() > 20*1024*1024 || len(files) >= 1000 {
			return nil, errors.New("snapshot exceeds functional validation limit (20 MiB, 1000 files, 1 MiB per file)")
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		files[name] = content
		total += int64(len(content))
	}
	return files, nil
}

func manifestHash(files map[string][]byte) string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		content := sha256.Sum256(files[name])
		hash.Write([]byte(name))
		hash.Write([]byte{0})
		hash.Write(content[:])
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func projectProof(ctx context.Context, snapshot Snapshot, proposal Proposal) (*FunctionalProof, error) {
	socket := os.Getenv("REVIEW_VALIDATOR_SOCKET")
	if socket == "" {
		return nil, errors.New("functional validator is not configured")
	}
	files, err := projectFiles(ctx, snapshot.Dir)
	if err != nil {
		return nil, err
	}
	baselineHash := manifestHash(files)
	payload, err := json.Marshal(map[string]any{"profile": snapshot.ValidationProfile, "directory": snapshot.ValidationDirectory, "files": files, "edits": proposal.Edits})
	if err != nil {
		return nil, err
	}
	if len(payload) > 30*1024*1024 {
		return nil, errors.New("validation payload exceeds maximum size")
	}
	for _, edit := range proposal.Edits {
		content, ok := files[edit.File]
		if !ok || bytes.Count(content, []byte(edit.Old)) != 1 {
			return nil, errors.New("cannot reproduce candidate edit")
		}
		files[edit.File] = bytes.Replace(content, []byte(edit.Old), []byte(edit.New), 1)
	}
	candidateHash := manifestHash(files)
	transport := &http.Transport{DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}}
	defer transport.CloseIdleConnections()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://validator/validate", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	response, err := (&http.Client{Transport: transport}).Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("project validator rejected request (%d)", response.StatusCode)
	}
	var proof FunctionalProof
	if err = json.NewDecoder(io.LimitReader(response.Body, 128*1024)).Decode(&proof); err != nil {
		return nil, err
	}
	if proof.Profile != snapshot.ValidationProfile || proof.SourceSHA256 != candidateHash || proof.BaselineSHA256 != baselineHash || !proof.TestsPreserved || proof.TestsSHA256 == "" || proof.Baseline == nil || (proof.Passed && (proof.ExitCode != 0 || proof.TestCount < 1)) {
		return nil, errors.New("project proof identity or acceptance mismatch")
	}
	return &proof, nil
}
