package pingService

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestCreateLogFilePath(t *testing.T) {
	path := createLogFilePath("example.com")
	if path == "" {
		t.Error("createLogFilePath returned empty string")
	}
	if !strings.Contains(path, "example.com") {
		t.Errorf("createLogFilePath(%q) = %q, expected to contain target name", "example.com", path)
	}
	if !strings.HasSuffix(path, ".log") {
		t.Errorf("createLogFilePath(%q) = %q, expected .log suffix", "example.com", path)
	}
}

func TestCreateLogFilePath_SanitizesTarget(t *testing.T) {
	// Colons and slashes in the target should be replaced in the filename
	path := createLogFilePath("http://example.com:8080")
	filename := filepath.Base(path)
	if strings.Contains(filename, ":") || strings.Contains(filename, "/") {
		t.Errorf("filename should not contain colons or slashes: %q", filename)
	}
}

func TestProtocolLabel(t *testing.T) {
	tests := []struct {
		name    string
		useHTTP bool
		want    string
	}{
		{"http mode", true, "HTTP"},
		{"icmp mode", false, "ICMP"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &Options{UseHTTP: tt.useHTTP}
			got := protocolLabel(opts)
			if got != tt.want {
				t.Errorf("protocolLabel() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPingStats_Initialization(t *testing.T) {
	stats := &PingStats{}
	if stats.Total != 0 || stats.Successes != 0 || stats.Failures != 0 {
		t.Error("PingStats should initialize to zero values")
	}
}

func newTestContext() context.Context {
	return context.Background()
}

func TestSleepOrCancel(t *testing.T) {
	t.Run("completes normally for short sleep", func(t *testing.T) {
		ctx := newTestContext()
		result := sleepOrCancel(ctx, 1*time.Millisecond)
		if !result {
			t.Error("sleepOrCancel should return true (completed normally) for a short sleep")
		}
	})
}

func TestRunPing_EmptyTarget(t *testing.T) {
	opts := &Options{
		Target:  "",
		Count:   1,
		Sleep:   time.Second,
		UseHTTP: true,
		Ctx:     context.Background(),
		Stats:   &PingStats{},
	}
	err := RunPing(opts)
	if err == nil {
		t.Error("RunPing with empty target should return error")
	}
}

func TestRunPing_WhitespaceTarget(t *testing.T) {
	opts := &Options{
		Target:  "   ",
		Count:   1,
		Sleep:   time.Second,
		UseHTTP: true,
		Ctx:     context.Background(),
		Stats:   &PingStats{},
	}
	err := RunPing(opts)
	if err == nil {
		t.Error("RunPing with whitespace target should return error")
	}
}
