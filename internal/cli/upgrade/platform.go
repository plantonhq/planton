package upgrade

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// UpgradeMethod represents the method used to upgrade the CLI
type UpgradeMethod int

const (
	// MethodHomebrew uses Homebrew to upgrade (macOS only)
	MethodHomebrew UpgradeMethod = iota
	// MethodDirectDownload downloads the binary directly from GitHub
	MethodDirectDownload
)

// String returns a human-readable name for the upgrade method
func (m UpgradeMethod) String() string {
	switch m {
	case MethodHomebrew:
		return "Homebrew"
	case MethodDirectDownload:
		return "Direct Download"
	default:
		return "Unknown"
	}
}

// DetectUpgradeMethod determines the best method to upgrade the CLI
// based on the current platform and installation method
func DetectUpgradeMethod() UpgradeMethod {
	// Only macOS uses Homebrew
	if runtime.GOOS != "darwin" {
		return MethodDirectDownload
	}

	// Check if brew is available
	if _, err := exec.LookPath("brew"); err != nil {
		return MethodDirectDownload
	}

	// Check if planton was installed via Homebrew cask
	cmd := exec.Command("brew", "list", "--cask", "planton")
	if err := cmd.Run(); err != nil {
		return MethodDirectDownload
	}

	return MethodHomebrew
}

// GetPlatformInfo returns the current OS and architecture
func GetPlatformInfo() (goos string, goarch string) {
	return runtime.GOOS, runtime.GOARCH
}

// BuildDownloadURL constructs the download URL for a specific version and platform
// Uses the GitHub releases asset URL pattern from GoReleaser config
func BuildDownloadURL(version, goos, goarch string) string {
	// Strip 'v' prefix from version for the archive filename
	versionNum := strings.TrimPrefix(version, "v")

	// Build archive name based on GoReleaser config:
	// name_template: "cli_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
	var archiveName string
	if goos == "windows" {
		archiveName = fmt.Sprintf("cli_%s_%s_%s.zip", versionNum, goos, goarch)
	} else {
		archiveName = fmt.Sprintf("cli_%s_%s_%s.tar.gz", versionNum, goos, goarch)
	}

	return fmt.Sprintf("https://github.com/plantonhq/planton/releases/download/%s/%s", version, archiveName)
}

// BuildChecksumURL constructs the checksum file URL for a specific version
// Uses the GoReleaser checksum template: "cli_{{ .Version }}_checksums.txt"
func BuildChecksumURL(version string) string {
	// Strip 'v' prefix from version for the checksum filename
	versionNum := strings.TrimPrefix(version, "v")
	checksumFile := fmt.Sprintf("cli_%s_checksums.txt", versionNum)
	return fmt.Sprintf("https://github.com/plantonhq/planton/releases/download/%s/%s", version, checksumFile)
}
