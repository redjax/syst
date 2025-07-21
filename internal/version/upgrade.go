package version

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// When --check is passed, don't do an upgrade, just check if one is available
var checkOnly bool

// NewUpgradeCommand creates the 'self upgrade' command.
// When adding this as a subcommand to another CLI, use:
//
//	cmd.AddCommand(version.NewUpgradeCommand())
func NewUpgradeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade syst CLI to the latest release",
		RunE: func(cmd *cobra.Command, args []string) error {
			return UpgradeSelf(cmd, args, checkOnly)
		},
	}

	// Register flags
	cmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for latest version, don't upgrade if one is found.")

	return cmd
}

// UpgradeSelf is the entrypoint for 'syst self upgrade'.
func UpgradeSelf(cmd *cobra.Command, args []string, checkOnly bool) error {
	info := GetPackageInfo()
	repo := fmt.Sprintf("%s/%s", info.RepoUser, info.RepoName)
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/releases/latest", repo)

	fmt.Fprintln(cmd.ErrOrStderr(), "Checking for latest release...")

	// Fetch latest GitHub release metadata
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

	fmt.Fprintln(cmd.ErrOrStderr(), "Latest version:", release.TagName)

	if checkOnly {
		fmt.Fprintln(cmd.ErrOrStderr(), "✅ A newer version may be available. Use this command without --check to upgrade.")
		return nil
	}

	// Match correct release asset
	normalizedOS := normalizeOS(runtime.GOOS)
	expectedPrefix := fmt.Sprintf("%s-%s", normalizedOS, runtime.GOARCH)

	var assetURL string
	for _, asset := range release.Assets {
		if strings.HasPrefix(asset.Name, expectedPrefix) && strings.HasSuffix(asset.Name, ".zip") {
			assetURL = asset.BrowserDownloadURL
			break
		}
	}

	if assetURL == "" {
		return fmt.Errorf("no suitable release found for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// Download the release zip file
	fmt.Fprintln(cmd.ErrOrStderr(), "Downloading:", assetURL)

	resp2, err := http.Get(assetURL)
	if err != nil {
		return fmt.Errorf("failed to download binary: %w", err)
	}
	defer resp2.Body.Close()

	zipTmp, err := os.CreateTemp("", "syst-upgrade-*.zip")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(zipTmp.Name())

	if _, err := io.Copy(zipTmp, resp2.Body); err != nil {
		return fmt.Errorf("failed to write zip file: %w", err)
	}
	zipTmp.Close()

	// Extract the binary
	binaryTmp, err := extractBinaryFromZip(zipTmp.Name())
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}
	defer os.Remove(binaryTmp)

	// Copy binary to current executable’s .new file. On next execution, bin will detect
	//   the .new version and replace it on-the-fly.
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to locate current executable: %w", err)
	}

	newPath := exePath + ".new"
	if err := copyFile(binaryTmp, newPath); err != nil {
		return fmt.Errorf("failed to save new binary: %w", err)
	}

	fmt.Fprintf(cmd.ErrOrStderr(),
		"✅ Upgrade file written to:\n  %s\n"+
			"  The upgrade will apply automatically the next time you run:\n  %s\n",
		newPath,
		filepath.Base(exePath))

	return nil
}

// normalizeOS maps runtime.GOOS to your release asset naming
func normalizeOS(goos string) string {
	if goos == "darwin" {
		return "macOS"
	}

	return goos
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
		return
	}

	newPath := exePath + ".new"

	if _, err := os.Stat(newPath); err == nil {
		// New file exists: perform replacement
		if err := os.Rename(newPath, exePath); err == nil {
			fmt.Fprintf(os.Stderr, "🔁 syst upgraded successfully.\n")
		}
	}
}
