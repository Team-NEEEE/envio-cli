package ui

import "github.com/charmbracelet/lipgloss"

type tuiTheme struct {
	title    lipgloss.Style
	subtle   lipgloss.Style
	panel    lipgloss.Style
	section  lipgloss.Style
	success  lipgloss.Style
	warning  lipgloss.Style
	error    lipgloss.Style
	running  lipgloss.Style
	pending  lipgloss.Style
	detail   lipgloss.Style
	hint     lipgloss.Style
	label    lipgloss.Style
	progress lipgloss.Style
}

func newTUITheme() tuiTheme {
	return tuiTheme{
		title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
		subtle:   lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		panel:    lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8")).Padding(1, 2),
		section:  lipgloss.NewStyle().MarginTop(1),
		success:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")),
		warning:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11")),
		error:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")),
		running:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
		pending:  lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		detail:   lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
		hint:     lipgloss.NewStyle().PaddingLeft(2),
		label:    lipgloss.NewStyle().Bold(true),
		progress: lipgloss.NewStyle().MarginBottom(1),
	}
}
