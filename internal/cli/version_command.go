package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/Team-NEEEE/envio-cli/internal/i18n"
)

func newVersionCommand(lang i18n.Language, version, commit, date string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "version",
		Short:   versionShort(lang),
		GroupID: commandGroupAdditional,
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			out := cmd.OutOrStdout()
			_, _ = fmt.Fprintf(out, "envio version %s\n", version)
			_, _ = fmt.Fprintf(out, "commit: %s\n", commit)
			_, _ = fmt.Fprintf(out, "built: %s\n", date)
			return nil
		},
	}
	applyHelpTemplate(cmd, lang)
	return cmd
}
