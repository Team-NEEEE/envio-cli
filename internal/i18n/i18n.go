package i18n

import (
	"embed"
	"strings"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localeFS embed.FS

type Language string

const (
	Korean  Language = "ko"
	English Language = "en"
)

type Text struct {
	Message string
	Hint    string
}

var bundle = mustBundle()

func Resolve(raw string) (Language, bool) {
	normalized := strings.ToLower(strings.TrimSpace(raw))
	switch {
	case normalized == "", normalized == string(English), strings.HasPrefix(normalized, "en-"):
		return English, true
	case normalized == string(Korean), strings.HasPrefix(normalized, "ko-"):
		return Korean, true
	default:
		return English, false
	}
}

func ErrorText(lang Language, code, fallbackMessage, fallbackHint string) Text {
	return Text{
		Message: localize(lang, "error."+code+".message", fallbackMessage),
		Hint:    localize(lang, "error."+code+".hint", fallbackHint),
	}
}

func WarningText(lang Language, code, fallbackMessage, fallbackHint string) Text {
	return Text{
		Message: localize(lang, "warning."+code+".message", fallbackMessage),
		Hint:    localize(lang, "warning."+code+".hint", fallbackHint),
	}
}

func ResultTitle(lang Language, fallback string) string {
	return localize(lang, "result.title."+fallback, fallback)
}

func SummaryLabel(lang Language, fallback string) string {
	return localize(lang, "summary."+fallback, fallback)
}

func StepLabel(lang Language, id, fallback string) string {
	return localize(lang, "step."+id+".label", fallback)
}

func StepDetail(lang Language, key, fallback string) string {
	if key == "" {
		return fallback
	}
	return localize(lang, "step.detail."+key, fallback)
}

func HelpLabel(lang Language) string {
	return localize(lang, "ui.help.nextStep", "Hint")
}

func DebugLabel(lang Language) string {
	return localize(lang, "ui.debug.title", "Debug details")
}

func StatusLabel(lang Language, status string) string {
	return localize(lang, "ui.status."+status, status)
}

func mustBundle() *goi18n.Bundle {
	b := goi18n.NewBundle(language.MustParse(string(English)))
	for _, path := range []string{
		"locales/active.ko.json",
		"locales/active.en.json",
	} {
		if _, err := b.LoadMessageFileFS(localeFS, path); err != nil {
			panic(err)
		}
	}
	return b
}

func localize(lang Language, id, fallback string) string {
	localized, err := goi18n.NewLocalizer(bundle, string(lang)).Localize(&goi18n.LocalizeConfig{
		DefaultMessage: &goi18n.Message{
			ID:    id,
			Other: fallback,
		},
	})
	if err != nil && localized == "" {
		return fallback
	}
	return localized
}
