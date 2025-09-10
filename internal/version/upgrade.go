package version

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// UpgradeSelf is the entrypoint for 'syst self upgrade'.
// UpgradeSelf is the entrypoint for 'syst self upgrade'.
func UpgradeSelf(cmd *cobra.Command, args []string, checkOnly bool) error {
	if runtime.GOOS == "windows" {
		// Show current version information first
		info := GetPackageInfo()
		fmt.Printf("Current version: %s (commit: %s, built: %s)\n\n",
			info.PackageVersion,
			info.PackageCommit,
			info.PackageReleaseDate)

		scriptBlock := "if ($p = (Get-Command -Name syst -ErrorAction SilentlyContinue)) { Remove-Item $p.Path }; & ([scriptblock]::Create((irm https://raw.githubusercontent.com/redjax/syst/refs/heads/main/scripts/install-syst.ps1))) -Auto"
		ghIssueLink := "https://github.com/redjax/syst/issues/81"
		fmt.Printf("⚠️  The 'self upgrade' command does not work correctly on Windows yet: %s.\n\nTo upgrade, run this in your terminal:\n  %s\n", ghIssueLink, scriptBlock)
		os.Exit(0)
	}

	info := GetPackageInfo()

	repo, err := getRepoUrlPath()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error getting repository path (user/repo): %v\n", err)
		return err
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	fmt.Fprintln(cmd.ErrOrStderr(), "Checking for latest release...")

	resp, err := http.Get(apiURL)
	if err != nil {
		return fmt.Errorf("failed to fetch latest release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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

	current := info.PackageVersion
	latest := release.TagName

	fmt.Fprintln(cmd.ErrOrStderr(), "Current version:", current)
	fmt.Fprintln(cmd.ErrOrStderr(), "Latest version: ", latest)

	if current == "dev" {
		fmt.Fprintf(cmd.ErrOrStderr(), "🛠️  This is a development release: %s\n", current)
		return nil
	}

	cmp := compareVersion(current, latest)

	switch cmp {
	case -1:
		fmt.Fprintf(cmd.ErrOrStderr(), "🚀 Upgrade available: %s → %s\n", current, latest)
		if checkOnly {
			fmt.Fprintln(cmd.ErrOrStderr(), "✅ Use this command without --check to upgrade.")
			return nil
		}
	case 0:
		fmt.Fprintf(cmd.ErrOrStderr(), "🔄 No new release available, syst is up to date (%s).\n", current)
		return nil
	case 1:
		fmt.Fprintf(cmd.ErrOrStderr(), "🤯 You're ahead of the latest release: current=%s, release=%s\n", current, latest)
		return nil
	}

	normalizedOS := normalizeOS(runtime.GOOS)
	arch := runtime.GOARCH

	// Prepare expected prefix with syst-<os>-<arch>-
	expectedPrefixLower := fmt.Sprintf("syst-%s-%s-", strings.ToLower(normalizedOS), strings.ToLower(arch))
	expectedPrefixMacOS := fmt.Sprintf("syst-macOS-%s-", arch) // preserve macOS casing as assets use it exactly

	var assetURL string
	for _, asset := range release.Assets {
		if asset.Name == "" {
			continue
		}
		if normalizedOS == "macOS" {
			// macOS casing exact match
			if strings.HasPrefix(asset.Name, expectedPrefixMacOS) && strings.HasSuffix(asset.Name, ".zip") {
				assetURL = asset.BrowserDownloadURL
				break
			}
		} else {
			// case-insensitive match for linux/windows
			if strings.HasPrefix(strings.ToLower(asset.Name), expectedPrefixLower) && strings.HasSuffix(strings.ToLower(asset.Name), ".zip") {
				assetURL = asset.BrowserDownloadURL
				break
			}
		}
	}

	if assetURL == "" {
		fmt.Fprintln(cmd.ErrOrStderr(), "Available assets:")
		for _, asset := range release.Assets {
			fmt.Fprintln(cmd.ErrOrStderr(), " -", asset.Name)
		}
		return fmt.Errorf("no suitable release found for platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	fmt.Fprintln(cmd.ErrOrStderr(), "Downloading:", assetURL)

	resp2, err := http.Get(assetURL)
	if err != nil {
		return fmt.Errorf("failed to download binary zip: %w", err)
	}
	defer resp2.Body.Close()

	zipTmp, err := os.CreateTemp("", "syst-upgrade-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp zip file: %w", err)
	}
	defer os.Remove(zipTmp.Name())

	if _, err := io.Copy(zipTmp, resp2.Body); err != nil {
		return fmt.Errorf("failed to write zip file: %w", err)
	}
	zipTmp.Close()

	binaryTmp, err := extractBinaryFromZip(zipTmp.Name())
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}
	defer os.Remove(binaryTmp)

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	newPath := exePath + ".new"
	if err := copyFile(binaryTmp, newPath); err != nil {
		if os.IsPermission(err) {
			fmt.Fprintln(cmd.ErrOrStderr(), "Permission denied: try running with 'sudo syst self upgrade'")
		}
		return fmt.Errorf("failed to save new binary: %w", err)
	}

	fmt.Fprintf(cmd.ErrOrStderr(),
		"✅ Upgrade downloaded:\n  %s\n"+
			"  It will be applied next time you run a syst command, i.e. syst --version.\n",
		newPath)

	return nil
}

// normalizeOS maps runtime.GOOS to your release asset naming
func normalizeOS(goos string) string {
	switch strings.ToLower(goos) {
	case "darwin":
		return "macOS" // Keep casing as 'macOS' because the assets use it exactly
	default:
		return strings.ToLower(goos)
	}
}

// extractBinaryFromZip extracts the binary from a zip file and returns path
func extractBinaryFromZip(zipPath string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		defer rc.Close()

		tmpBin, err := os.CreateTemp("", "syst-bin-*")
		if err != nil {
			return "", err
		}

		if _, err := io.Copy(tmpBin, rc); err != nil {
			tmpBin.Close()
			return "", err
		}

		tmpBin.Close()

		if err := os.Chmod(tmpBin.Name(), 0755); err != nil {
			return "", err
		}

		return tmpBin.Name(), nil
	}

	return "", fmt.Errorf("no binary found in zip archive")
}

// copyFile utility
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

// TrySelfUpgrade checks if "<binary>.new" exists and replaces current binary with it.
func TrySelfUpgrade() {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get executable path: %v\n", err)
		return
	}

	newPath := exePath + ".new"

	if _, err := os.Stat(newPath); err == nil {
		// New file exists: perform replacement

		if runtime.GOOS == "windows" {
			// Use Windows-specific updater
			err := RunWindowsSelfUpgrade(exePath, newPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Syst Windows self-upgrade failed: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "🔁 syst upgraded successfully.\n")
				// Exit after successful upgrade so new exe is run by RunWindowsSelfUpgrade
				os.Exit(0)
			}
		}
		errRename := os.Rename(newPath, exePath)

		if errRename != nil {
			fmt.Fprintf(os.Stderr, "Failed to replace executable: %v\n", errRename)
		} else {
			fmt.Fprintf(os.Stderr, "🔁 syst upgraded successfully.\n")
		}
	}
}
