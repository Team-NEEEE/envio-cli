package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestRunHelpUsesEnglishByDefault(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"--help"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "USAGE") {
		t.Fatalf("help should use gh-style usage heading: %s", out.String())
	}
	if !strings.Contains(out.String(), "ADDITIONAL COMMANDS") || !strings.Contains(out.String(), "login:") {
		t.Fatalf("help should show login as a user-facing command: %s", out.String())
	}
	if strings.Contains(out.String(), "completion") {
		t.Fatalf("help should hide shell completion command: %s", out.String())
	}
	if strings.Contains(out.String(), "api-url") {
		t.Fatalf("help should not expose internal API base URL option: %s", out.String())
	}
	if strings.Contains(out.String(), "Envio는 안전한") || strings.Contains(out.String(), "shared CLI foundation") {
		t.Fatalf("help should not render foundation marketing copy: %s", out.String())
	}
}

func TestRunHelpSupportsKoreanWhenRequested(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"--lang", "ko", "--help"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "사용법") {
		t.Fatalf("help should use Korean description when requested: %s", out.String())
	}
}

func TestRunWithoutCommandShowsHelp(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "USAGE") {
		t.Fatalf("root command should show foundation help: %s", out.String())
	}
	if strings.Contains(out.String(), "Envio는 안전한") || strings.Contains(out.String(), "shared CLI foundation") {
		t.Fatalf("root command should not render foundation marketing copy: %s", out.String())
	}
}

func TestRunUnknownCommandPlainShowsUsageAndAvailableCommands(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"unknown"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 2 {
		t.Fatalf("Run() exit = %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), `unknown command "unknown" for "envio"`) {
		t.Fatalf("plain error should show cobra-style command error: %s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "Usage:  envio <command> [flags]") {
		t.Fatalf("plain error should show command usage: %s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "Available commands:\n  login") {
		t.Fatalf("plain error should show available user-facing commands: %s", errOut.String())
	}
	if strings.Contains(errOut.String(), "completion") {
		t.Fatalf("plain error should not show hidden commands: %s", errOut.String())
	}
	if strings.Contains(errOut.String(), "Next step") || strings.Contains(errOut.String(), "UNKNOWN_COMMAND") {
		t.Fatalf("plain error should not use UI hint/debug shape: %s", errOut.String())
	}
}

func TestRunInvalidCompletionShellShowsUsageAndAvailableValues(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"completion", "cmd"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 2 {
		t.Fatalf("Run() exit = %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), `invalid argument "cmd" for "envio completion"`) {
		t.Fatalf("plain error should show invalid argument: %s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "Usage:  envio completion [bash|zsh|fish|powershell] [flags]") {
		t.Fatalf("plain error should show completion usage: %s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "Available values:\n  bash\n  zsh\n  fish\n  powershell") {
		t.Fatalf("plain error should show valid completion shells: %s", errOut.String())
	}
	if strings.Contains(errOut.String(), "Hint:") || strings.Contains(errOut.String(), "UNKNOWN_ARGUMENT") {
		t.Fatalf("plain input error should not use UI hint/debug shape: %s", errOut.String())
	}
}

func TestRunUnknownCommandDebugUsesJSONContract(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"--debug", "unknown"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 2 {
		t.Fatalf("Run() exit = %d, want 2", code)
	}
	var payload struct {
		Status string `json:"status"`
		Error  struct {
			Code     string `json:"code"`
			Message  string `json:"message"`
			Hint     string `json:"hint"`
			Severity string `json:"severity"`
			ExitCode int    `json:"exitCode"`
		} `json:"error"`
	}
	if err := json.Unmarshal(errOut.Bytes(), &payload); err != nil {
		t.Fatalf("json error output invalid: %v, output = %s", err, errOut.String())
	}
	if payload.Status != "error" || payload.Error.Code != "UNKNOWN_COMMAND" {
		t.Fatalf("payload = %#v", payload)
	}
	if payload.Error.ExitCode != 2 || payload.Error.Severity != "error" {
		t.Fatalf("payload has invalid error contract: %#v", payload.Error)
	}
	if !strings.Contains(errOut.String(), "\n  ") {
		t.Fatalf("json output should be pretty printed: %s", errOut.String())
	}
}

func TestRunDebugFromEnvUsesJSONContract(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:    []string{"unknown"},
		Environ: []string{"ENVIO_DEBUG=true"},
		Stdout:  &out,
		Stderr:  &errOut,
		CWD:     t.TempDir(),
	})
	if code != 2 {
		t.Fatalf("Run() exit = %d, want 2", code)
	}
	if !strings.Contains(errOut.String(), `"code": "UNKNOWN_COMMAND"`) {
		t.Fatalf("ENVIO_DEBUG should force JSON debug output: %s", errOut.String())
	}
}

func TestRunCompletionPowershell(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	var errOut bytes.Buffer
	code := Run(context.Background(), Runtime{
		Args:   []string{"completion", "powershell"},
		Stdout: &out,
		Stderr: &errOut,
		CWD:    t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("Run() exit = %d, stderr = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "powershell completion for envio") {
		t.Fatalf("completion output = %s", out.String())
	}
}
