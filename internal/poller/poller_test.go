package poller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/OWNER/jules-tui/internal/julesapi"
	"github.com/OWNER/jules-tui/internal/notify"
)

func TestPollerSuccessAndNotification(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"sessions": [
				{
					"id": "s-urg",
					"title": "Need Approval",
					"state": "AWAITING_PLAN_APPROVAL"
				}
			]
		}`))
	}))
	defer ts.Close()

	storePath := filepath.Join(t.TempDir(), "store.json")
	store, _ := notify.NewStore(storePath)
	notifier := notify.NewNotifier(store)

	client := julesapi.NewClient(ts.URL, "key")
	p := New(client, notifier, Config{
		Interval:    10 * time.Second,
		MaxInterval: 60 * time.Second,
	})

	// First poll should notify
	res := p.PollOnce(context.Background())
	if res.Error != nil {
		t.Fatalf("unexpected error: %v", res.Error)
	}
	if len(res.Sessions) != 1 {
		t.Errorf("expected 1 session, got %d", len(res.Sessions))
	}
	if len(res.Notifications) != 1 {
		t.Errorf("expected 1 notification on initial transition, got %d", len(res.Notifications))
	}

	// Second poll with same state should NOT notify again
	res2 := p.PollOnce(context.Background())
	if len(res2.Notifications) != 0 {
		t.Errorf("expected 0 notifications on repeated state, got %d", len(res2.Notifications))
	}
}

func TestPollerBackoffOnError(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"message":"server error"}}`))
	}))
	defer ts.Close()

	client := julesapi.NewClient(ts.URL, "key", julesapi.WithMaxRetries(0))
	p := New(client, nil, Config{
		Interval:    2 * time.Second,
		MaxInterval: 10 * time.Second,
	})

	res := p.PollOnce(context.Background())
	if res.Error == nil {
		t.Fatalf("expected error")
	}
	if p.NextInterval() != 4*time.Second {
		t.Errorf("expected interval 4s, got %v", p.NextInterval())
	}

	// Another error backs off further
	p.PollOnce(context.Background())
	if p.NextInterval() != 8*time.Second {
		t.Errorf("expected interval 8s, got %v", p.NextInterval())
	}

	// Another error capped at MaxInterval (10s)
	p.PollOnce(context.Background())
	if p.NextInterval() != 10*time.Second {
		t.Errorf("expected interval capped at 10s, got %v", p.NextInterval())
	}
}
