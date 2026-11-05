package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/OWNER/jules-tui/internal/julesapi"
	"github.com/OWNER/jules-tui/internal/poller"
	tea "github.com/charmbracelet/bubbletea"
)

func TestTUIRebuildGroupsAndUrgencySorting(t *testing.T) {
	now := time.Now()
	sessions := []julesapi.Session{
		{
			ID:    "s1",
			Title: "Low priority task",
			State: julesapi.StateCompleted,
			SourceContext: &julesapi.SourceContext{
				Source: "sources/github-org-repoA",
			},
			UpdateTime: &now,
		},
		{
			ID:    "s2",
			Title: "Needs approval!",
			State: julesapi.StateAwaitingPlanApproval,
			SourceContext: &julesapi.SourceContext{
				Source: "sources/github-org-repoA",
			},
			UpdateTime: &now,
		},
		{
			ID:         "s3",
			Title:      "Repoless task",
			State:      julesapi.StateInProgress,
			UpdateTime: &now,
		},
	}

	m := NewModel(nil)
	m.sessions = sessions
	m.rebuildGroups()

	// Should have 2 groups: github-org-repoA and Repoless / Unknown
	if len(m.groups) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(m.groups))
	}

	// Inside github-org-repoA, urgent session s2 must be first, before s1
	var repoAGroup *SessionGroup
	for _, g := range m.groups {
		if g.Name == "github-org-repoA" {
			repoAGroup = g
			break
		}
	}
	if repoAGroup == nil {
		t.Fatalf("github-org-repoA group not found")
	}
	if len(repoAGroup.Sessions) != 2 || repoAGroup.Sessions[0].ID != "s2" {
		t.Errorf("expected urgent session s2 to be sorted first in repoA, got %s", repoAGroup.Sessions[0].ID)
	}

	// Test Needs-Me filter toggle
	m.needsMeOnly = true
	m.rebuildGroups()
	if len(m.groups) != 1 {
		t.Fatalf("expected 1 group in needs-me filter mode, got %d", len(m.groups))
	}
	if len(m.groups[0].Sessions) != 1 || m.groups[0].Sessions[0].ID != "s2" {
		t.Errorf("expected only s2 in needs-me view, got %+v", m.groups[0].Sessions)
	}
}

func TestTUICollapseExpand(t *testing.T) {
	now := time.Now()
	m := NewModel(nil)
	m.sessions = []julesapi.Session{
		{
			ID:    "s1",
			Title: "Task 1",
			State: julesapi.StateCompleted,
			SourceContext: &julesapi.SourceContext{
				Source: "sources/repo1",
			},
			UpdateTime: &now,
		},
	}
	m.rebuildGroups()

	// Initially flatItems has GroupHeader and Session (length 2)
	if len(m.flatItems) != 2 {
		t.Fatalf("expected 2 items, got %d", len(m.flatItems))
	}

	// Press Enter on group header to collapse
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{' '}})
	m = newModel.(Model)

	// Now flatItems should only have GroupHeader (length 1)
	if len(m.flatItems) != 1 {
		t.Fatalf("expected 1 item after collapse, got %d", len(m.flatItems))
	}
	if !m.flatItems[0].Collapsed {
		t.Errorf("expected group header to be marked collapsed")
	}
}

func TestTUIViewOutput(t *testing.T) {
	m := NewModel(nil)
	m.sessions = []julesapi.Session{
		{
			ID:    "sess-99",
			Title: "Test Session",
			State: julesapi.StateAwaitingUserFeedback,
		},
	}
	m.rebuildGroups()

	view := m.View()
	if !strings.Contains(view, "Jules TUI") {
		t.Errorf("expected header in view")
	}
	if !strings.Contains(view, "sess-99") {
		t.Errorf("expected sess-99 in view")
	}
	if !strings.Contains(view, "AWAITING_USER_FEEDBACK") {
		t.Errorf("expected state in view")
	}
}

func TestTUIQuit(t *testing.T) {
	m := NewModel(nil)
	newModel, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	m = newModel.(Model)
	if !m.quitting {
		t.Errorf("expected quitting=true on 'q'")
	}
	if cmd == nil {
		t.Errorf("expected non-nil tea.Quit cmd")
	}
}

func TestTUIPollResultMsg(t *testing.T) {
	m := NewModel(nil)
	res := poller.PollResult{
		Sessions: []julesapi.Session{
			{
				ID:    "polled-1",
				Title: "Polled Task",
				State: julesapi.StateQueued,
			},
		},
		Notifications: []string{"New notification alert"},
		NextInterval:  10 * time.Second,
	}

	newModel, cmd := m.Update(pollResultMsg(res))
	m = newModel.(Model)

	if len(m.sessions) != 1 || m.sessions[0].ID != "polled-1" {
		t.Errorf("expected sessions to be updated from poll result")
	}
	if m.statusMsg != "New notification alert" {
		t.Errorf("expected statusMsg to be updated, got: %s", m.statusMsg)
	}
	if cmd == nil {
		t.Errorf("expected scheduled next poll tick cmd")
	}
}
