package weathercommand

import "github.com/spf13/cobra"

func NewWeatherCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "weather",
		Short: "Weather related commands",
		Long: `Get weather information from various weather providers.

Description:
  Uses subcommands to request weather from different sources.
  Pass an argument to use a specific API, like "syst weather wttr"
`,
	}

	cmd.AddCommand(NewWttrCommand())

	return cmd
}
