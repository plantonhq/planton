package pulumibinary

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/cliprint"
	"github.com/plantonhq/planton/internal/cli/version"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/pkg/downloads"
	"github.com/plantonhq/planton/pkg/fileutil"
)

const (
	// PulumiDirName is the base directory name for all Pulumi-related files
	// All Pulumi files are stored under ~/.planton/pulumi/
	PulumiDirName = "pulumi"

	// BinariesSubDir is the subdirectory for cached binaries
	// Full path: ~/.planton/pulumi/binaries/{version}/
	BinariesSubDir = "binaries"

	// WorkspacesSubDir is the subdirectory for Pulumi workspaces
	// Full path: ~/.planton/pulumi/workspaces/{stack-fqdn}/
	WorkspacesSubDir = "workspaces"

	// BinaryPrefix is the prefix for locally cached Pulumi component binaries.
	BinaryPrefix = "pulumi-"
)

// GetPlatformSuffix returns the platform suffix for binary names based on the current OS and architecture.
// Examples: "linux_amd64", "darwin_arm64", "darwin_amd64", "windows_amd64"
func GetPlatformSuffix() string {
	return fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
}

// GetPulumiBaseDir returns the base directory for all Pulumi-related files
// (~/.planton/pulumi/)
func GetPulumiBaseDir() (string, error) {
	workspaceDir, err := workspace.GetWorkspaceDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get workspace directory")
	}
	return filepath.Join(workspaceDir, PulumiDirName), nil
}

// GetBinaryCacheDir returns the path to the binary cache directory
// (~/.planton/pulumi/binaries/{version}/)
func GetBinaryCacheDir(releaseVersion string) (string, error) {
	pulumiBaseDir, err := GetPulumiBaseDir()
	if err != nil {
		return "", err
	}

	// Normalize version for directory name
	versionDir := releaseVersion
	if versionDir == "" || versionDir == version.DefaultVersion {
		versionDir = "dev"
	}

	return filepath.Join(pulumiBaseDir, BinariesSubDir, versionDir), nil
}

// GetBinaryPath returns the expected path for a cached binary
func GetBinaryPath(componentName, releaseVersion string) (string, error) {
	cacheDir, err := GetBinaryCacheDir(releaseVersion)
	if err != nil {
		return "", err
	}

	binaryName := BuildBinaryName(componentName)
	return filepath.Join(cacheDir, binaryName), nil
}

// BuildBinaryName constructs the platform-specific binary filename for a component.
// The binary name includes the platform suffix based on the current OS and architecture.
// Examples:
//   - Linux:   "AwsEcsService" -> "pulumi-awsecsservice_linux_amd64"
//   - macOS:   "AwsEcsService" -> "pulumi-awsecsservice_darwin_arm64"
//   - Windows: "AwsEcsService" -> "pulumi-awsecsservice_windows_amd64.exe"
func BuildBinaryName(componentName string) string {
	baseName := BinaryPrefix + strings.ToLower(componentName)
	suffix := GetPlatformSuffix()

	if runtime.GOOS == "windows" {
		return fmt.Sprintf("%s_%s.exe", baseName, suffix)
	}
	return fmt.Sprintf("%s_%s", baseName, suffix)
}

// BuildDownloadURL constructs the Cloudflare R2 download URL for a platform-specific component binary.
//
// Examples (on darwin/arm64):
//
//	BuildDownloadURL("AwsEcsService", "v0.3.50")
//	  -> https://downloads.planton.dev/releases/v0.3.50/modules/pulumi/awsecsservice_darwin_arm64.gz
func BuildDownloadURL(componentName, releaseVersion string) string {
	return downloads.BuildPulumiDownloadURL(componentName, releaseVersion, GetPlatformSuffix())
}

