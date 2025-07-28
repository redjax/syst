package weathercommand

import (
	"fmt"

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
			usedLocation, weatherText, err := weatherservice.FetchWeather("wttr", location, current, forecast)
			if err != nil {
				return err
			}

			// If the user did not specify a location, show the detected location
			if location == "" {
				fmt.Printf("Detected Location: %s\n", usedLocation)
			} else {
				fmt.Printf("Location: %s\n", usedLocation)
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
