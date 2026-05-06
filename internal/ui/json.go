package ui

import (
	"context"
	"encoding/json"

	"github.com/Team-NEEEE/envio-cli/internal/command"
)

type jsonResult struct {
	Error    *localizedError        `json:"error,omitempty"`
	Status   command.Status         `json:"status"`
	Title    string                 `json:"title,omitempty"`
	Summary  []localizedSummaryItem `json:"summary,omitempty"`
	Warnings []localizedWarning     `json:"warnings,omitempty"`
}

func runJSON(ctx context.Context, cmd command.Command, options Options) int {
	result, appErr := cmd.Run(ctx, command.NoopReporter{})
	if appErr != nil {
		return renderJSONError(appErr, options)
	}
	return renderJSONResult(result, options)
}

func renderJSONResult(result command.Result, options Options) int {
	lang := options.language()
	status := command.StatusSuccess
	if len(result.Warnings) > 0 {
		status = command.StatusWarning
	}
	payload := jsonResult{
		Status:   status,
		Title:    localizedTitle(result, lang),
		Summary:  localizedSummary(result.Summary, lang),
		Warnings: localizedWarnings(result.Warnings, lang),
	}
	encoder := json.NewEncoder(options.Output)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		return 1
	}
	return 0
}

func renderJSONError(appErr *command.AppError, options Options) int {
	if appErr == nil {
		appErr = command.NewAppError("UNKNOWN_ERROR", "unknown error", "", 1, command.SeverityError)
	}
	status := command.StatusError
	if appErr.Severity == command.SeverityWarning {
		status = command.StatusWarning
	}
	localized := localizedAppError(appErr, options.language())
	payload := jsonResult{Status: status, Error: &localized}
	encoder := json.NewEncoder(options.ErrOutput)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(payload); err != nil {
		return appErr.ExitCode
	}
	return appErr.ExitCode
}
