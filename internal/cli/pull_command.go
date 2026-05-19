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

const pullResultTitle = "Pull completed"

type pullService interface {
	Pull(context.Context, string, string, command.Reporter) (*project.PullResult, *command.AppError)
}

type pullRunner struct {
	service         pullService
	cwd             string
	environmentFile string
}

func newPullCommand(rt Runtime, lang i18n.Language, global *globalOptions, exitCode *int) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pull [env-file]",
		Short:   pullShort(lang),
		Long:    pullLong(lang),
		Example: pullExample(lang),
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
			runner := pullRunner{
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

func (r pullRunner) Name() string {
	return "pull"
}

func (r pullRunner) Steps() []command.Step {
	return []command.Step{
		{ID: project.StepLoadProjectContext, Label: "Load project context", Status: command.StatusPending},
		{ID: project.StepPullEnvironment, Label: "Pull environment", Status: command.StatusPending},
		{ID: project.StepDecryptEnvironment, Label: "Decrypt environment", Status: command.StatusPending},
		{ID: project.StepWriteEnvironment, Label: "Write environment file", Status: command.StatusPending},
		{ID: project.StepSaveSyncState, Label: "Save sync state", Status: command.StatusPending},
	}
}

func (r pullRunner) Run(ctx context.Context, reporter command.Reporter) (command.Result, *command.AppError) {
	if r.service == nil {
		return command.Result{}, command.NewAppError(
			project.ErrorPullEnvironmentFailed,
			"pull service is not configured",
			"",
			1,
			command.SeverityError,
		)
	}

	result, appErr := r.service.Pull(ctx, r.cwd, r.environmentFile, reporter)
	if appErr != nil {
		return command.Result{}, appErr
	}

	return command.Result{
		Title: pullResultTitle,
		Summary: []command.SummaryItem{
			{Label: "Project ID", Value: fmt.Sprintf("%d", result.ProjectID)},
			{Label: "Version ID", Value: fmt.Sprintf("%d", result.VersionID)},
			{Label: "History ID", Value: fmt.Sprintf("%d", result.HistoryID)},
			{Label: "Variable Count", Value: fmt.Sprintf("%d", result.VariableCount)},
			{Label: "Environment File", Value: result.EnvironmentFile},
			{Label: "Local Repository", Value: result.LocalRepository},
		},
	}, nil
}
