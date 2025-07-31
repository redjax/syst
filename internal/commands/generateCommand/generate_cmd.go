package generatecommand

import "github.com/spf13/cobra"

func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate",
		Short:   "Commands to generate various outputs, like UUIDs and OpenSSL secrets.",
		Aliases: []string{"gen"},
	}

	cmd.AddCommand(NewUUIDCommand())
	cmd.AddCommand(NewSecretsCommand())

	return cmd
}
