package weatherservice

import (
	"errors"
)

// FetchWeather dispatches weather queries to the requested service.
// It returns the actual location used (detected or user passed) and the weather string.
func FetchWeather(serviceName, location string, current, forecast bool) (string, string, error) {
	switch serviceName {
	case "wttr":
		return fetchWttrWeather(location, current, forecast)
	// case "otherservice":
	//     return fetchOtherServiceWeather(...)
	default:
		return "", "", errors.New("unsupported weather service: " + serviceName)
	}
}
