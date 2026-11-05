package julesapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestAuthHeader(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if key := r.Header.Get("x-goog-api-key"); key != "test-key" {
			t.Errorf("expected x-goog-api-key 'test-key', got %s", key)
		}
		if r.URL.Query().Get("key") != "" {
			t.Errorf("API key should never be passed via query parameter")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"sessions":[]}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	_, err := client.ListSessions(context.Background(), 10, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRetryOn5xxAnd429(t *testing.T) {
	var attempts int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":{"message":"temporary outage"}}`))
			return
		}
		if n == 2 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":{"message":"rate limited"}}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"s123","state":"COMPLETED"}`))
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key", WithMaxRetries(3), WithBaseDelay(5*time.Millisecond))
	session, err := client.GetSession(context.Background(), "s123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if session.ID != "s123" || session.State != StateCompleted {
		t.Errorf("unexpected session response: %+v", session)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestExitCodesForErrors(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected int
	}{
		{"Unauthorized", 401, ExitAuthError},
		{"Forbidden", 403, ExitAuthError},
		{"NotFound", 404, ExitNotFound},
		{"BadRequest", 400, ExitAPIError},
		{"InternalError", 500, ExitAPIError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = w.Write([]byte(`{"error":{"message":"test failure"}}`))
			}))
			defer ts.Close()

			client := NewClient(ts.URL, "test-key", WithMaxRetries(0))
			_, err := client.GetSession(context.Background(), "any")
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			var apiErr *APIError
			if !strings.Contains(err.Error(), "API error") {
				t.Errorf("expected APIError type string, got %v", err)
			}
			apiErr = NewAPIError(tc.status, "test failure")
			if apiErr.ExitCode != tc.expected {
				t.Errorf("status %d: expected exit code %d, got %d", tc.status, tc.expected, apiErr.ExitCode)
			}
		})
	}
}

func TestResolveSource(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ListSourcesResponse{
			Sources: []Source{
				{
					Name: "sources/github-org-repo1",
					GitHubRepo: &GitHubRepo{
						Owner: "org",
						Repo:  "repo1",
					},
				},
				{
					Name: "sources/github-other-repo2",
					GitHubRepo: &GitHubRepo{
						Owner: "other",
						Repo:  "repo2",
					},
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")

	// 1. Raw sources/...
	src, err := client.ResolveSource(context.Background(), "sources/custom-source", nil)
	if err != nil || src != "sources/custom-source" {
		t.Errorf("expected sources/custom-source, got %s (err: %v)", src, err)
	}

	// 2. Matched owner/repo
	src, err = client.ResolveSource(context.Background(), "org/repo1", nil)
	if err != nil || src != "sources/github-org-repo1" {
		t.Errorf("expected sources/github-org-repo1, got %s (err: %v)", src, err)
	}

	// 3. Not found
	_, err = client.ResolveSource(context.Background(), "org/unconnected", nil)
	if err == nil || !strings.Contains(err.Error(), "not connected") {
		t.Errorf("expected 'not connected' error, got %v", err)
	}
}
