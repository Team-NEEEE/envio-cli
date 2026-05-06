package ui

import (
	"io"

	"github.com/Team-NEEEE/envio-cli/internal/i18n"
)

type Options struct {
	Mode       RequestedMode
	Language   i18n.Language
	Debug      bool
	Env        map[string]string
	IsTerminal func() bool
	Input      io.Reader
	Output     io.Writer
	ErrOutput  io.Writer
}

func (o Options) selectedMode() RequestedMode {
	if o.Debug {
		return ModeJSON
	}
	isTerminal := false
	if o.IsTerminal != nil {
		isTerminal = o.IsTerminal()
	}
	return SelectMode(o.Mode, isTerminal, o.Env)
}

func (o Options) language() i18n.Language {
	if o.Language == "" {
		return i18n.English
	}
	return o.Language
}
