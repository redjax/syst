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
func UpgradeSelf(cmd *cobra.Command, args []string, checkOnly bool) error {
	info := GetPackageInfo()

	repo, err := getRepoUrlPath()
	if err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Error getting repository path (user/repo): %v\n", err)
		return err
	}

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)
	fmt.Fprintln(cmd.ErrOrStderr(), "Checking for latest release...")

	// #nosec G107 - URL is constructed from hardcoded GitHub API endpoint and repo constant
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

	// #nosec G107 - URL is from GitHub release API response, validated to be from github.com
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
	// #nosec G104 - Close error is non-critical, file is fully written
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

	// Look for the binary file specifically (not README, LICENSE, etc.)
	var binaryFile *zip.File
	expectedBinaryName := "syst"
	if runtime.GOOS == "windows" {
		expectedBinaryName = "syst.exe"
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		// Match the binary file by name (case-insensitive)
		fileName := strings.ToLower(f.Name)
		if fileName == expectedBinaryName || strings.HasSuffix(fileName, "/"+expectedBinaryName) {
			binaryFile = f
			break
		}
	}

	if binaryFile == nil {
		return "", fmt.Errorf("binary '%s' not found in zip archive", expectedBinaryName)
	}

	rc, err := binaryFile.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()

	tmpBin, err := os.CreateTemp("", "syst-bin-*")
	if err != nil {
		return "", err
	}

	// Limit extraction size to 500MB to prevent decompression bomb attacks
	// #nosec G110 - Size limit implemented via io.LimitReader
	limitedReader := io.LimitReader(rc, 500*1024*1024) // 500MB max
	if _, err := io.Copy(tmpBin, limitedReader); err != nil {
		// #nosec G104 - Error from Close is non-critical here, primary error is from Copy
		tmpBin.Close()
		return "", err
	}

	// #nosec G104 - Error from Close checked below via Chmod
	tmpBin.Close()

	// #nosec G302 - Binary must be executable (0755 is appropriate for executables)
	if err := os.Chmod(tmpBin.Name(), 0755); err != nil {
		return "", err
	}

	return tmpBin.Name(), nil
}

// copyFile utility
func copyFile(src, dst string) error {
	// #nosec G304 - CLI tool copies files during self-upgrade by design
	in, err := os.Open(src)

	if err != nil {
		return err
	}
	defer in.Close()

	// #nosec G304 - CLI tool creates files during self-upgrade by design
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
			// Use Windows-specific updater (launches background script and exits)
			err := RunWindowsSelfUpgrade(exePath, newPath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "syst Windows self-upgrade failed: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "🔁 Applying upgrade...\n")
				// Exit immediately so the background script can replace the .exe
				os.Exit(0)
			}
			return
		}

		// Unix: direct rename works because the file isn't locked
		errRename := os.Rename(newPath, exePath)

		if errRename != nil {
			fmt.Fprintf(os.Stderr, "Failed to replace executable: %v\n", errRename)
		} else {
			fmt.Fprintf(os.Stderr, "🔁 syst upgraded successfully.\n")
		}
	}
}
