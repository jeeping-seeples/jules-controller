package tui

import (
	"os"

	"github.com/charmbracelet/lipgloss"
)

// Styles bundles UI styles for the TUI, respecting NO_COLOR.
type Styles struct {
	Header         lipgloss.Style
	GroupHeader    lipgloss.Style
	SelectedRow    lipgloss.Style
	NormalRow      lipgloss.Style
	UrgentState    lipgloss.Style
	MediumState    lipgloss.Style
	LowState       lipgloss.Style
	StatusBanner   lipgloss.Style
	ErrorBanner    lipgloss.Style
	HelpBar        lipgloss.Style
	FilterActive   lipgloss.Style
	FilterInactive lipgloss.Style
}

// NewStyles returns standard styles or uncolored styles if NO_COLOR is set.
func NewStyles() Styles {
	if os.Getenv("NO_COLOR") != "" {
		return Styles{
			Header:         lipgloss.NewStyle().Bold(true),
			GroupHeader:    lipgloss.NewStyle().Bold(true),
			SelectedRow:    lipgloss.NewStyle().Bold(true),
			NormalRow:      lipgloss.NewStyle(),
			UrgentState:    lipgloss.NewStyle().Bold(true),
			MediumState:    lipgloss.NewStyle(),
			LowState:       lipgloss.NewStyle(),
			StatusBanner:   lipgloss.NewStyle().Bold(true),
			ErrorBanner:    lipgloss.NewStyle().Bold(true),
			HelpBar:        lipgloss.NewStyle(),
			FilterActive:   lipgloss.NewStyle().Bold(true),
			FilterInactive: lipgloss.NewStyle(),
		}
	}

	return Styles{
		Header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#5A56E0")).
			Padding(0, 1),

		GroupHeader: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")),

		SelectedRow: lipgloss.NewStyle().
			Bold(true).
			Background(lipgloss.Color("#2E2E3E")).
			Foreground(lipgloss.Color("#FFFFFF")),

		NormalRow: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D0D0D0")),

		UrgentState: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF5555")),

		MediumState: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E5C07B")),

		LowState: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")),

		StatusBanner: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#98C379")).
			Bold(true),

		ErrorBanner: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E06C75")).
			Bold(true),

		HelpBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),

		FilterActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#E06C75")).
			Padding(0, 1),

		FilterInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7F848E")).
			Padding(0, 1),
	}
}
