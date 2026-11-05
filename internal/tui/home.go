package tui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/OWNER/jules-tui/internal/julesapi"
	"github.com/OWNER/jules-tui/internal/notify"
	"github.com/OWNER/jules-tui/internal/poller"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Messages
type pollResultMsg poller.PollResult
type tickMsg time.Time

// GroupItem represents an item in the navigation list (either group header or session row).
type ListItemType int

const (
	ItemTypeGroupHeader ListItemType = iota
	ItemTypeSession
)

type ListItem struct {
	Type      ListItemType
	GroupName string
	Session   julesapi.Session
	Collapsed bool
}

// SessionGroup holds grouped sessions.
type SessionGroup struct {
	Name      string
	Sessions  []julesapi.Session
	Collapsed bool
}

// Model represents the Home screen state.
type Model struct {
	poller        *poller.Poller
	sessions      []julesapi.Session
	groups        []*SessionGroup
	flatItems     []ListItem
	selectedIndex int
	needsMeOnly   bool
	collapsedMap  map[string]bool
	statusMsg     string
	errorMsg      string
	lastPollTime  time.Time
	width         int
	height        int
	styles        Styles
	quitting      bool
}

// NewModel creates a new Home screen model.
func NewModel(p *poller.Poller) Model {
	return Model{
		poller:       p,
		collapsedMap: make(map[string]bool),
		styles:       NewStyles(),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.doPollCmd(),
	)
}

func (m Model) doPollCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		res := m.poller.PollOnce(ctx)
		return pollResultMsg(res)
	}
}

func (m Model) scheduleNextPollCmd(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tickMsg:
		return m, m.doPollCmd()

	case pollResultMsg:
		res := poller.PollResult(msg)
		if res.Error != nil {
			m.errorMsg = fmt.Sprintf("Refresh failed: %v", res.Error)
		} else {
			m.errorMsg = ""
			m.sessions = res.Sessions
			m.lastPollTime = time.Now()
			m.rebuildGroups()
			if len(res.Notifications) > 0 {
				m.statusMsg = res.Notifications[len(res.Notifications)-1]
			}
		}
		return m, m.scheduleNextPollCmd(res.NextInterval)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit

		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
			}
			return m, nil

		case "down", "j":
			if m.selectedIndex < len(m.flatItems)-1 {
				m.selectedIndex++
			}
			return m, nil

		case "tab", "n":
			m.needsMeOnly = !m.needsMeOnly
			m.rebuildGroups()
			if m.selectedIndex >= len(m.flatItems) && len(m.flatItems) > 0 {
				m.selectedIndex = len(m.flatItems) - 1
			}
			return m, nil

		case "enter", " ", "space":
			if m.selectedIndex >= 0 && m.selectedIndex < len(m.flatItems) {
				item := m.flatItems[m.selectedIndex]
				if item.Type == ItemTypeGroupHeader {
					m.collapsedMap[item.GroupName] = !m.collapsedMap[item.GroupName]
					m.rebuildGroups()
				}
			}
			return m, nil

		case "r":
			m.statusMsg = "Refreshing..."
			return m, m.doPollCmd()
		}
	}

	return m, nil
}

