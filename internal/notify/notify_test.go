package notify

import (
	"path/filepath"
	"testing"

	"github.com/OWNER/jules-tui/internal/julesapi"
)

func TestUrgencyLevels(t *testing.T) {
	if GetUrgency(julesapi.StateAwaitingPlanApproval) != UrgencyHigh {
		t.Errorf("expected UrgencyHigh for AWAITING_PLAN_APPROVAL")
	}
	if GetUrgency(julesapi.StateAwaitingUserFeedback) != UrgencyHigh {
		t.Errorf("expected UrgencyHigh for AWAITING_USER_FEEDBACK")
	}
	if GetUrgency(julesapi.StateInProgress) != UrgencyMedium {
		t.Errorf("expected UrgencyMedium for IN_PROGRESS")
	}
	if GetUrgency(julesapi.StateCompleted) != UrgencyLow {
		t.Errorf("expected UrgencyLow for COMPLETED")
	}
}

func TestStorePersistenceAndShouldNotify(t *testing.T) {
	tmpDir := t.TempDir()
	storePath := filepath.Join(tmpDir, "notifications.json")

	store, err := NewStore(storePath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	sessionID := "sess-100"

	// 1. First time entering AWAITING_PLAN_APPROVAL -> should notify
	if !store.ShouldNotify(sessionID, julesapi.StateAwaitingPlanApproval) {
		t.Errorf("expected ShouldNotify to be true for new approval state")
	}

	// Record transition
	store.RecordTransition(sessionID, julesapi.StateAwaitingPlanApproval)

	// 2. Same state again -> should NOT notify
	if store.ShouldNotify(sessionID, julesapi.StateAwaitingPlanApproval) {
		t.Errorf("expected ShouldNotify to be false for already notified state")
	}

	// 3. Low urgency transition (e.g. IN_PROGRESS) -> should NOT notify
	if store.ShouldNotify(sessionID, julesapi.StateInProgress) {
		t.Errorf("expected ShouldNotify to be false for IN_PROGRESS")
	}

	// 4. Reload store from disk
	store2, err := NewStore(storePath)
	if err != nil {
		t.Fatalf("failed to reload store: %v", err)
	}
	if store2.ShouldNotify(sessionID, julesapi.StateAwaitingPlanApproval) {
		t.Errorf("persisted store should remember already notified state")
	}

	// 5. Completion transition -> should notify
	if !store2.ShouldNotify(sessionID, julesapi.StateCompleted) {
		t.Errorf("expected ShouldNotify to be true for completion")
	}
}
