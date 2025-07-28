package weathercommand

import "github.com/spf13/cobra"

func NewWeatherCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "weather",
		Short: "Weather related commands",
		Long:  "Get weather information from various weather providers.",
	}

	cmd.AddCommand(NewWttrCommand())

	return cmd
}
