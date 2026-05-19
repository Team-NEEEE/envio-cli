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

const linkResultTitle = "Link completed"

type linkService interface {
	Link(context.Context, string, string, command.Reporter) (*project.LinkResult, *command.AppError)
}

type linkRunner struct {
	service       linkService
	cwd           string
	repositoryURL string
}

func newLinkCommand(rt Runtime, lang i18n.Language, global *globalOptions, exitCode *int) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "link [repository-url]",
		Short:   linkShort(lang),
		Long:    linkLong(lang),
		Example: linkExample(lang),
		Args:    validateLinkArgs(),
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := ui.ModeAuto
			if global.plain {
				mode = ui.ModePlain
			}
			if global.json {
				mode = ui.ModeJSON
			}

			repositoryURL := ""
			if len(args) == 1 {
				repositoryURL = strings.TrimSpace(args[0])
			}

			renderOptions := renderOptionsWithLanguage(
				rt,
				mode,
				lang,
				global.envDebug || global.debug || global.json,
			)
			runner := linkRunner{
				service:       project.NewLinkService(global.apiURL),
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
	return cmd
}

func validateLinkArgs() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("accepts at most 1 arg(s), received %d", len(args))
		}
		if len(args) == 1 && strings.TrimSpace(args[0]) == "" {
			return fmt.Errorf("repository URL must not be empty for %q", cmd.CommandPath())
		}
		return nil
	}
}

func (r linkRunner) Name() string {
	return "link"
}

func (r linkRunner) Steps() []command.Step {
	return []command.Step{
		{ID: project.StepCheckRepository, Label: "Check repository", Status: command.StatusPending},
		{ID: project.StepLinkProject, Label: "Link project", Status: command.StatusPending},
		{ID: project.StepUnwrapProjectKey, Label: "Unwrap project key", Status: command.StatusPending},
		{ID: project.StepSaveLinkConfig, Label: "Save link config", Status: command.StatusPending},
	}
}

func (r linkRunner) Run(ctx context.Context, reporter command.Reporter) (command.Result, *command.AppError) {
	if r.service == nil {
		return command.Result{}, command.NewAppError(
			project.ErrorLinkProjectFailed,
			"link service is not configured",
			"",
			1,
			command.SeverityError,
		)
	}

	result, appErr := r.service.Link(ctx, r.cwd, r.repositoryURL, reporter)
	if appErr != nil {
		return command.Result{}, appErr
	}

	summary := []command.SummaryItem{
		{Label: "Project ID", Value: fmt.Sprintf("%d", result.ProjectID)},
		{Label: "Project Name", Value: result.ProjectName},
		{Label: "Owner", Value: result.Owner},
		{Label: "Repository", Value: result.RepoName},
		{Label: "GitHub Repository", Value: result.GithubRepoName},
	}
	if strings.TrimSpace(result.JoinStatus) != "" {
		summary = append(summary, command.SummaryItem{Label: "Join Status", Value: result.JoinStatus})
	}
	summary = append(summary,
		command.SummaryItem{Label: "Local Repository", Value: result.LocalRepositoryRoot},
		command.SummaryItem{Label: "Link Config", Value: result.LocalConfigPath},
	)

	return command.Result{
		Title:   linkResultTitle,
		Summary: summary,
	}, nil
}
