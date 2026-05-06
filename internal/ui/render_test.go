package ui

import (
	"bytes"
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Team-NEEEE/envio-cli/internal/command"
)

func TestRenderersHideSensitiveSummary(t *testing.T) {
	t.Parallel()

	result := command.Result{
		Title: "Command completed",
		Summary: []command.SummaryItem{
			{Label: "Name", Value: "envio"},
			{Label: "Secret value", Value: "secret-value", Sensitive: true},
		},
	}

	var plain bytes.Buffer
	renderPlainResult(result, Options{Output: &plain})
	if strings.Contains(plain.String(), "secret-value") {
		t.Fatalf("plain output leaked sensitive summary: %s", plain.String())
	}

	var jsonOut bytes.Buffer
	renderJSONResult(result, Options{Output: &jsonOut})
	if strings.Contains(jsonOut.String(), "secret-value") {
		t.Fatalf("json output leaked sensitive summary: %s", jsonOut.String())
	}
}

func TestRenderErrorUsesCommonMessageShape(t *testing.T) {
	t.Parallel()

	appErr := command.NewAppError("CONFIG_REQUIRED", "required configuration is missing", "Check configuration.", 1, command.SeverityError)

	var plainErr bytes.Buffer
	code := renderPlainError(appErr, Options{ErrOutput: &plainErr})
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(plainErr.String(), "Hint: Check the required configuration and run the command again.") {
		t.Fatalf("plain error did not include user-facing hint: %s", plainErr.String())
	}
	if strings.Contains(plainErr.String(), "CONFIG_REQUIRED") || strings.Contains(plainErr.String(), "Required configuration is missing.") {
		t.Fatalf("plain error should hide debug details: %s", plainErr.String())
	}

	var jsonErr bytes.Buffer
	code = renderJSONError(appErr, Options{ErrOutput: &jsonErr})
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(jsonErr.String(), `"code": "CONFIG_REQUIRED"`) {
		t.Fatalf("json error did not include standard shape: %s", jsonErr.String())
	}
	if !strings.Contains(jsonErr.String(), "Required configuration is missing.") {
		t.Fatalf("json error should be localized: %s", jsonErr.String())
	}
	if !strings.Contains(jsonErr.String(), "\n  ") {
		t.Fatalf("json error should be pretty printed: %s", jsonErr.String())
	}
}

func TestTUIErrorShowsOnlyHint(t *testing.T) {
	t.Parallel()

	appErr := command.NewAppError("CONFIG_REQUIRED", "required configuration is missing", "Check configuration.", 1, command.SeverityError)
	model := newTUIModel(context.Background(), stubCommand{}, "ko")
	rendered := model.renderTUIError(appErr)
	if !strings.Contains(rendered, "필수 설정을 확인한 뒤 명령을 다시 실행해 주세요.") {
		t.Fatalf("TUI error did not include user-facing hint: %s", rendered)
	}
	if strings.Contains(rendered, "CONFIG_REQUIRED") || strings.Contains(rendered, "필수 설정이 누락되었습니다") {
		t.Fatalf("TUI error should hide debug details: %s", rendered)
	}
}

func TestTUITracksWindowSizeAndHelpState(t *testing.T) {
	t.Parallel()

	model := newTUIModel(context.Background(), stubCommand{}, "ko")
	updated, _ := model.Update(tea.WindowSizeMsg{Width: 42, Height: 20})
	model, ok := updated.(tuiModel)
	if !ok {
		t.Fatalf("updated model = %T, want tuiModel", updated)
	}
	if model.width != 42 || model.help.Width != 42 {
		t.Fatalf("window size was not applied: width=%d help=%d", model.width, model.help.Width)
	}
	if model.progress.Width != progressWidth(42) {
		t.Fatalf("progress width = %d, want %d", model.progress.Width, progressWidth(42))
	}

	updated, _ = model.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	model, ok = updated.(tuiModel)
	if !ok {
		t.Fatalf("updated model = %T, want tuiModel", updated)
	}
	if !model.help.ShowAll {
		t.Fatalf("help key should toggle full help")
	}
}

func TestTUIDoneEnablesConfirmKey(t *testing.T) {
	t.Parallel()

	model := newTUIModel(context.Background(), stubCommand{}, "ko")
	if model.keys.Confirm.Enabled() {
		t.Fatalf("confirm key should be disabled while command is running")
	}
	updated, _ := model.Update(doneMsg{result: command.Result{Title: "Command completed"}})
	model, ok := updated.(tuiModel)
	if !ok {
		t.Fatalf("updated model = %T, want tuiModel", updated)
	}
	if !model.keys.Confirm.Enabled() {
		t.Fatalf("confirm key should be enabled after command completion")
	}
	if !strings.Contains(model.renderFooter(), "enter") {
		t.Fatalf("footer should show confirm key after completion: %s", model.renderFooter())
	}
}

type stubCommand struct{}

func (stubCommand) Name() string {
	return "foundation"
}

func (stubCommand) Steps() []command.Step {
	return []command.Step{
		{ID: "prepare", Label: "Prepare", Status: command.StatusPending},
		{ID: "execute", Label: "Execute", Status: command.StatusPending},
	}
}

func (stubCommand) Run(context.Context, command.Reporter) (command.Result, *command.AppError) {
	return command.Result{Title: "Command completed"}, nil
}
