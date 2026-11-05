package tui

import (
	"strings"
	"testing"

	"github.com/OWNER/jules-tui/internal/julesapi"
	tea "github.com/charmbracelet/bubbletea"
)

func TestAppNavigationHomeToDetailAndBack(t *testing.T) {
	app := NewAppModel(nil, nil)
	app.homeModel.sessions = []julesapi.Session{
		{
			ID:    "s-detail-1",
			Title: "Test Detail Task",
			State: julesapi.StateInProgress,
		},
	}
	app.homeModel.rebuildGroups()
	app.width = 80
	app.height = 24

	// Select the session row (index 1, as index 0 is group header)
	app.homeModel.selectedIndex = 1

	// Press Enter to enter Detail view
	newModel, _ := app.Update(tea.KeyMsg{Type: tea.KeyEnter})
	app = newModel.(AppModel)

	if app.currentView != ViewModeDetail {
		t.Fatalf("expected currentView to be ViewModeDetail, got %d", app.currentView)
	}
	if app.detailModel.sessionID != "s-detail-1" {
		t.Errorf("expected detailModel sessionID to be s-detail-1, got %s", app.detailModel.sessionID)
	}

	// Press Esc to return to Home view
	newModel, _ = app.Update(tea.KeyMsg{Type: tea.KeyEsc})
	app = newModel.(AppModel)

	if app.currentView != ViewModeHome {
		t.Fatalf("expected currentView to be ViewModeHome, got %d", app.currentView)
	}
}

func TestDetailToggleDiffsAndRawJSON(t *testing.T) {
	styles := NewStyles()
	detail := NewDetailModel(nil, "s1", 80, 24, styles)
	detail.ready = true
	detail.activities = []julesapi.Activity{
		{
			ID:      "act-diff",
			RawJSON: `{"hello":"world"}`,
			Artifacts: []julesapi.Artifact{
				{
					ChangeSet: &julesapi.ChangeSet{
						GitPatch: &julesapi.GitPatch{
							UnidiffPatch: "+added line\n",
						},
					},
				},
			},
		},
	}
	detail.updateContent()

	// Toggle diffs ('d')
	detail, _ = detail.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'d'}})
	if !detail.showDiffs {
		t.Errorf("expected showDiffs to be true")
	}

	// Toggle raw json ('v')
	detail, _ = detail.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'v'}})
	if !detail.showRawJSON {
		t.Errorf("expected showRawJSON to be true")
	}

	view := detail.View()
	if !strings.Contains(view, "RAW JSON") {
		t.Errorf("expected RAW JSON badge in view, got: %s", view)
	}
}
