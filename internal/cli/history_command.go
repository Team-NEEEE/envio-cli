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

type historyService interface {
	ListHistory(context.Context, string, command.Reporter) (*project.HistoryListResult, *command.AppError)
	DecryptHistoryVersion(context.Context, string, string, command.Reporter) (*project.HistoryVersionResult, *command.AppError)
}

func newHistoryCommand(rt Runtime, lang i18n.Language, global *globalOptions, exitCode *int) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "history [version]",
		Short: historyShort(lang),
		Args:  validateOptionalHistoryVersionArg(),
		RunE: func(cmd *cobra.Command, args []string) error {
			version := ""
			if len(args) == 1 {
				version = strings.TrimSpace(args[0])
			}

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
			selectedMode := selectedHistoryMode(renderOptions, mode)
			code := runHistoryCommand(
				cmd.Context(),
				project.NewSyncService(global.apiURL),
				rt.CWD,
				version,
				selectedMode,
				renderOptions,
			)
			if exitCode != nil {
				*exitCode = code
			}
			return nil
		},
	}
	applyHelpTemplate(cmd, lang)
	return cmd
}

func validateOptionalHistoryVersionArg() cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("accepts at most 1 arg(s), received %d", len(args))
		}
		if len(args) == 1 && strings.TrimSpace(args[0]) == "" {
			return fmt.Errorf("history version must not be empty for %q", cmd.CommandPath())
		}
		return nil
	}
}

func runHistoryCommand(
	ctx context.Context,
	service historyService,
	cwd string,
	version string,
	selectedMode ui.RequestedMode,
	options ui.Options,
) int {
	if service == nil {
		return ui.RenderAppError(command.NewAppError(
			project.ErrorHistoryFailed,
			"history service is not configured",
			"",
			1,
			command.SeverityError,
		), options)
	}

	if strings.TrimSpace(version) == "" {
		list, appErr := service.ListHistory(ctx, cwd, command.NoopReporter{})
		if appErr != nil {
			return ui.RenderAppError(appErr, options)
		}
		histories := historyEntryViews(list.Histories)
		if selectedMode == ui.ModeTUI {
			selection, err := ui.PromptHistorySelection(options.Input, options.Output, histories, options.Language)
			if err != nil {
				return ui.RenderAppError(command.NewAppError(
					project.ErrorHistoryVersionNotFound,
					"history version was not selected",
					"Choose a version from the history list.",
					1,
					command.SeverityError,
				), options)
			}
			version = selection
		} else {
			if selectedMode == ui.ModeJSON {
				return ui.RenderHistoryListJSON(options.Output, histories)
			}
			return ui.RenderHistoryList(options.Output, histories, options.Language)
		}
	}

	result, appErr := service.DecryptHistoryVersion(ctx, cwd, version, command.NoopReporter{})
	if appErr != nil {
		return ui.RenderAppError(appErr, options)
	}
	view := historyVersionView(result)
	if selectedMode == ui.ModeJSON {
		return ui.RenderHistoryVersionJSON(options.Output, view)
	}
	return ui.RenderHistoryVersion(options.Output, view, options.Language)
}

func selectedHistoryMode(options ui.Options, requested ui.RequestedMode) ui.RequestedMode {
	if options.Debug {
		return ui.ModeJSON
	}
	isTerminal := false
	if options.IsTerminal != nil {
		isTerminal = options.IsTerminal()
	}
	return ui.SelectMode(requested, isTerminal, options.Env)
}

func historyEntryViews(histories []project.HistoryEntry) []ui.HistoryEntryView {
	views := make([]ui.HistoryEntryView, 0, len(histories))
	for _, history := range histories {
		views = append(views, ui.HistoryEntryView{
			ProjectID:     history.ProjectID,
			HistoryID:     history.HistoryID,
			VersionID:     history.VersionID,
			BaseVersionID: history.BaseVersionID,
			GithubID:      history.GithubID,
			CreatedAt:     history.CreatedAt,
			Latest:        history.Latest,
		})
	}
	return views
}

func historyVersionView(history *project.HistoryVersionResult) ui.HistoryVersionView {
	if history == nil {
		return ui.HistoryVersionView{}
	}
	return ui.HistoryVersionView{
		ProjectID:     history.ProjectID,
		HistoryID:     history.HistoryID,
		VersionID:     history.VersionID,
		BaseVersionID: history.BaseVersionID,
		GithubID:      history.GithubID,
		CreatedAt:     history.CreatedAt,
		Environment:   history.Environment,
		VariableCount: history.VariableCount,
	}
}
