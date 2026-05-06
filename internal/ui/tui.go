package ui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Team-NEEEE/envio-cli/internal/command"
)

type stepMsg command.StepUpdate

type doneMsg struct {
	appErr *command.AppError
	result command.Result
}

func runTUI(ctx context.Context, cmd command.Command, options Options) int {
	if ctx == nil {
		ctx = context.Background()
	}
	model := newTUIModel(ctx, cmd, options.language())
	programOptions := []tea.ProgramOption{tea.WithContext(ctx)}
	if options.Input != nil {
		programOptions = append(programOptions, tea.WithInput(options.Input))
	}
	if options.Output != nil {
		programOptions = append(programOptions, tea.WithOutput(options.Output))
	}

	finalModel, err := tea.NewProgram(model, programOptions...).Run()
	if err != nil {
		return renderPlainError(command.NewAppError(
			"TUI_RENDER_FAILED",
			"terminal UI failed to run",
			err.Error(),
			1,
			command.SeverityError,
		), options)
	}

	completed, ok := finalModel.(tuiModel)
	if !ok {
		return 1
	}
	if completed.appErr != nil {
		return completed.appErr.ExitCode
	}
	return 0
}

type channelReporter struct {
	updates chan<- command.StepUpdate
}

func (r channelReporter) UpdateStep(update command.StepUpdate) {
	r.updates <- update
}

func cloneSteps(steps []command.Step) []command.Step {
	cloned := make([]command.Step, len(steps))
	copy(cloned, steps)
	return cloned
}
