package reviews

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type githubClient struct {
	base, token string
	client      *http.Client
}
type githubError struct{ status int }

func (e githubError) Error() string {
	return fmt.Sprintf("GitHub returned HTTP %d; check repository permissions and reference conflicts", e.status)
}

func newGitHub(token string) *githubClient {
	return &githubClient{base: "https://api.github.com", token: token, client: &http.Client{Timeout: 30 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (g *githubClient) call(ctx context.Context, method, path string, input, output any) error {
	var body io.Reader
	if input != nil {
		encoded, err := json.Marshal(input)
		if err != nil {
			return err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, g.base+path, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+g.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2026-03-10")
	req.Header.Set("Content-Type", "application/json")
	response, err := g.client.Do(req)
	if err != nil {
		return fmt.Errorf("GitHub request could not complete")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return githubError{response.StatusCode}
	}
	if output == nil {
		return nil
	}
	return json.NewDecoder(io.LimitReader(response.Body, 2*1024*1024)).Decode(output)
}

func (g *githubClient) ref(ctx context.Context, repo, ref string) (string, error) {
	var result struct {
		Object struct {
			SHA string `json:"sha"`
		} `json:"object"`
	}
	err := g.call(ctx, "GET", "/repos/"+repo+"/git/ref/"+escapeRef(ref), nil, &result)
	return result.Object.SHA, err
}

func (g *githubClient) ensureRef(ctx context.Context, repo, ref, sha string) error {
	existing, err := g.ref(ctx, repo, ref)
	if err == nil {
		if existing != sha {
			return fmt.Errorf("remote reference already exists with a different commit; it will not be overwritten")
		}
		return nil
	}
	if status, ok := err.(githubError); !ok || status.status != 404 {
		return err
	}
	if err = g.call(ctx, "POST", "/repos/"+repo+"/git/refs", map[string]string{"ref": "refs/" + ref, "sha": sha}, nil); err != nil {
		return err
	}
	confirmed, err := g.ref(ctx, repo, ref)
	if err != nil {
		return err
	}
	if confirmed != sha {
		return fmt.Errorf("remote reference verification failed")
	}
	return nil
}

type publishedFile struct {
	Content []byte
	Mode    string
}

func (g *githubClient) commit(ctx context.Context, repo, parent, message string, files map[string]publishedFile) (string, error) {
	var original struct {
		Tree struct {
			SHA string `json:"sha"`
		} `json:"tree"`
	}
	if err := g.call(ctx, "GET", "/repos/"+repo+"/git/commits/"+parent, nil, &original); err != nil {
		return "", err
	}
	var tree []map[string]string
	for name, file := range files {
		var blob struct {
			SHA string `json:"sha"`
		}
		if err := g.call(ctx, "POST", "/repos/"+repo+"/git/blobs", map[string]string{"content": base64.StdEncoding.EncodeToString(file.Content), "encoding": "base64"}, &blob); err != nil {
			return "", err
		}
		tree = append(tree, map[string]string{"path": name, "mode": file.Mode, "type": "blob", "sha": blob.SHA})
	}
	var created struct {
		SHA string `json:"sha"`
	}
	if err := g.call(ctx, "POST", "/repos/"+repo+"/git/trees", map[string]any{"base_tree": original.Tree.SHA, "tree": tree}, &created); err != nil {
		return "", err
	}
	var commit struct {
		SHA string `json:"sha"`
	}
	if err := g.call(ctx, "POST", "/repos/"+repo+"/git/commits", map[string]any{"message": message, "tree": created.SHA, "parents": []string{parent}}, &commit); err != nil {
		return "", err
	}
	if len(commit.SHA) != 40 {
		return "", fmt.Errorf("invalid remote commit identity")
	}
	return commit.SHA, nil
}

func (g *githubClient) release(ctx context.Context, repo, tag, sha, name, body string) (string, error) {
	var release struct {
		HTMLURL string `json:"html_url"`
	}
	err := g.call(ctx, "GET", "/repos/"+repo+"/releases/tags/"+url.PathEscape(tag), nil, &release)
	if err == nil {
		return release.HTMLURL, nil
	}
	if status, ok := err.(githubError); !ok || status.status != 404 {
		return "", err
	}
	err = g.call(ctx, "POST", "/repos/"+repo+"/releases", map[string]any{"tag_name": tag, "target_commitish": sha, "name": name, "body": body, "draft": false, "prerelease": false, "make_latest": "false"}, &release)
	return release.HTMLURL, err
}

func githubRepository(raw string) (string, error) {
	if strings.HasPrefix(raw, "git@github.com:") {
		raw = "https://github.com/" + strings.TrimPrefix(raw, "git@github.com:")
	}
	canonical, err := validateGitHubURL(raw)
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(strings.TrimPrefix(canonical, "https://github.com/"), ".git"), nil
}

// Keep the reference hierarchy while escaping each path component.
func escapeRef(ref string) string {
	parts := strings.Split(ref, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}