// IsBinaryCached checks if a binary is already cached
func IsBinaryCached(componentName, releaseVersion string) (bool, error) {
	binaryPath, err := GetBinaryPath(componentName, releaseVersion)
	if err != nil {
		return false, err
	}

	exists, err := fileutil.IsExists(binaryPath)
	if err != nil {
		return false, errors.Wrapf(err, "failed to check if binary exists at %s", binaryPath)
	}

	if !exists {
		return false, nil
	}

	// Verify it's executable
	info, err := os.Stat(binaryPath)
	if err != nil {
		return false, errors.Wrapf(err, "failed to stat binary at %s", binaryPath)
	}

	// Check if file has execute permission
	return info.Mode()&0111 != 0, nil
}

// EnsureBinary ensures the binary for a component is downloaded and cached.
// The releaseVersion can be:
// - CLI version like "v0.3.2" (uses main planton release)
// - Module version like "v0.3.1-pulumi-awsecsservice-20260107.01" (uses component-specific release)
// Returns the path to the binary.
func EnsureBinary(componentName, releaseVersion string) (string, error) {
	// Check if already cached
	cached, err := IsBinaryCached(componentName, releaseVersion)
	if err != nil {
		return "", errors.Wrap(err, "failed to check binary cache")
	}

	binaryPath, err := GetBinaryPath(componentName, releaseVersion)
	if err != nil {
		return "", err
	}

	if cached {
		cliprint.PrintSuccess(fmt.Sprintf("Using cached module: %s", componentName))
		return binaryPath, nil
	}

	// Download the binary
	cliprint.PrintStep(fmt.Sprintf("Downloading Pulumi module: %s...", componentName))

	if err := DownloadBinary(componentName, releaseVersion); err != nil {
		return "", errors.Wrapf(err, "failed to download binary for %s", componentName)
	}

	cliprint.PrintSuccess(fmt.Sprintf("Module downloaded: %s", componentName))
	return binaryPath, nil
}

// DownloadBinary downloads and extracts a platform-specific component binary from Cloudflare R2.
// The binary downloaded is platform-specific based on runtime.GOOS and runtime.GOARCH.
func DownloadBinary(componentName, releaseVersion string) error {
	// Ensure cache directory exists
	cacheDir, err := GetBinaryCacheDir(releaseVersion)
	if err != nil {
		return err
	}

	if !fileutil.IsDirExists(cacheDir) {
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return errors.Wrapf(err, "failed to create cache directory %s", cacheDir)
		}
	}

	// Build download URL - the release version IS the tag
	downloadURL := BuildDownloadURL(componentName, releaseVersion)

	cliprint.PrintInfo(fmt.Sprintf("Downloading from: %s", downloadURL))

	// Download the gzipped binary
	resp, err := http.Get(downloadURL)
	if err != nil {
		return errors.Wrapf(err, "failed to download from %s", downloadURL)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("download failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	// Create gzip reader
	gzReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return errors.Wrap(err, "failed to create gzip reader")
	}
	defer gzReader.Close()

	// Write to destination
	binaryPath, err := GetBinaryPath(componentName, releaseVersion)
	if err != nil {
		return err
	}

	outFile, err := os.OpenFile(binaryPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return errors.Wrapf(err, "failed to create binary file at %s", binaryPath)
	}
	defer outFile.Close()

	written, err := io.Copy(outFile, gzReader)
	if err != nil {
		// Clean up partial file
		os.Remove(binaryPath)
		return errors.Wrap(err, "failed to extract binary")
	}

	cliprint.PrintInfo(fmt.Sprintf("Extracted %d bytes", written))

	return nil
}

// GetCurrentCLIVersion returns the current CLI version, falling back to "dev" if not set
func GetCurrentCLIVersion() string {
	if version.Version == "" || version.Version == version.DefaultVersion {
		return "dev"
	}
	return version.Version
}

// IsDevVersion checks if the current CLI is a development version
func IsDevVersion() bool {
	return version.Version == "" || version.Version == version.DefaultVersion
}
