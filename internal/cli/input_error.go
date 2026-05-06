package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Team-NEEEE/envio-cli/internal/i18n"
)

func exactValidArgs(count int, valid []string) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(count)(cmd, args); err != nil {
			return err
		}
		for _, arg := range args {
			if !containsString(valid, arg) {
				return fmt.Errorf("invalid argument %q for %q", arg, cmd.CommandPath())
			}
		}
		return nil
	}
}

func renderErrorUsage(cmd *cobra.Command) string {
	return fmt.Sprintf("Usage:  %s", cmd.UseLine())
}

func renderInputError(root *cobra.Command, args []string, err error, lang i18n.Language) string {
	if root == nil || err == nil {
		return ""
	}

	target := commandForArgs(root, args)
	localizeBuiltInFlags(target, lang)

	sections := []string{
		err.Error(),
		renderErrorUsage(target),
	}
	if isUnknownCommandError(err) {
		if commands := renderAvailableCommands(target); commands != "" {
			sections = append(sections, commands)
		}
	} else if values := renderValidArgs(target); values != "" {
		sections = append(sections, values)
	}

	return strings.Join(sections, "\n\n") + "\n"
}

func renderAvailableCommands(cmd *cobra.Command) string {
	commands := availableCommands(cmd)
	if len(commands) == 0 {
		return ""
	}

	lines := []string{"Available commands:"}
	for _, command := range commands {
		lines = append(lines, "  "+command.Name())
	}
	return strings.Join(lines, "\n")
}

func renderValidArgs(cmd *cobra.Command) string {
	if len(cmd.ValidArgs) == 0 {
		return ""
	}

	lines := []string{"Available values:"}
	for _, value := range cmd.ValidArgs {
		lines = append(lines, "  "+value)
	}
	return strings.Join(lines, "\n")
}

func commandForArgs(root *cobra.Command, args []string) *cobra.Command {
	current := root
	skipNext := false

	for _, arg := range args {
		if skipNext {
			skipNext = false
			continue
		}
		if arg == "" {
			continue
		}
		if strings.HasPrefix(arg, "-") {
			if flagConsumesNextArg(current, arg) {
				skipNext = true
			}
			continue
		}

		child := matchingChild(current, arg)
		if child == nil {
			return current
		}
		current = child
	}

	return current
}

func matchingChild(cmd *cobra.Command, name string) *cobra.Command {
	for _, child := range cmd.Commands() {
		if child.Name() == name {
			return child
		}
		for _, alias := range child.Aliases {
			if alias == name {
				return child
			}
		}
	}
	return nil
}

func flagConsumesNextArg(cmd *cobra.Command, arg string) bool {
	if strings.Contains(arg, "=") {
		return false
	}

	name := strings.TrimLeft(arg, "-")
	if len(name) == 1 {
		flag := cmd.Flags().ShorthandLookup(name)
		return flag != nil && flag.Value.Type() != "bool"
	}

	flag := cmd.Flags().Lookup(name)
	return flag != nil && flag.Value.Type() != "bool"
}

func isUnknownCommandError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unknown command")
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
