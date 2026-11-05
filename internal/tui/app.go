package tui

import (
	"github.com/OWNER/jules-tui/internal/julesapi"
	"github.com/OWNER/jules-tui/internal/poller"
	tea "github.com/charmbracelet/bubbletea"
)

type ViewMode int

const (
	ViewModeHome ViewMode = iota
	ViewModeDetail
)

// AppModel wraps both Home and Detail views with smooth navigation.
type AppModel struct {
	client      *julesapi.Client
	homeModel   Model
	detailModel DetailModel
	currentView ViewMode
	width       int
	height      int
	styles      Styles
}

// NewAppModel creates the top-level application model.
func NewAppModel(client *julesapi.Client, p *poller.Poller) AppModel {
	styles := NewStyles()
	return AppModel{
		client:    client,
		homeModel: NewModel(p),
		styles:    styles,
	}
}

func (a AppModel) Init() tea.Cmd {
	return a.homeModel.Init()
}

func (a AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		var cmd1, cmd2 tea.Cmd
		var newHome tea.Model
		newHome, cmd1 = a.homeModel.Update(msg)
		a.homeModel = newHome.(Model)
		if a.currentView == ViewModeDetail {
			a.detailModel, cmd2 = a.detailModel.Update(msg)
		}
		return a, tea.Batch(cmd1, cmd2)

	case tea.KeyMsg:
		if a.currentView == ViewModeDetail {
			switch msg.String() {
			case "esc", "backspace", "h":
				a.currentView = ViewModeHome
				return a, nil
			case "ctrl+c", "q":
				return a, tea.Quit
			default:
				var cmd tea.Cmd
				a.detailModel, cmd = a.detailModel.Update(msg)
				return a, cmd
			}
		}

		// When on Home View:
		switch msg.String() {
		case "enter":
			// If selected item is a session, open Detail view
			if a.homeModel.selectedIndex >= 0 && a.homeModel.selectedIndex < len(a.homeModel.flatItems) {
				item := a.homeModel.flatItems[a.homeModel.selectedIndex]
				if item.Type == ItemTypeSession {
					id := item.Session.ID
					if id == "" {
						id = item.Session.Name
					}
					a.detailModel = NewDetailModel(a.client, id, a.width, a.height, a.styles)
					a.currentView = ViewModeDetail
					return a, a.detailModel.Init()
				}
			}
		}
	}

	if a.currentView == ViewModeDetail {
		var cmd tea.Cmd
		a.detailModel, cmd = a.detailModel.Update(msg)
		return a, cmd
	}

	var cmd tea.Cmd
	var newHome tea.Model
	newHome, cmd = a.homeModel.Update(msg)
	a.homeModel = newHome.(Model)
	return a, cmd
}

func (a AppModel) View() string {
	if a.currentView == ViewModeDetail {
		return a.detailModel.View()
	}
	return a.homeModel.View()
}
