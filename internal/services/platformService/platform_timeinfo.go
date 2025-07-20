package platformservice

import (
	"fmt"
	"strings"
	"time"
)

type TimeInfo struct {
	CurrentTime   string
	Timezone      string
	TimezoneLong  string
	OffsetSeconds int
}

func getTimeInfo() TimeInfo {
	now := time.Now() // local time, as set by OS
	zoneName, offsetSecs := now.Zone()

	// Try to get the IANA zone name if possible
	locationName := ""
	if loc := now.Location(); loc != nil {
		locationName = loc.String()
	}

	return TimeInfo{
		CurrentTime:   now.Format(time.RFC3339),
		Timezone:      zoneName,
		TimezoneLong:  locationName,
		OffsetSeconds: offsetSecs,
	}
}

func (p PlatformInfo) PrintTimeFormat() string {
	var builder strings.Builder

	builder.WriteString("Time Info:\n")

	builder.WriteString(fmt.Sprintf("  Current Time: %s\n", p.Time.CurrentTime))
	builder.WriteString(fmt.Sprintf("  Timezone: %s (%s)\n", p.Time.TimezoneLong, p.Time.Timezone))
	builder.WriteString(fmt.Sprintf("  Offset: %vs\n", p.Time.OffsetSeconds))

	return builder.String()
}
