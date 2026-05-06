package cli

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/Team-NEEEE/envio-cli/internal/i18n"
)

func applyHelpTemplate(cmd *cobra.Command, lang i18n.Language) {
	cmd.SetHelpFunc(func(cmd *cobra.Command, _ []string) {
		localizeBuiltInFlags(cmd, lang)
		_, _ = fmt.Fprint(cmd.OutOrStdout(), renderHelp(cmd, lang))
	})
	cmd.SetUsageFunc(func(cmd *cobra.Command) error {
		localizeBuiltInFlags(cmd, lang)
		_, err := fmt.Fprint(cmd.ErrOrStderr(), renderErrorUsage(cmd)+"\n")
		return err
	})
}

func localizeBuiltInFlags(cmd *cobra.Command, lang i18n.Language) {
	if flag := cmd.LocalFlags().Lookup("help"); flag != nil {
		flag.Usage = flagText(lang, "help")
	}
	if flag := cmd.LocalFlags().Lookup("version"); flag != nil {
		flag.Usage = flagText(lang, "version")
	}
}

type cliHelpTheme struct {
	heading lipgloss.Style
	command lipgloss.Style
}

func newCLIHelpTheme() cliHelpTheme {
	return cliHelpTheme{
		heading: lipgloss.NewStyle().Bold(true),
		command: lipgloss.NewStyle().Bold(true),
	}
}

func renderHelp(cmd *cobra.Command, lang i18n.Language) string {
	theme := newCLIHelpTheme()
	sections := make([]string, 0, 8)

	if description := strings.TrimSpace(cmd.Long); description != "" {
		sections = append(sections, description)
	}
	sections = append(sections, renderRootUsage(cmd, lang, theme))

	if commands := renderCommandGroups(cmd, lang, theme); commands != "" {
		sections = append(sections, commands)
	}
	if flags := renderFlagBlock(flagsHeading(lang), cmd.LocalFlags().FlagUsages(), theme); flags != "" {
		sections = append(sections, flags)
	}
	if flags := renderFlagBlock(flagsHeading(lang), cmd.InheritedFlags().FlagUsages(), theme); flags != "" {
		sections = append(sections, flags)
	}
	if examples := renderExamples(cmd, lang, theme); examples != "" {
		sections = append(sections, examples)
	}
	if learnMore := renderLearnMore(cmd, lang, theme); learnMore != "" {
		sections = append(sections, learnMore)
	}

	return strings.Join(sections, "\n\n") + "\n"
}

func renderRootUsage(cmd *cobra.Command, lang i18n.Language, theme cliHelpTheme) string {
	return fmt.Sprintf("%s\n  %s", theme.heading.Render(usageHeading(lang)), cmd.UseLine())
}

func renderCommandGroups(cmd *cobra.Command, lang i18n.Language, theme cliHelpTheme) string {
	commands := availableCommands(cmd)
	if len(commands) == 0 {
		return ""
	}

	grouped := map[string][]*cobra.Command{}
	for _, command := range commands {
		groupID := command.GroupID
		if groupID == "" {
			groupID = commandGroupAdditional
		}
		grouped[groupID] = append(grouped[groupID], command)
	}

	sections := make([]string, 0, len(grouped))
	for _, groupID := range []string{commandGroupAdditional} {
		if groupCommands := grouped[groupID]; len(groupCommands) > 0 {
			sections = append(sections, renderDescribedCommands(commandGroupHeading(lang, groupID), groupCommands, theme))
		}
	}
	return strings.Join(sections, "\n\n")
}

func renderDescribedCommands(title string, commands []*cobra.Command, theme cliHelpTheme) string {
	width := commandNameWidth(commands)
	lines := []string{theme.heading.Render(title)}
	for _, command := range commands {
		lines = append(lines, fmt.Sprintf("  %-*s  %s", width+1, theme.command.Render(command.Name()+":"), command.Short))
	}
	return strings.Join(lines, "\n")
}

func renderFlagBlock(title, flagUsages string, theme cliHelpTheme) string {
	flagUsages = strings.TrimRight(flagUsages, "\r\n\t ")
	if flagUsages == "" {
		return ""
	}
	return theme.heading.Render(title) + "\n" + flagUsages
}

func renderExamples(cmd *cobra.Command, lang i18n.Language, theme cliHelpTheme) string {
	example := strings.TrimSpace(cmd.Example)
	if example == "" {
		return ""
	}
	return theme.heading.Render(examplesHeading(lang)) + "\n" + indentLines(example, "  ")
}

func renderLearnMore(cmd *cobra.Command, lang i18n.Language, theme cliHelpTheme) string {
	if cmd.HasParent() {
		return ""
	}
	line := fmt.Sprintf("Use `%s <command> --help` for more information about a command.", cmd.Name())
	if lang == i18n.Korean {
		line = fmt.Sprintf("명령어별 도움말은 `%s <command> --help`로 확인하세요.", cmd.Name())
	}
	return theme.heading.Render(learnMoreHeading(lang)) + "\n" + indentLines(line, "  ")
}

func availableCommands(cmd *cobra.Command) []*cobra.Command {
	commands := make([]*cobra.Command, 0, len(cmd.Commands()))
	for _, command := range cmd.Commands() {
		if command.IsAvailableCommand() && command.Name() != "help" {
			commands = append(commands, command)
		}
	}
	return commands
}

func commandNameWidth(commands []*cobra.Command) int {
	width := 12
	for _, command := range commands {
		if len(command.Name()) > width {
			width = len(command.Name())
		}
	}
	return width
}

func indentLines(value, prefix string) string {
	lines := strings.Split(value, "\n")
	for i := range lines {
		if strings.TrimSpace(lines[i]) != "" {
			lines[i] = prefix + lines[i]
		}
	}
	return strings.Join(lines, "\n")
}
