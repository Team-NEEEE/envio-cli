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

const createResultTitle = "Create completed"

type createOptions struct {
	repo string
}

type createService interface {
	Create(context.Context, string, string, command.Reporter) (*project.CreateResult, *command.AppError)
}

type createRunner struct {
	service       createService
	cwd           string
	repositoryURL string
}

func newCreateCommand(rt Runtime, lang i18n.Language, global *globalOptions, exitCode *int) *cobra.Command {
	options := createOptions{}
	cmd := &cobra.Command{
		Use:     "create <repository-url>",
		Short:   createShort(lang),
		Long:    createLong(lang),
		Example: createExample(lang),
		Args:    validateCreateArgs(&options),
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := ui.ModeAuto
			if global.plain {
				mode = ui.ModePlain
			}
			if global.json {
				mode = ui.ModeJSON
			}

			repositoryURL := strings.TrimSpace(options.repo)
			if len(args) == 1 {
				repositoryURL = strings.TrimSpace(args[0])
			}

			renderOptions := renderOptionsWithLanguage(
				rt,
				mode,
				lang,
				global.envDebug || global.debug || global.json,
			)
			runner := createRunner{
				service:       project.NewCreateService(global.apiURL),
				cwd:           rt.CWD,
				repositoryURL: repositoryURL,
			}
			code := ui.Execute(cmd.Context(), runner, renderOptions)
			if exitCode != nil {
				*exitCode = code
			}
			return nil
		},
	}
	applyHelpTemplate(cmd, lang)
	cmd.Flags().StringVar(&options.repo, "repo", "", flagText(lang, "repo"))
	return cmd
}

func validateCreateArgs(options *createOptions) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("accepts at most 1 arg(s), received %d", len(args))
		}

		hasPositional := len(args) == 1 && strings.TrimSpace(args[0]) != ""
		hasRepoFlag := options != nil && strings.TrimSpace(options.repo) != ""
		switch {
		case hasPositional && hasRepoFlag:
			return fmt.Errorf("use either repository-url argument or --repo, not both")
		case !hasPositional && !hasRepoFlag:
			return fmt.Errorf("repository URL is required for %q", cmd.CommandPath())
		default:
			return nil
		}
	}
}

func (r createRunner) Name() string {
	return "create"
}

func (r createRunner) Steps() []command.Step {
	return []command.Step{
		{ID: project.StepCheckRepository, Label: "Check repository", Status: command.StatusPending},
		{ID: project.StepCreateProject, Label: "Create project", Status: command.StatusPending},
		{ID: project.StepWrapProjectKey, Label: "Wrap project key", Status: command.StatusPending},
		{ID: project.StepSaveWrappedKeys, Label: "Save wrapped keys", Status: command.StatusPending},
		{ID: project.StepSaveProjectConfig, Label: "Save project config", Status: command.StatusPending},
	}
}

func (r createRunner) Run(ctx context.Context, reporter command.Reporter) (command.Result, *command.AppError) {
	if r.service == nil {
		return command.Result{}, command.NewAppError(
			project.ErrorCreateProjectFailed,
			"create service is not configured",
			"",
			1,
			command.SeverityError,
		)
	}

	result, appErr := r.service.Create(ctx, r.cwd, r.repositoryURL, reporter)
	if appErr != nil {
		return command.Result{}, appErr
	}

	return command.Result{
		Title: createResultTitle,
		Summary: []command.SummaryItem{
			{Label: "Project ID", Value: fmt.Sprintf("%d", result.ProjectID)},
			{Label: "Project Name", Value: result.ProjectName},
			{Label: "GitHub Repository", Value: result.GithubRepoName},
			{Label: "Local Repository", Value: result.LocalRepositoryRoot},
			{Label: "Link Config", Value: result.LocalConfigPath},
			{Label: "Wrapped Key Targets", Value: fmt.Sprintf("%d", result.WrappedKeyTargetCount)},
			{Label: "Updated Count", Value: fmt.Sprintf("%d", result.UpdatedCount)},
		},
	}, nil
}
