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
		// Detect location via JSON API call
		detectedLocation, err := detectLocation()
		if err != nil {
			return "", "", err
		}
		// Recursively fetch weather with detected location
		_, weatherText, err := fetchWttrWeather(detectedLocation, current, forecast)
		return detectedLocation, weatherText, err
	}

	// Fetch weather text for a given location with query params
	weatherText, err := getWttrWeatherText(location, current, forecast)
	return location, weatherText, err
}

// detectLocation calls the wttr.in JSON API to determine the detected location string
func detectLocation() (string, error) {
	url := "https://wttr.in/auto?format=j1"
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get detected location from wttr.in")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var w WttrJSON
	if err := json.Unmarshal(body, &w); err != nil {
		return "", err
	}

	if len(w.NearestArea) == 0 || len(w.NearestArea[0].AreaName) == 0 || len(w.NearestArea[0].Country) == 0 {
		return "", errors.New("no location info from wttr.in JSON response")
	}

	detectedLocation := fmt.Sprintf("%s, %s", w.NearestArea[0].AreaName[0].Value, w.NearestArea[0].Country[0].Value)
	return detectedLocation, nil
}

// getWttrWeatherText fetches the weather text (ASCII/plain) for the specified location and flags
func getWttrWeatherText(location string, current, forecast bool) (string, error) {
	var query string
	if current {
		query = "?format=3" // brief current weather
	} else if forecast {
		query = "" // full forecast (default)
	} else {
		query = "?format=3" // default: brief current weather
	}

	url := fmt.Sprintf("https://wttr.in/%s%s", location, query)
	// Build HTTP request with custom User-Agent to avoid HTML response
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "curl")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get weather info from wttr.in")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
