package selfcommand

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// NewUpgradeCommand creates the 'self upgrade' command
func NewUpgradeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade syst CLI to the latest release",
		RunE:  runUpgrade,
	}
}

func runUpgrade(cmd *cobra.Command, args []string) error {
	const repo = "redjax/syst"

	apiURL := "https://api.github.com/repos/" + repo + "/releases/latest"

	fmt.Fprintln(cmd.ErrOrStderr(), "Checking for latest release...")

	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("GitHub API returned status: %s", resp.Status)
	}

	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return fmt.Errorf("failed to parse release JSON: %w", err)
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "Latest version:", release.TagName)

	expected := fmt.Sprintf("syst_%s_%s", runtime.GOOS, runtime.GOARCH)
	var targetAssetURL string
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, expected) {
			targetAssetURL = asset.BrowserDownloadURL
			break
		}
	}
	if targetAssetURL == "" {
		return fmt.Errorf("no suitable binary found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download the new binary
	fmt.Fprintln(cmd.ErrOrStderr(), "Downloading:", targetAssetURL)
	assetResp, err := http.Get(targetAssetURL)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer assetResp.Body.Close()

	tmpBin, err := os.CreateTemp("", "syst-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpBin.Name())
	if _, err := io.Copy(tmpBin, assetResp.Body); err != nil {
		return fmt.Errorf("failed to write new binary: %w", err)
	}
	if err := tmpBin.Chmod(0755); err != nil {
		return fmt.Errorf("failed to set executable permissions: %w", err)
	}
	tmpBin.Close()

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine current executable path: %w", err)
	}
	// On POSIX, you can overwrite the running binary.
	// On Windows, you may need to write to a new file and move on next run.

	// Replace the file
	if err := os.Rename(tmpBin.Name(), exePath); err != nil {
		return fmt.Errorf("failed to replace running binary: %w", err)
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "Upgrade successful!")
	return nil
}
