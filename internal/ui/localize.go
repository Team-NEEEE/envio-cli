package ui

import (
	"github.com/Team-NEEEE/envio-cli/internal/command"
	"github.com/Team-NEEEE/envio-cli/internal/i18n"
)

type localizedSummaryItem struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type localizedWarning struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Hint    string `json:"hint,omitempty"`
}

type localizedError struct {
	Code     string           `json:"code"`
	Message  string           `json:"message"`
	Hint     string           `json:"hint,omitempty"`
	ExitCode int              `json:"exitCode"`
	Severity command.Severity `json:"severity"`
}

func localizedTitle(result command.Result, lang i18n.Language) string {
	return i18n.ResultTitle(lang, result.Title)
}

func localizedSummary(items []command.SummaryItem, lang i18n.Language) []localizedSummaryItem {
	public := make([]localizedSummaryItem, 0, len(items))
	for _, item := range items {
		if item.Sensitive {
			continue
		}
		public = append(public, localizedSummaryItem{
			Label: i18n.SummaryLabel(lang, item.Label),
			Value: item.Value,
		})
	}
	return public
}

func localizedWarnings(warnings []command.Warning, lang i18n.Language) []localizedWarning {
	localized := make([]localizedWarning, 0, len(warnings))
	for _, warning := range warnings {
		text := i18n.WarningText(lang, warning.Code, warning.Message, warning.Hint)
		localized = append(localized, localizedWarning{
			Code:    warning.Code,
			Message: text.Message,
			Hint:    text.Hint,
		})
	}
	return localized
}

func localizedAppError(appErr *command.AppError, lang i18n.Language) localizedError {
	if appErr == nil {
		appErr = command.NewAppError("UNKNOWN_ERROR", "unknown error", "", 1, command.SeverityError)
	}

	text := i18n.ErrorText(lang, appErr.CodeOrDefault(), appErr.Message, appErr.Hint)
	if appErr.Severity == command.SeverityWarning {
		text = i18n.WarningText(lang, appErr.CodeOrDefault(), appErr.Message, appErr.Hint)
	}

	return localizedError{
		Code:     appErr.CodeOrDefault(),
		Message:  text.Message,
		Hint:     text.Hint,
		ExitCode: appErr.ExitCode,
		Severity: appErr.Severity,
	}
}

func userFacingHint(appErr *command.AppError, lang i18n.Language) string {
	localized := localizedAppError(appErr, lang)
	if localized.Hint != "" {
		return localized.Hint
	}
	return localized.Message
}