func (m *Model) rebuildGroups() {
	// Filter sessions if needsMeOnly is set
	var filtered []julesapi.Session
	for _, s := range m.sessions {
		if m.needsMeOnly {
			if notify.GetUrgency(s.State) == notify.UrgencyHigh {
				filtered = append(filtered, s)
			}
		} else {
			filtered = append(filtered, s)
		}
	}

	// Group by repo
	grouped := make(map[string][]julesapi.Session)
	for _, s := range filtered {
		repoKey := "Repoless / Unknown"
		if s.SourceContext != nil && s.SourceContext.Source != "" {
			src := strings.TrimPrefix(s.SourceContext.Source, "sources/")
			repoKey = src
		}
		grouped[repoKey] = append(grouped[repoKey], s)
	}

	// Sort sessions in each group by urgency descending, then by updateTime descending
	for k := range grouped {
		sort.Slice(grouped[k], func(i, j int) bool {
			urgI := notify.GetUrgency(grouped[k][i].State)
			urgJ := notify.GetUrgency(grouped[k][j].State)
			if urgI != urgJ {
				return urgI > urgJ
			}
			var timeI, timeJ time.Time
			if grouped[k][i].UpdateTime != nil {
				timeI = *grouped[k][i].UpdateTime
			}
			if grouped[k][j].UpdateTime != nil {
				timeJ = *grouped[k][j].UpdateTime
			}
			return timeI.After(timeJ)
		})
	}

	// Sort group names
	var groupNames []string
	for k := range grouped {
		groupNames = append(groupNames, k)
	}
	sort.Strings(groupNames)

	m.groups = make([]*SessionGroup, 0, len(groupNames))
	m.flatItems = make([]ListItem, 0)

	for _, gName := range groupNames {
		collapsed := m.collapsedMap[gName]
		grp := &SessionGroup{
			Name:      gName,
			Sessions:  grouped[gName],
			Collapsed: collapsed,
		}
		m.groups = append(m.groups, grp)

		// Header item
		m.flatItems = append(m.flatItems, ListItem{
			Type:      ItemTypeGroupHeader,
			GroupName: gName,
			Collapsed: collapsed,
		})

		// Child items if not collapsed
		if !collapsed {
			for _, s := range grp.Sessions {
				m.flatItems = append(m.flatItems, ListItem{
					Type:      ItemTypeSession,
					GroupName: gName,
					Session:   s,
				})
			}
		}
	}

	if m.selectedIndex >= len(m.flatItems) {
		m.selectedIndex = len(m.flatItems) - 1
	}
	if m.selectedIndex < 0 {
		m.selectedIndex = 0
	}
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title Bar
	title := m.styles.Header.Render(" Jules TUI ")
	var filterBadge string
	if m.needsMeOnly {
		filterBadge = m.styles.FilterActive.Render("FILTER: NEEDS HUMAN ACTION")
	} else {
		filterBadge = m.styles.FilterInactive.Render("FILTER: ALL SESSIONS")
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, title, " ", filterBadge))
	b.WriteString("\n\n")

	// Status / Error Banners
	if m.errorMsg != "" {
		b.WriteString(m.styles.ErrorBanner.Render("! " + m.errorMsg))
		b.WriteString("\n")
	} else if m.statusMsg != "" {
		b.WriteString(m.styles.StatusBanner.Render("* " + m.statusMsg))
		b.WriteString("\n")
	}

	// Content List
	if len(m.flatItems) == 0 {
		if m.needsMeOnly {
			b.WriteString(m.styles.NormalRow.Render("No sessions currently require human action."))
		} else {
			b.WriteString(m.styles.NormalRow.Render("No sessions found."))
		}
		b.WriteString("\n")
	} else {
		for i, item := range m.flatItems {
			isSelected := i == m.selectedIndex
			prefix := "  "
			if isSelected {
				prefix = "> "
			}

			if item.Type == ItemTypeGroupHeader {
				caret := "▾"
				if item.Collapsed {
					caret = "▸"
				}
				line := fmt.Sprintf("%s%s %s", prefix, caret, item.GroupName)
				if isSelected {
					b.WriteString(m.styles.SelectedRow.Render(line))
				} else {
					b.WriteString(m.styles.GroupHeader.Render(line))
				}
			} else {
				// Session line
				id := item.Session.ID
				if id == "" {
					id = strings.TrimPrefix(item.Session.Name, "sessions/")
				}
				title := item.Session.Title
				if title == "" {
					title = item.Session.Prompt
				}
				if len(title) > 35 {
					title = title[:32] + "..."
				}

				state := string(item.Session.State)
				var stateFormatted string
				urg := notify.GetUrgency(item.Session.State)
				switch urg {
				case notify.UrgencyHigh:
					stateFormatted = m.styles.UrgentState.Render(fmt.Sprintf("[! %s]", state))
				case notify.UrgencyMedium:
					stateFormatted = m.styles.MediumState.Render(fmt.Sprintf("[%s]", state))
				default:
					stateFormatted = m.styles.LowState.Render(fmt.Sprintf("[%s]", state))
				}

				rowText := fmt.Sprintf("%s    %-12s  %-36s  %s", prefix, id, title, stateFormatted)
				if isSelected {
					b.WriteString(m.styles.SelectedRow.Render(rowText))
				} else {
					b.WriteString(m.styles.NormalRow.Render(rowText))
				}
			}
			b.WriteString("\n")
		}
	}

	// Help Bar at bottom
	b.WriteString("\n")
	help := m.styles.HelpBar.Render("↑/↓: navigate • space/enter: toggle group • tab/n: needs-me filter • r: refresh • q: quit")
	b.WriteString(help)
	b.WriteString("\n")

	return b.String()
}
