package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/OWNER/jules-tui/internal/julesapi"
	"github.com/OWNER/jules-tui/internal/render"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type detailLoadedMsg struct {
	session    *julesapi.Session
	activities []julesapi.Activity
	err        error
}

// DetailModel manages the session detail view.
type DetailModel struct {
	client      *julesapi.Client
	sessionID   string
	session     *julesapi.Session
	activities  []julesapi.Activity
	viewport    viewport.Model
	ready       bool
	showDiffs   bool
	showRawJSON bool
	errorMsg    string
	loading     bool
	width       int
	height      int
	styles      Styles
}

// NewDetailModel initializes a DetailModel.
func NewDetailModel(client *julesapi.Client, sessionID string, width, height int, styles Styles) DetailModel {
	vp := viewport.New(width, max(1, height-6))
	return DetailModel{
		client:    client,
		sessionID: sessionID,
		viewport:  vp,
		width:     width,
		height:    height,
		styles:    styles,
		loading:   true,
	}
}

func (m DetailModel) Init() tea.Cmd {
	return m.loadDataCmd()
}

func (m DetailModel) loadDataCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		sess, err := m.client.GetSession(ctx, m.sessionID)
		if err != nil {
			return detailLoadedMsg{err: err}
		}

		actResp, err := m.client.ListActivities(ctx, m.sessionID, 100, "", "")
		if err != nil {
			// Even if activities fail, show session metadata
			return detailLoadedMsg{session: sess, err: err}
		}

		return detailLoadedMsg{
			session:    sess,
			activities: actResp.Activities,
		}
	}
}

func (m DetailModel) Update(msg tea.Msg) (DetailModel, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := 4
		footerHeight := 2
		verticalMargin := headerHeight + footerHeight

		if !m.ready {
			m.viewport = viewport.New(msg.Width, max(1, msg.Height-verticalMargin))
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = max(1, msg.Height-verticalMargin)
		}
		m.updateContent()

	case detailLoadedMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
		} else {
			m.errorMsg = ""
		}
		if msg.session != nil {
			m.session = msg.session
		}
		m.activities = msg.activities
		m.updateContent()

	case tea.KeyMsg:
		switch msg.String() {
		case "d":
			m.showDiffs = !m.showDiffs
			m.updateContent()
			return m, nil

		case "v", "o":
			m.showRawJSON = !m.showRawJSON
			m.updateContent()
			return m, nil

		case "r":
			m.loading = true
			return m, m.loadDataCmd()

		default:
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *DetailModel) updateContent() {
	var b strings.Builder

	if m.session != nil {
		b.WriteString(fmt.Sprintf("Title:  %s\n", m.session.Title))
		b.WriteString(fmt.Sprintf("State:  %s\n", m.session.State))
		if m.session.URL != "" {
			b.WriteString(fmt.Sprintf("URL:    %s\n", m.session.URL))
		}
		if m.session.Prompt != "" {
			b.WriteString(fmt.Sprintf("Prompt: %s\n", m.session.Prompt))
		}
		if len(m.session.Outputs) > 0 {
			b.WriteString("Outputs:\n")
			for _, out := range m.session.Outputs {
				if out.PullRequest != nil {
					b.WriteString(fmt.Sprintf("  - PR: %s (%s)\n", out.PullRequest.URL, out.PullRequest.Title))
				}
			}
		}
	}
	b.WriteString(strings.Repeat("─", max(20, m.width-2)))
	b.WriteString("\n\n")

	if len(m.activities) == 0 {
		if m.loading {
			b.WriteString("Loading activities...\n")
		} else {
			b.WriteString("No activities recorded for this session yet.\n")
		}
	} else {
		for i, act := range m.activities {
			rendered := render.Activity(act, m.showDiffs, m.showRawJSON, max(40, m.width-6))
			b.WriteString(rendered)
			if i < len(m.activities)-1 {
				b.WriteString("\n\n" + strings.Repeat("┄", max(15, m.width/2)) + "\n\n")
			}
		}
	}

	m.viewport.SetContent(b.String())
}

func (m DetailModel) View() string {
	var b strings.Builder

	// Header
	headerText := fmt.Sprintf(" Session: %s ", m.sessionID)
	b.WriteString(m.styles.Header.Render(headerText))
	if m.loading {
		b.WriteString("  " + m.styles.StatusBanner.Render("[Loading...]"))
	}
	if m.showDiffs {
		b.WriteString("  " + m.styles.FilterActive.Render("DIFFS EXPANDED"))
	}
	if m.showRawJSON {
		b.WriteString("  " + m.styles.FilterActive.Render("RAW JSON"))
	}
	b.WriteString("\n\n")

	if m.errorMsg != "" {
		b.WriteString(m.styles.ErrorBanner.Render("! " + m.errorMsg))
		b.WriteString("\n")
	}

	// Scrollable Viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Footer Help Bar
	help := m.styles.HelpBar.Render("↑/↓/PgUp/PgDn: scroll • d: toggle diffs • v: toggle raw JSON • r: refresh • esc/h: back")
	b.WriteString(help)

	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
