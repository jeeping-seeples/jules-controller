package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OWNER/jules-tui/internal/julesapi"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1alpha/sessions":
			resp := julesapi.ListSessionsResponse{
				Sessions: []julesapi.Session{
					{
						ID:    "sess-1",
						Title: "First Task",
						State: julesapi.StateCompleted,
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodGet && r.URL.Path == "/v1alpha/sessions/sess-1":
			resp := julesapi.Session{
				ID:     "sess-1",
				Title:  "First Task",
				State:  julesapi.StateCompleted,
				Prompt: "Do things",
			}
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodPost && r.URL.Path == "/v1alpha/sessions/sess-1:sendMessage":
			var req julesapi.SendMessageRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			if req.Prompt == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":{"message":"empty prompt"}}`))
				return
			}
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1alpha/sessions/sess-1:approvePlan":
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodGet && r.URL.Path == "/v1alpha/sources":
			resp := julesapi.ListSourcesResponse{
				Sources: []julesapi.Source{
					{
						Name: "sources/github-testorg-myrepo",
						GitHubRepo: &julesapi.GitHubRepo{
							Owner: "testorg",
							Repo:  "myrepo",
						},
					},
				},
			}
			_ = json.NewEncoder(w).Encode(resp)
		case r.Method == http.MethodPost && r.URL.Path == "/v1alpha/sessions":
			resp := julesapi.Session{
				ID:    "sess-new-123",
				State: julesapi.StateQueued,
				URL:   "https://jules.google.com/session/sess-new-123",
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"not found"}}`))
		}
	}))
}

func writeConfigFile(t *testing.T, baseURL, apiKey string) string {
	t.Helper()
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "config.toml")
	content := []byte("base_url = \"" + baseURL + "\"\napi_key = \"" + apiKey + "\"\n")
	if err := os.WriteFile(cfgPath, content, 0o600); err != nil {
		t.Fatalf("failed to write config: %v", err)
	}
	return cfgPath
}

func TestCLISessions(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()
	cfgPath := writeConfigFile(t, ts.URL, "dummy-key")

	var stdout, stderr bytes.Buffer
	code := run([]string{"sessions", "--config", cfgPath}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "sess-1") {
		t.Errorf("expected sess-1 in output, got: %s", stdout.String())
	}

	// JSON mode
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"sessions", "--config", cfgPath, "--json"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d. stderr: %s", code, stderr.String())
	}
	var listResp julesapi.ListSessionsResponse
	if err := json.Unmarshal(stdout.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to parse json output: %v", err)
	}
	if len(listResp.Sessions) != 1 || listResp.Sessions[0].ID != "sess-1" {
		t.Errorf("unexpected json sessions: %+v", listResp)
	}
}

func TestCLIShow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()
	cfgPath := writeConfigFile(t, ts.URL, "dummy-key")

	var stdout, stderr bytes.Buffer
	code := run([]string{"show", "sess-1", "--config", cfgPath}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Session ID: sess-1") {
		t.Errorf("expected session details, got: %s", stdout.String())
	}

	// Missing ID usage error
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"show", "--config", cfgPath}, nil, &stdout, &stderr)
	if code != julesapi.ExitUsageError {
		t.Fatalf("expected code %d, got %d", julesapi.ExitUsageError, code)
	}

	// 404
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"show", "sess-unknown", "--config", cfgPath}, nil, &stdout, &stderr)
	if code != julesapi.ExitNotFound {
		t.Fatalf("expected code %d for 404, got %d", julesapi.ExitNotFound, code)
	}
}

func TestCLISend(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()
	cfgPath := writeConfigFile(t, ts.URL, "dummy-key")

	var stdout, stderr bytes.Buffer
	// Direct message
	code := run([]string{"send", "sess-1", "--message", "hello from test", "--config", cfgPath}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d. stderr: %s", code, stderr.String())
	}

	// Stdin message
	stdout.Reset()
	stderr.Reset()
	stdin := strings.NewReader("message from stdin")
	code = run([]string{"send", "sess-1", "--message-file", "-", "--config", cfgPath}, stdin, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d. stderr: %s", code, stderr.String())
	}

	// Both flags provided -> usage error
	stdout.Reset()
	stderr.Reset()
	code = run([]string{"send", "sess-1", "--message", "foo", "--message-file", "-", "--config", cfgPath}, nil, &stdout, &stderr)
	if code != julesapi.ExitUsageError {
		t.Fatalf("expected code %d, got %d", julesapi.ExitUsageError, code)
	}
}

func TestCLIApprove(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()
	cfgPath := writeConfigFile(t, ts.URL, "dummy-key")

	var stdout, stderr bytes.Buffer
	code := run([]string{"approve", "sess-1", "--config", cfgPath}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "approved successfully") {
		t.Errorf("expected success message, got: %s", stdout.String())
	}
}

func TestCLINew(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()
	cfgPath := writeConfigFile(t, ts.URL, "dummy-key")

	var stdout, stderr bytes.Buffer
	code := run([]string{
		"new",
		"--repo", "testorg/myrepo",
		"--branch", "main",
		"--prompt", "Build something awesome",
		"--require-approval",
		"--config", cfgPath,
	}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("expected code 0, got %d. stderr: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Created session sess-new-123") {
		t.Errorf("expected success output, got: %s", stdout.String())
	}
}
