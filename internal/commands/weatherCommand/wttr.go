package weathercommand

import (
	"fmt"
	"os"

	weatherservice "github.com/redjax/syst/internal/services/weatherService"
	"github.com/spf13/cobra"
)

func NewWttrCommand() *cobra.Command {
	var current bool
	var forecast bool
	var location string

	cmd := &cobra.Command{
		Use:   "wttr",
		Short: "Get weather from wttr.in",
		RunE: func(cmd *cobra.Command, args []string) error {
			effectiveLocation := location

			// If CLI flag not set, try env var
			if effectiveLocation == "" {
				if envLoc := os.Getenv("WTTR_LOCATION"); envLoc != "" {
					effectiveLocation = envLoc
				}
			}

			_, weatherText, err := weatherservice.FetchWeather("wttr", effectiveLocation, current, forecast)
			if err != nil {
				return err
			}

			fmt.Println(weatherText)
			return nil

		},
	}

	cmd.Flags().BoolVar(&current, "current", false, "Show current weather")
	cmd.Flags().BoolVar(&forecast, "forecast", false, "Show weather forecast")
	cmd.Flags().StringVarP(&location, "location", "l", "", "Location for weather info (default auto-detected)")

	return cmd
}
