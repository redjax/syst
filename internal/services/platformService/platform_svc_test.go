package platformservice

import (
	"strings"
	"testing"
	"time"
)

func TestPlatformInfo_PrintFormat(t *testing.T) {
	info := PlatformInfo{
		Hostname:     "testhost",
		OS:           "linux",
		Arch:         "amd64",
		OSRelease:    "24.04",
		CurrentUser:  PlatformUser{Username: "testuser", Name: "Test User"},
		DefaultShell: "/bin/bash",
		UserHomeDir:  "/home/testuser",
		Uptime:       2 * time.Hour,
		TotalRAM:     8 * 1024 * 1024 * 1024,
		CPUCores:     4,
		CPUThreads:   8,
		CPUSockets:   1,
		CPUModel:     "Test CPU",
		CPUVendor:    "TestVendor",
	}

	output := info.PrintFormat(false, false)

	checks := []string{
		"testhost",
		"linux",
		"amd64",
		"24.04",
		"testuser",
		"/bin/bash",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("PrintFormat output missing %q", check)
		}
	}
}

func TestPlatformInfo_PrintFormat_WithNetwork(t *testing.T) {
	info := PlatformInfo{
		Hostname: "testhost",
		OS:       "linux",
		Arch:     "amd64",
		Interfaces: []NetworkInterface{
			{Name: "eth0", IPAddresses: []string{"192.168.1.10"}},
		},
		DNSServers: []string{"8.8.8.8"},
		GatewayIPs: []string{"192.168.1.1"},
	}

	output := info.PrintFormat(true, false)
	if !strings.Contains(output, "Network") || !strings.Contains(output, "eth0") {
		t.Errorf("PrintFormat(includeNet=true) should include network info")
	}
}

func TestGatherPlatformInfo(t *testing.T) {
	info, err := GatherPlatformInfo(false)
	if err != nil {
		t.Fatalf("GatherPlatformInfo error: %v", err)
	}
	if info == nil {
		t.Fatal("GatherPlatformInfo returned nil")
	}
	if info.OS == "" {
		t.Error("OS is empty")
	}
	if info.Arch == "" {
		t.Error("Arch is empty")
	}
	if info.Hostname == "" {
		t.Error("Hostname is empty")
	}
}
