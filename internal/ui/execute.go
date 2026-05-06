package ui

import (
	"context"
	"io"

	"github.com/Team-NEEEE/envio-cli/internal/command"
)

func Execute(ctx context.Context, cmd command.Command, options Options) int {
	options = withDefaultWriters(options)
	switch options.selectedMode() {
	case ModeJSON:
		return runJSON(ctx, cmd, options)
	case ModeTUI:
		return runTUI(ctx, cmd, options)
	default:
		return runPlain(ctx, cmd, options)
	}
}

func RenderAppError(appErr *command.AppError, options Options) int {
	options = withDefaultWriters(options)
	switch options.selectedMode() {
	case ModeJSON:
		return renderJSONError(appErr, options)
	default:
		return renderPlainError(appErr, options)
	}
}

func withDefaultWriters(options Options) Options {
	if options.Output == nil {
		options.Output = io.Discard
	}
	if options.ErrOutput == nil {
		options.ErrOutput = io.Discard
	}
	return options
}
