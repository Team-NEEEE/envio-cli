package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Team-NEEEE/envio-cli/internal/command"
	"github.com/Team-NEEEE/envio-cli/internal/i18n"
)

func (m tuiModel) renderHeader() string {
	title := m.theme.title.Render("envio " + m.command.Name())
	meta := m.theme.subtle.Render(fmt.Sprintf("%s | interactive", m.lang))
	return lipgloss.JoinHorizontal(lipgloss.Top, title, m.theme.subtle.Render("  "), meta)
}

func (m tuiModel) renderProgress() string {
	progressLine := m.progress.View()
	count := fmt.Sprintf("%d/%d", completedStepCount(m.steps), len(m.steps))
	return m.theme.progress.Render(progressLine + "  " + m.theme.subtle.Render(count))
}

func (m tuiModel) renderSteps() string {
	lines := make([]string, 0, len(m.steps))
	for _, step := range m.steps {
		lines = append(lines, m.renderStep(step))
	}
	return strings.Join(lines, "\n")
}

func (m tuiModel) renderStep(step command.Step) string {
	marker := m.statusMarker(step.Status)
	label := i18n.StepLabel(m.lang, step.ID, step.Label)
	line := fmt.Sprintf("  %-4s %s", marker, label)
	detail := i18n.StepDetail(m.lang, step.DetailKey, step.Detail)
	if detail != "" {
		line += m.theme.detail.Render("  " + detail)
	}
	return line
}

func (m tuiModel) renderFinalState() string {
	if m.appErr != nil {
		return m.theme.section.Render(m.theme.panel.Render(m.renderTUIError(m.appErr)))
	}
	return m.theme.section.Render(m.theme.panel.Render(m.renderTUIResult(m.result)))
}

func (m tuiModel) renderFooter() string {
	return "\n" + m.theme.subtle.Render(m.help.View(m.keys))
}

func (m tuiModel) renderTUIResult(result command.Result) string {
	var b strings.Builder
	title := localizedTitle(result, m.lang)
	if len(result.Warnings) > 0 {
		b.WriteString(m.theme.warning.Render(i18n.StatusLabel(m.lang, "WARN") + ": " + title))
	} else {
		b.WriteString(m.theme.success.Render(i18n.StatusLabel(m.lang, "OK") + ": " + title))
	}
	for _, item := range localizedSummary(result.Summary, m.lang) {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("%s: %s", m.theme.label.Render(item.Label), item.Value))
	}
	for _, warning := range localizedWarnings(result.Warnings, m.lang) {
		hint := warning.Hint
		if hint == "" {
			hint = warning.Message
		}
		b.WriteString("\n")
		b.WriteString(m.theme.warning.Render(i18n.HelpLabel(m.lang)))
		b.WriteString("\n")
		b.WriteString(m.theme.hint.Render(hint))
	}
	return b.String()
}

func (m tuiModel) renderTUIError(appErr *command.AppError) string {
	hint := userFacingHint(appErr, m.lang)
	heading := i18n.HelpLabel(m.lang)
	if appErr.Severity == command.SeverityWarning {
		return m.theme.warning.Render(heading) + "\n" + m.theme.hint.Render(hint)
	}
	return m.theme.error.Render(heading) + "\n" + m.theme.hint.Render(hint)
}

func (m tuiModel) statusMarker(status command.Status) string {
	switch status {
	case command.StatusRunning:
		return m.theme.running.Render(m.spinner.View())
	case command.StatusSuccess:
		return m.theme.success.Render("OK")
	case command.StatusWarning:
		return m.theme.warning.Render("!!")
	case command.StatusError:
		return m.theme.error.Render("XX")
	default:
		return m.theme.pending.Render("--")
	}
}

func completedStepCount(steps []command.Step) int {
	count := 0
	for _, step := range steps {
		switch step.Status {
		case command.StatusSuccess, command.StatusWarning, command.StatusError:
			count++
		}
	}
	return count
}
