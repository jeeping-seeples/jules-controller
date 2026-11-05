package notify

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/OWNER/jules-tui/internal/julesapi"
)

// UrgencyLevel represents the urgency of a session state.
type UrgencyLevel int

const (
	UrgencyLow UrgencyLevel = iota
	UrgencyMedium
	UrgencyHigh
)

// GetUrgency returns the urgency level of a session state.
func GetUrgency(state julesapi.SessionState) UrgencyLevel {
	switch state {
	case julesapi.StateAwaitingPlanApproval, julesapi.StateAwaitingUserFeedback:
		return UrgencyHigh
	case julesapi.StateInProgress, julesapi.StatePlanning, julesapi.StateQueued:
		return UrgencyMedium
	default:
		return UrgencyLow
	}
}

// NotificationRecord tracks the last notified state for a session.
type NotificationRecord struct {
	SessionID  string                `json:"sessionId"`
	LastState  julesapi.SessionState `json:"lastState"`
	NotifiedAt time.Time             `json:"notifiedAt"`
}

// Store manages persisted notification state across runs.
type Store struct {
	mu       sync.Mutex
	filePath string
	records  map[string]NotificationRecord
}

// DefaultStatePath returns the standard state path for notifications.
func DefaultStatePath() string {
	if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
		return filepath.Join(xdg, "jules-tui", "notifications.json")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".jules-notifications.json")
	}
	return filepath.Join(home, ".local", "state", "jules-tui", "notifications.json")
}

// NewStore initializes a notification record store.
func NewStore(path string) (*Store, error) {
	if path == "" {
		path = DefaultStatePath()
	}
	s := &Store{
		filePath: path,
		records:  make(map[string]NotificationRecord),
	}
	_ = s.load()
	return s, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &s.records)
}

func (s *Store) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.records, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0o600)
}

// ShouldNotify returns true if this state change represents a new transition that requires notification.
// Specifically: transitions into AWAITING_PLAN_APPROVAL, AWAITING_USER_FEEDBACK, COMPLETED, or FAILED.
func (s *Store) ShouldNotify(sessionID string, newState julesapi.SessionState) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	rec, exists := s.records[sessionID]
	if exists && rec.LastState == newState {
		return false
	}

	switch newState {
	case julesapi.StateAwaitingPlanApproval,
		julesapi.StateAwaitingUserFeedback,
		julesapi.StateCompleted,
		julesapi.StateFailed:
		return true
	default:
		return false
	}
}

// RecordTransition updates the store with the newly notified state.
func (s *Store) RecordTransition(sessionID string, state julesapi.SessionState) {
	s.mu.Lock()
	s.records[sessionID] = NotificationRecord{
		SessionID:  sessionID,
		LastState:  state,
		NotifiedAt: time.Now(),
	}
	s.mu.Unlock()
	_ = s.Save()
}

// Notifier dispatches alerts to the user (terminal bell, desktop notifications).
type Notifier struct {
	Store *Store
}

// NewNotifier creates a Notifier.
func NewNotifier(store *Store) *Notifier {
	return &Notifier{Store: store}
}

// Dispatch triggers terminal bell and best-effort OS notification.
func (n *Notifier) Dispatch(title, body string) {
	// Terminal bell
	fmt.Print("\a")

	// Best-effort desktop notification
	switch runtime.GOOS {
	case "darwin":
		script := fmt.Sprintf(`display notification %q with title %q`, body, title)
		_ = exec.Command("osascript", "-e", script).Start()
	case "linux":
		_ = exec.Command("notify-send", title, body).Start()
	}
}
