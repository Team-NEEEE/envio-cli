package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Team-NEEEE/envio-cli/internal/command"
	"github.com/Team-NEEEE/envio-cli/internal/i18n"
	"github.com/Team-NEEEE/envio-cli/internal/project"
	"github.com/Team-NEEEE/envio-cli/internal/ui"
)

const pushResultTitle = "Push completed"

type pushService interface {
	Push(context.Context, string, string, command.Reporter) (*project.PushResult, *command.AppError)
}

type pushRunner struct {
	service         pushService
	cwd             string
	environmentFile string
}

func newPushCommand(rt Runtime, lang i18n.Language, global *globalOptions, exitCode *int) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "push [env-file]",
		Short:   pushShort(lang),
		Long:    pushLong(lang),
		Example: pushExample(lang),
		Args:    validateOptionalEnvironmentFileArg(),
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := ui.ModeAuto
			if global.plain {
				mode = ui.ModePlain
			}
			if global.json {
				mode = ui.ModeJSON
			}

			environmentFile := ""
			if len(args) == 1 {
				environmentFile = strings.TrimSpace(args[0])
			}

			renderOptions := renderOptionsWithLanguage(
				rt,
				mode,
				lang,
				global.envDebug || global.debug || global.json,
			)
			runner := pushRunner{
				service:         project.NewSyncService(global.apiURL),
				cwd:             rt.CWD,
				environmentFile: environmentFile,
			}
			code := ui.Execute(cmd.Context(), runner, renderOptions)
			if exitCode != nil {
				*exitCode = code
			}
			return nil
		},
	}
	applyHelpTemplate(cmd, lang)
	return cmd
}

func validateOptionalEnvironmentFileArg() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("accepts at most 1 arg(s), received %d", len(args))
		}
		if len(args) == 1 && strings.TrimSpace(args[0]) == "" {
			return fmt.Errorf("environment file must not be empty for %q", cmd.CommandPath())
		}
		return nil
	}
}

func (r pushRunner) Name() string {
	return "push"
}

func (r pushRunner) Steps() []command.Step {
	return []command.Step{
		{ID: project.StepLoadProjectContext, Label: "Load project context", Status: command.StatusPending},
		{ID: project.StepReadEnvironmentFile, Label: "Read environment file", Status: command.StatusPending},
		{ID: project.StepEncryptEnvironment, Label: "Encrypt environment", Status: command.StatusPending},
		{ID: project.StepPushEnvironment, Label: "Push environment", Status: command.StatusPending},
		{ID: project.StepSaveSyncState, Label: "Save sync state", Status: command.StatusPending},
	}
}

func (r pushRunner) Run(ctx context.Context, reporter command.Reporter) (command.Result, *command.AppError) {
	if r.service == nil {
		return command.Result{}, command.NewAppError(
			project.ErrorPushEnvironmentFailed,
			"push service is not configured",
			"",
			1,
			command.SeverityError,
		)
	}

	result, appErr := r.service.Push(ctx, r.cwd, r.environmentFile, reporter)
	if appErr != nil {
		return command.Result{}, appErr
	}

	return command.Result{
		Title: pushResultTitle,
		Summary: []command.SummaryItem{
			{Label: "Project ID", Value: fmt.Sprintf("%d", result.ProjectID)},
			{Label: "Version ID", Value: fmt.Sprintf("%d", result.VersionID)},
			{Label: "Parent Version ID", Value: fmt.Sprintf("%d", result.ParentVersionID)},
			{Label: "History ID", Value: fmt.Sprintf("%d", result.HistoryID)},
			{Label: "Variable Count", Value: fmt.Sprintf("%d", result.VariableCount)},
			{Label: "Environment File", Value: result.EnvironmentFile},
			{Label: "Local Repository", Value: result.LocalRepository},
		},
	}, nil
}
