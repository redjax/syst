package selfcommand

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/redjax/syst/internal/version"

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
	pkgInfo := version.GetPackageInfo()
	repo := fmt.Sprintf("%s/%s", pkgInfo.RepoUser, pkgInfo.RepoName)
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

	// Adjust if asset is named differently.
	//   You may need to do some serious reworking of this section for your release.
	normalizedOS := normalizeOS(runtime.GOOS)
	expected := fmt.Sprintf("%s-%s", normalizedOS, runtime.GOARCH)

	// Build URL from expected release name
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

	if runtime.GOOS == "windows" {
		newExePath := exePath + ".new"
		if err := copyFile(tmpBin.Name(), newExePath); err != nil {
			return fmt.Errorf("failed to save new binary: %w", err)
		}
		fmt.Fprintf(cmd.ErrOrStderr(),
			"\nWindows upgrade requires manual replacement.\n"+
				"A new binary was saved to:\n  %s\n"+
				"Please close this program and replace the existing exe with the new file.\n", newExePath)
		return nil
	}

	// POSIX (UNIX/macOS/Linux): Overwrite running binary
	if err := os.Rename(tmpBin.Name(), exePath); err != nil {
		return fmt.Errorf("failed to replace running binary: %w", err)
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "Upgrade successful!")
	return nil
}

// copyFile does a buffered copy from src to dst
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Chmod(0755)
}

func normalizeOS(goos string) string {
	switch goos {
	case "darwin":
		return "macOS"
	default:
		return goos
	}
}
