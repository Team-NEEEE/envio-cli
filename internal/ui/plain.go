package ui

import (
	"context"
	"fmt"

	"github.com/Team-NEEEE/envio-cli/internal/command"
	"github.com/Team-NEEEE/envio-cli/internal/i18n"
)

const plainLabelValueFormat = "%s: %s\n"

func runPlain(ctx context.Context, cmd command.Command, options Options) int {
	result, appErr := cmd.Run(ctx, command.NoopReporter{})
	if appErr != nil {
		return renderPlainError(appErr, options)
	}
	return renderPlainResult(result, options)
}

func renderPlainResult(result command.Result, options Options) int {
	lang := options.language()
	title := localizedTitle(result, lang)
	if len(result.Warnings) > 0 {
		_, _ = fmt.Fprintf(options.Output, plainLabelValueFormat, i18n.StatusLabel(lang, "WARN"), title)
	} else {
		_, _ = fmt.Fprintf(options.Output, plainLabelValueFormat, i18n.StatusLabel(lang, "OK"), title)
	}
	for _, item := range localizedSummary(result.Summary, lang) {
		_, _ = fmt.Fprintf(options.Output, plainLabelValueFormat, item.Label, item.Value)
	}
	for _, warning := range localizedWarnings(result.Warnings, lang) {
		hint := warning.Hint
		if hint == "" {
			hint = warning.Message
		}
		_, _ = fmt.Fprintf(options.Output, plainLabelValueFormat, i18n.HelpLabel(lang), hint)
	}
	return 0
}

func renderPlainError(appErr *command.AppError, options Options) int {
	if appErr == nil {
		return 1
	}
	lang := options.language()
	_, _ = fmt.Fprintf(options.ErrOutput, plainLabelValueFormat, i18n.HelpLabel(lang), userFacingHint(appErr, lang))
	return appErr.ExitCode
}
