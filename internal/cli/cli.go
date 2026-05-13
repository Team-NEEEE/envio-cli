package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Team-NEEEE/envio-cli/internal/command"
	"github.com/Team-NEEEE/envio-cli/internal/config"
	"github.com/Team-NEEEE/envio-cli/internal/i18n"
	"github.com/Team-NEEEE/envio-cli/internal/ui"
)

const defaultVersion = "dev"

var completionShells = []string{"bash", "zsh", "fish", "powershell"}

type Runtime struct {
	Stdin      io.Reader
	Stdout     io.Writer
	Stderr     io.Writer
	IsTerminal func() bool
	CWD        string
	Args       []string
	Environ    []string
}

type globalOptions struct {
	language string
	apiURL   string
	plain    bool
	json     bool
	debug    bool
	envDebug bool
}

func Run(ctx context.Context, rt Runtime) int {
	rt = normalizeRuntime(rt)

	env := config.EnvMap(rt.Environ)
	lang := detectLanguage(rt.Args, env)
	global := globalOptions{
		apiURL:   config.APIURLFromEnv(env),
		plain:    hasBoolFlag(rt.Args, "--plain"),
		json:     hasBoolFlag(rt.Args, "--json"),
		debug:    hasBoolFlag(rt.Args, "--debug"),
		envDebug: config.IsTruthy(env["ENVIO_DEBUG"]),
		language: env["ENVIO_LANG"],
	}

	exitCode := 0
	root := newRootCommand(rt, lang, &global, &exitCode)
	root.SetArgs(rt.Args)

	if err := root.ExecuteContext(ctx); err != nil {
		var appErr *command.AppError
		if errors.As(err, &appErr) || global.json || global.debug || global.envDebug {
			options := renderOptionsForError(rt, global)
			return ui.RenderAppError(appErrorFromCobra(err), options)
		}

		appErr = appErrorFromCobra(err)
		_, _ = fmt.Fprint(rt.Stderr, renderInputError(root, rt.Args, err, lang))
		return appErr.ExitCode
	}
	return exitCode
}

func normalizeRuntime(rt Runtime) Runtime {
	if rt.Stdout == nil {
		rt.Stdout = io.Discard
	}
	if rt.Stderr == nil {
		rt.Stderr = io.Discard
	}
	if rt.CWD == "" {
		rt.CWD = "."
	}
	return rt
}

func newRootCommand(rt Runtime, lang i18n.Language, global *globalOptions, exitCode *int) *cobra.Command {
	root := &cobra.Command{
		Use:           "envio <command>",
		Short:         rootShort(lang),
		Long:          rootLong(lang),
		Example:       rootExample(),
		Version:       defaultVersion,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.SetIn(rt.Stdin)
	root.SetOut(rt.Stdout)
	root.SetErr(rt.Stderr)
	applyHelpTemplate(root, lang)
	root.AddGroup(&cobra.Group{
		ID:    commandGroupAdditional,
		Title: commandGroupHeading(lang, commandGroupAdditional),
	})

	flags := root.PersistentFlags()
	flags.BoolVar(&global.plain, "plain", global.plain, flagText(lang, "plain"))
	flags.BoolVar(&global.json, "json", global.json, flagText(lang, "json"))
	flags.BoolVar(&global.debug, "debug", global.debug, flagText(lang, "debug"))
	flags.StringVar(&global.language, "lang", global.language, flagText(lang, "lang"))
	flags.StringVar(&global.apiURL, "api-url", global.apiURL, flagText(lang, "api-url"))
	if err := flags.MarkHidden("api-url"); err != nil {
		panic(fmt.Sprintf("hide api-url flag: %v", err))
	}

	root.AddCommand(newLoginCommand(rt, lang, global, exitCode))
	root.AddCommand(newCreateCommand(rt, lang, global, exitCode))
	root.AddCommand(newCompletionCommand(lang, root))
	return root
}

func newCompletionCommand(lang i18n.Language, root *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     completionShort(lang),
		GroupID:   commandGroupAdditional,
		Hidden:    true,
		ValidArgs: completionShells,
		Args:      exactValidArgs(1, completionShells),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(out)
			case "zsh":
				return root.GenZshCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, true)
			case "powershell":
				return root.GenPowerShellCompletion(out)
			default:
				return fmt.Errorf("invalid argument %q for %q", args[0], cmd.CommandPath())
			}
		},
	}
	applyHelpTemplate(cmd, lang)
	return cmd
}

func renderOptionsForError(rt Runtime, global globalOptions) ui.Options {
	lang, ok := i18n.Resolve(global.language)
	if !ok {
		lang = i18n.English
	}

	mode := ui.ModeAuto
	if global.plain {
		mode = ui.ModePlain
	}
	if global.json {
		mode = ui.ModeJSON
	}

	return renderOptionsWithLanguage(rt, mode, lang, global.envDebug || global.debug || global.json)
}

func renderOptionsWithLanguage(rt Runtime, mode ui.RequestedMode, lang i18n.Language, debug bool) ui.Options {
	env := config.EnvMap(rt.Environ)
	isTerminal := rt.IsTerminal
	if isTerminal == nil {
		isTerminal = func() bool {
			file, ok := rt.Stdout.(*os.File)
			if !ok {
				return false
			}
			info, err := file.Stat()
			return err == nil && info.Mode()&os.ModeCharDevice != 0
		}
	}

	return ui.Options{
		Mode:       mode,
		Language:   lang,
		Debug:      debug,
		Env:        env,
		IsTerminal: isTerminal,
		Input:      rt.Stdin,
		Output:     rt.Stdout,
		ErrOutput:  rt.Stderr,
	}
}

func appErrorFromCobra(err error) *command.AppError {
	if err == nil {
		return nil
	}
	var appErr *command.AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	message := err.Error()
	lower := strings.ToLower(message)
	switch {
	case strings.Contains(lower, "unknown command"):
		return command.NewAppError(
			"UNKNOWN_COMMAND",
			message,
			"Run `envio --help` to see available commands.",
			2,
			command.SeverityError,
		)
	case strings.Contains(lower, "flag needs an argument") || strings.Contains(lower, "requires an argument"):
		return command.NewAppError(
			"MISSING_FLAG_VALUE",
			message,
			"Pass a non-empty value after the flag.",
			2,
			command.SeverityError,
		)
	default:
		return command.NewAppError(
			"UNKNOWN_ARGUMENT",
			message,
			"Run `envio --help` to see the supported syntax.",
			2,
			command.SeverityError,
		)
	}
}

func detectLanguage(args []string, env map[string]string) i18n.Language {
	raw := stringFlagValue(args, "--lang", env["ENVIO_LANG"])
	lang, ok := i18n.Resolve(raw)
	if !ok {
		return i18n.English
	}
	return lang
}

func hasBoolFlag(args []string, name string) bool {
	for _, arg := range args {
		if arg == name {
			return true
		}
	}
	return false
}

func stringFlagValue(args []string, name, fallback string) string {
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == name && i+1 < len(args):
			return args[i+1]
		case strings.HasPrefix(args[i], name+"="):
			return strings.TrimPrefix(args[i], name+"=")
		}
	}
	return fallback
}
