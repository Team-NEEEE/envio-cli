package cli

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Team-NEEEE/envio-cli/internal/auth"
	"github.com/Team-NEEEE/envio-cli/internal/command"
	"github.com/Team-NEEEE/envio-cli/internal/i18n"
	"github.com/Team-NEEEE/envio-cli/internal/ui"
)

const loginResultTitle = "Login completed"

type loginOptions struct {
	deviceName string
}

type loginRunner struct {
	service    *auth.LoginService
	deviceName string
}

func newLoginCommand(rt Runtime, lang i18n.Language, global *globalOptions, exitCode *int) *cobra.Command {
	options := loginOptions{}
	cmd := &cobra.Command{
		Use:   "login",
		Short: loginShort(lang),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			mode := ui.ModeAuto
			if global.plain {
				mode = ui.ModePlain
			}
			if global.json {
				mode = ui.ModeJSON
			}

			renderOptions := renderOptionsWithLanguage(
				rt,
				mode,
				lang,
				global.envDebug || global.debug || global.json,
			)
			runner := loginRunner{
				service:    auth.NewLoginService(global.apiURL),
				deviceName: options.deviceName,
			}
			code := ui.Execute(cmd.Context(), runner, renderOptions)
			if exitCode != nil {
				*exitCode = code
			}
			return nil
		},
	}
	applyHelpTemplate(cmd, lang)
	cmd.Flags().StringVar(&options.deviceName, "device-name", "", flagText(lang, "device-name"))
	return cmd
}

func (r loginRunner) Name() string {
	return "login"
}

func (r loginRunner) Steps() []command.Step {
	return []command.Step{
		{ID: "login-start", Label: "Start GitHub login", Status: command.StatusPending},
		{ID: "login-save-session", Label: "Save session and key", Status: command.StatusPending},
	}
}

func (r loginRunner) Run(ctx context.Context, reporter command.Reporter) (command.Result, *command.AppError) {
	reporter.UpdateStep(command.StepUpdate{ID: "login-start", Status: command.StatusRunning})
	response, err := r.service.Login(ctx, r.deviceName)
	if err != nil {
		reporter.UpdateStep(command.StepUpdate{ID: "login-start", Status: command.StatusError})
		if errors.Is(err, auth.ErrAlreadyLoggedIn) {
			return command.Result{}, command.NewAppError(
				"ALREADY_LOGGED_IN",
				"already logged in",
				err.Error(),
				1,
				command.SeverityWarning,
			)
		}
		return command.Result{}, command.NewAppError(
			"LOGIN_FAILED",
			"login failed",
			err.Error(),
			1,
			command.SeverityError,
		)
	}

	reporter.UpdateStep(command.StepUpdate{ID: "login-start", Status: command.StatusSuccess})
	reporter.UpdateStep(command.StepUpdate{ID: "login-save-session", Status: command.StatusSuccess})

	return command.Result{
		Title: loginResultTitle,
		Summary: []command.SummaryItem{
			{Label: "User ID", Value: fmt.Sprintf("%d", response.UserID)},
			{Label: "GitHub ID", Value: response.GithubID},
			{Label: "Email", Value: response.Email},
			{Label: "Device ID", Value: fmt.Sprintf("%d", response.DeviceID)},
		},
	}, nil
}
