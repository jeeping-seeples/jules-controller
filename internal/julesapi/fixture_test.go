package julesapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFixturesAndTolerantParsing(t *testing.T) {
	sessionsData, err := os.ReadFile("../../testdata/sessions_list.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	sessionGetData, err := os.ReadFile("../../testdata/session_get.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	sessionCreateData, err := os.ReadFile("../../testdata/session_create.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/v1alpha/sessions":
			_, _ = w.Write(sessionsData)
		case r.Method == http.MethodGet && r.URL.Path == "/v1alpha/sessions/abc123":
			_, _ = w.Write(sessionGetData)
		case r.Method == http.MethodPost && r.URL.Path == "/v1alpha/sessions":
			_, _ = w.Write(sessionCreateData)
		case r.Method == http.MethodPost && r.URL.Path == "/v1alpha/sessions/abc123:sendMessage":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		case r.Method == http.MethodPost && r.URL.Path == "/v1alpha/sessions/abc123:approvePlan":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"message":"not found"}}`))
		}
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")

	// List
	listResp, err := client.ListSessions(context.Background(), 10, "", "")
	if err != nil {
		t.Fatalf("ListSessions error: %v", err)
	}
	if len(listResp.Sessions) != 2 {
		t.Errorf("expected 2 sessions, got %d", len(listResp.Sessions))
	}
	if listResp.NextPageToken != "eyJvZmZzZXQiOjEwfQ==" {
		t.Errorf("unexpected nextPageToken: %s", listResp.NextPageToken)
	}

	// Filter by state
	filteredResp, err := client.ListSessions(context.Background(), 10, "", StateCompleted)
	if err != nil {
		t.Fatalf("ListSessions with filter error: %v", err)
	}
	if len(filteredResp.Sessions) != 1 || filteredResp.Sessions[0].ID != "abc123" {
		t.Errorf("expected 1 filtered session abc123, got %+v", filteredResp.Sessions)
	}

	// Get
	session, err := client.GetSession(context.Background(), "abc123")
	if err != nil {
		t.Fatalf("GetSession error: %v", err)
	}
	if session.ID != "abc123" || len(session.Outputs) == 0 || session.Outputs[0].PullRequest == nil {
		t.Fatalf("unexpected session outputs: %+v", session)
	}
	if session.Outputs[0].PullRequest.URL != "https://github.com/myorg/myrepo/pull/42" {
		t.Errorf("unexpected PR URL: %s", session.Outputs[0].PullRequest.URL)
	}

	// Create
	created, err := client.CreateSession(context.Background(), &CreateSessionRequest{Prompt: "test"})
	if err != nil {
		t.Fatalf("CreateSession error: %v", err)
	}
	if created.ID != "abc123" || created.State != StateQueued {
		t.Errorf("unexpected created session: %+v", created)
	}

	// Send message
	if err := client.SendMessage(context.Background(), "abc123", "hello"); err != nil {
		t.Fatalf("SendMessage error: %v", err)
	}

	// Approve plan
	if err := client.ApprovePlan(context.Background(), "abc123"); err != nil {
		t.Fatalf("ApprovePlan error: %v", err)
	}
}

func TestListActivitiesFixture(t *testing.T) {
	activitiesData, err := os.ReadFile("../../testdata/activities_list.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/v1alpha/sessions/1234567/activities" {
			_, _ = w.Write(activitiesData)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	client := NewClient(ts.URL, "test-key")
	resp, err := client.ListActivities(context.Background(), "1234567", 10, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Activities) != 3 {
		t.Fatalf("expected 3 activities, got %d", len(resp.Activities))
	}
	if resp.Activities[1].PlanGenerated == nil || len(resp.Activities[1].PlanGenerated.Plan.Steps) != 1 {
		t.Errorf("expected plan step in activity 2, got %+v", resp.Activities[1])
	}
	if len(resp.Activities[2].Artifacts) == 0 || resp.Activities[2].Artifacts[0].ChangeSet == nil {
		t.Errorf("expected changeSet artifact in activity 3, got %+v", resp.Activities[2])
	}
}
