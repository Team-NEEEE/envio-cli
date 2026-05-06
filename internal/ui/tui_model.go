package ui

import (
	"context"
	"math"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Team-NEEEE/envio-cli/internal/command"
	"github.com/Team-NEEEE/envio-cli/internal/i18n"
)

const (
	defaultTUIWidth    = 80
	minProgressWidth   = 24
	maxProgressWidth   = 64
	progressSideMargin = 14
)

type tuiModel struct {
	command  command.Command
	ctx      context.Context
	steps    []command.Step
	updates  chan command.StepUpdate
	spinner  spinner.Model
	progress progress.Model
	help     help.Model
	keys     tuiKeyMap
	theme    tuiTheme
	result   command.Result
	appErr   *command.AppError
	done     bool
	lang     i18n.Language
	width    int
	height   int
}

func newTUIModel(ctx context.Context, cmd command.Command, lang i18n.Language) tuiModel {
	if ctx == nil {
		ctx = context.Background()
	}
	theme := newTUITheme()
	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(theme.running),
	)
	p := progress.New(
		progress.WithDefaultGradient(),
		progress.WithWidth(48),
	)
	h := help.New()
	h.Width = defaultTUIWidth

	return tuiModel{
		command:  cmd,
		ctx:      ctx,
		steps:    cloneSteps(cmd.Steps()),
		updates:  make(chan command.StepUpdate),
		spinner:  s,
		progress: p,
		help:     h,
		keys:     newTUIKeyMap(lang),
		theme:    theme,
		lang:     lang,
		width:    defaultTUIWidth,
	}
}

func (m tuiModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.runCommand(), m.waitForStep())
}

func (m tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.help.Width = msg.Width
		m.progress.Width = progressWidth(msg.Width)
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.Confirm):
			if m.done {
				return m, tea.Quit
			}
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		}
	case spinner.TickMsg:
		if m.done {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case progress.FrameMsg:
		model, cmd := m.progress.Update(msg)
		if progressModel, ok := model.(progress.Model); ok {
			m.progress = progressModel
		}
		return m, cmd
	case stepMsg:
		m.applyStep(command.StepUpdate(msg))
		progressCmd := m.progress.SetPercent(m.progressPercent())
		return m, tea.Batch(progressCmd, m.waitForStep())
	case doneMsg:
		m.result = msg.result
		m.appErr = msg.appErr
		m.done = true
		m.keys.Confirm.SetEnabled(true)
		if msg.appErr == nil {
			return m, m.progress.SetPercent(1)
		}
		return m, nil
	}
	return m, nil
}

func (m tuiModel) View() string {
	sections := []string{
		m.renderHeader(),
		m.renderProgress(),
		m.renderSteps(),
	}

	if m.done {
		sections = append(sections, m.renderFinalState())
	}
	sections = append(sections, m.renderFooter())

	body := lipgloss.JoinVertical(lipgloss.Left, sections...)
	if m.width > 0 && m.width < defaultTUIWidth {
		return body
	}
	return body + "\n"
}

func (m tuiModel) runCommand() tea.Cmd {
	return func() tea.Msg {
		result, appErr := m.command.Run(m.ctx, channelReporter{updates: m.updates})
		close(m.updates)
		return doneMsg{result: result, appErr: appErr}
	}
}

func (m tuiModel) waitForStep() tea.Cmd {
	return func() tea.Msg {
		update, ok := <-m.updates
		if !ok {
			return nil
		}
		return stepMsg(update)
	}
}

func (m *tuiModel) applyStep(update command.StepUpdate) {
	for i := range m.steps {
		if m.steps[i].ID == update.ID {
			m.steps[i].Status = update.Status
			m.steps[i].Detail = update.Detail
			m.steps[i].DetailKey = update.DetailKey
			return
		}
	}
}

func (m tuiModel) progressPercent() float64 {
	if len(m.steps) == 0 {
		return 0
	}
	var done float64
	for _, step := range m.steps {
		switch step.Status {
		case command.StatusSuccess, command.StatusWarning, command.StatusError:
			done++
		case command.StatusRunning:
			done += 0.5
		}
	}
	return math.Max(0, math.Min(1, done/float64(len(m.steps))))
}

func progressWidth(width int) int {
	if width <= 0 {
		return 48
	}
	size := width - progressSideMargin
	if size < minProgressWidth {
		return minProgressWidth
	}
	if size > maxProgressWidth {
		return maxProgressWidth
	}
	return size
}
