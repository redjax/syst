package weatherservice

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type WttrJSON struct {
	NearestArea []struct {
		AreaName []struct {
			Value string `json:"value"`
		} `json:"areaName"`
		Country []struct {
			Value string `json:"value"`
		} `json:"country"`
	} `json:"nearest_area"`
	CurrentCondition []struct {
		TempC string `json:"temp_C"`
		// Add other fields if needed
	} `json:"current_condition"`
}

// fetchWttrWeather returns the *actual* location used and the weather text.
func fetchWttrWeather(location string, current, forecast bool) (string, string, error) {
	if location == "" {
		// First, detect location via JSON API call
		url := "https://wttr.in/auto?format=j1"
		resp, err := http.Get(url)
		if err != nil {
			return "", "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return "", "", errors.New("failed to get weather info from wttr.in")
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", "", err
		}

		var w WttrJSON
		if err := json.Unmarshal(body, &w); err != nil {
			return "", "", err
		}

		if len(w.NearestArea) == 0 || len(w.NearestArea[0].AreaName) == 0 || len(w.NearestArea[0].Country) == 0 {
			return "", "", errors.New("no location info from wttr.in JSON response")
		}

		// Construct detected location string
		detectedLocation := fmt.Sprintf("%s, %s", w.NearestArea[0].AreaName[0].Value, w.NearestArea[0].Country[0].Value)

		// Recursively call fetchWttrWeather with detected location to get weather text
		_, weatherText, err := fetchWttrWeather(detectedLocation, current, forecast)
		return detectedLocation, weatherText, err
	}

	// Compose query parameters based on flags
	var query string
	if current {
		query = "?format=3" // brief current weather
	} else if forecast {
		query = "" // full forecast (default)
	} else {
		query = "?format=3" // default to current weather brief
	}

	url := fmt.Sprintf("https://wttr.in/%s%s", location, query)
	resp, err := http.Get(url)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", errors.New("failed to get weather info from wttr.in")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	return location, string(body), nil
}
