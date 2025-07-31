package generatecommand

import "github.com/spf13/cobra"

func NewGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Commands to generate various outputs, like UUIDs and OpenSSL secrets.",
	}

	cmd.AddCommand(NewUUIDCommand())

	return cmd
}
