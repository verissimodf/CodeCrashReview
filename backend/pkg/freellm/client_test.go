package freellm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUnsupportedOutputControlRetriesOnceWithoutMutatingRequest(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		var request ChatRequest
		json.NewDecoder(r.Body).Decode(&request)
		if request.Model != "coding-model" {
			t.Error("model changed")
		}
		if calls == 1 {
			if request.MaxTokens != 100 {
				t.Error("initial control missing")
			}
			w.WriteHeader(400)
			w.Write([]byte(`{"error":{"message":"OpenAI cannot represent Chat control max_output_tokens"}}`))
			return
		}
		if request.MaxTokens != 0 {
			t.Error("unsupported control retained")
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"choices":[{"message":{"content":"OK"}}]}`))
	}))
	defer server.Close()
	c := &Client{baseURL: server.URL, httpClient: server.Client()}
	req := &ChatRequest{Model: "coding-model", MaxTokens: 100}
	got, err := c.ChatCompletion(context.Background(), req)
	if err != nil || len(got.Choices) != 1 || calls != 2 || req.MaxTokens != 100 {
		t.Fatalf("calls=%d request=%+v err=%v", calls, req, err)
	}
}
