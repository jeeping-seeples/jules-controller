package poller

import (
	"context"
	"fmt"
	"time"

	"github.com/OWNER/jules-tui/internal/julesapi"
	"github.com/OWNER/jules-tui/internal/notify"
)

// PollResult represents the outcome of a polling cycle.
type PollResult struct {
	Sessions      []julesapi.Session
	Notifications []string
	Error         error
	NextInterval  time.Duration
}

// Config specifies polling parameters.
type Config struct {
	Interval    time.Duration
	MaxInterval time.Duration
}

// DefaultConfig returns default polling configuration.
func DefaultConfig() Config {
	return Config{
		Interval:    30 * time.Second,
		MaxInterval: 5 * time.Minute,
	}
}

// Poller manages adaptive background session fetching.
type Poller struct {
	client       *julesapi.Client
	notifier     *notify.Notifier
	config       Config
	currInterval time.Duration
}

// New creates a new Poller.
func New(client *julesapi.Client, notifier *notify.Notifier, cfg Config) *Poller {
	if cfg.Interval <= 0 {
		cfg.Interval = 30 * time.Second
	}
	if cfg.MaxInterval <= cfg.Interval {
		cfg.MaxInterval = 5 * time.Minute
	}
	return &Poller{
		client:       client,
		notifier:     notifier,
		config:       cfg,
		currInterval: cfg.Interval,
	}
}

// PollOnce executes a single fetch and checks for state transitions.
func (p *Poller) PollOnce(ctx context.Context) PollResult {
	resp, err := p.client.ListSessions(ctx, 100, "", "")
	if err != nil {
		// Backoff on error
		p.currInterval = p.currInterval * 2
		if p.currInterval > p.config.MaxInterval {
			p.currInterval = p.config.MaxInterval
		}
		return PollResult{
			Error:        err,
			NextInterval: p.currInterval,
		}
	}

	// Reset interval to base on success
	p.currInterval = p.config.Interval

	var newNotifications []string
	if p.notifier != nil && p.notifier.Store != nil {
		for _, s := range resp.Sessions {
			id := s.ID
			if id == "" {
				id = s.Name
			}
			if p.notifier.Store.ShouldNotify(id, s.State) {
				msg := formatNotificationMessage(s)
				newNotifications = append(newNotifications, msg)
				p.notifier.Dispatch("Google Jules Update", msg)
				p.notifier.Store.RecordTransition(id, s.State)
			}
		}
	}

	return PollResult{
		Sessions:      resp.Sessions,
		Notifications: newNotifications,
		NextInterval:  p.currInterval,
	}
}

// NextInterval returns current scheduled interval.
func (p *Poller) NextInterval() time.Duration {
	return p.currInterval
}

func formatNotificationMessage(s julesapi.Session) string {
	id := s.ID
	if id == "" {
		id = s.Name
	}
	title := s.Title
	if title == "" {
		title = s.Prompt
	}
	if len(title) > 30 {
		title = title[:27] + "..."
	}

	switch s.State {
	case julesapi.StateAwaitingPlanApproval:
		return fmt.Sprintf("Session %s needs plan approval: %s", id, title)
	case julesapi.StateAwaitingUserFeedback:
		return fmt.Sprintf("Session %s needs user feedback: %s", id, title)
	case julesapi.StateCompleted:
		return fmt.Sprintf("Session %s completed: %s", id, title)
	case julesapi.StateFailed:
		return fmt.Sprintf("Session %s failed: %s", id, title)
	default:
		return fmt.Sprintf("Session %s transitioned to %s", id, s.State)
	}
}
