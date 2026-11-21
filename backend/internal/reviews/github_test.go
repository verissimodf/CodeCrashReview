package reviews

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGitHubReferenceConflictNeverWrites(t *testing.T) {
	writes := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			writes++
			t.Error("unexpected reference mutation")
		}
		if r.URL.EscapedPath() != "/repos/owner/repo/git/ref/heads/review/fix" {
			t.Errorf("reference hierarchy lost: %s", r.URL.EscapedPath())
		}
		json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": "existing"}})
	}))
	defer server.Close()
	g := newGitHub("test-token")
	g.base = server.URL
	if err := g.ensureRef(context.Background(), "owner/repo", "heads/review/fix", "different"); err == nil {
		t.Fatal("conflicting reference accepted")
	}
	if writes != 0 {
		t.Fatal("conflicting reference overwritten")
	}
	if err := g.ensureRef(context.Background(), "owner/repo", "heads/review/fix", "existing"); err != nil {
		t.Fatal(err)
	}
}

func TestGitHubCreateReferenceVerified(t *testing.T) {
	created := false
	reads := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer transient" {
			t.Error("missing authentication")
		}
		if r.Method == "POST" {
			var ref map[string]string
			json.NewDecoder(r.Body).Decode(&ref)
			if ref["ref"] != "refs/heads/review/fix" || ref["sha"] != "expected" {
				t.Error("incorrect reference")
			}
			created = true
			w.WriteHeader(201)
			return
		}
		reads++
		if !created {
			w.WriteHeader(404)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"object": map[string]string{"sha": "expected"}})
	}))
	defer server.Close()
	g := newGitHub("transient")
	g.base = server.URL
	if err := g.ensureRef(context.Background(), "owner/repo", "heads/review/fix", "expected"); err != nil {
		t.Fatal(err)
	}
	if reads != 2 || !created {
		t.Fatal("missing read-after-write verification")
	}
}

func TestGitHubNeverFollowsCredentialRedirect(t *testing.T) {
	reached := false
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { reached = true }))
	defer target.Close()
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, target.URL, 302) }))
	defer source.Close()
	g := newGitHub("secret-do-not-echo")
	g.base = source.URL
	err := g.call(context.Background(), "GET", "/", nil, nil)
	if err == nil || reached || strings.Contains(err.Error(), g.token) {
		t.Fatal("credential redirect protection failed")
	}
}
