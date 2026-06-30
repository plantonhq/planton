package upgrade

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/plantonhq/planton/internal/cli/cliprint"
)

// UpgradeViaDirect upgrades the CLI by downloading the binary directly from GitHub
func UpgradeViaDirect(version string) error {
	goos, goarch := GetPlatformInfo()
	downloadURL := BuildDownloadURL(version, goos, goarch)
	checksumURL := BuildChecksumURL(version)

	dim := color.New(color.Faint).SprintFunc()

	// Step 1: Download archive to temp file
	fmt.Println()
	cliprint.PrintStep(fmt.Sprintf("Downloading planton %s...", version))
	fmt.Printf("  %s\n", dim(downloadURL))

	tempArchive, err := downloadToTemp(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}
	defer os.Remove(tempArchive)

	cliprint.PrintSuccess(fmt.Sprintf("Downloaded planton %s", version))

	// Step 2: Verify checksum
	cliprint.PrintStep("Verifying checksum...")

	if err := verifyChecksum(tempArchive, checksumURL, version, goos, goarch); err != nil {
		return fmt.Errorf("checksum verification failed: %w", err)
	}
	cliprint.PrintSuccess("Checksum verified")

	// Step 3: Extract binary from archive
	cliprint.PrintStep("Extracting binary...")

	tempDir, err := os.MkdirTemp("", "planton-upgrade-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	binaryPath, err := extractBinary(tempArchive, tempDir, goos)
	if err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}
	cliprint.PrintSuccess("Extracted binary")

	// Step 4: Determine installation path
	installPath, pathWarning := getInstallPath(goos)

	// Step 5: Ensure install directory exists
	installDir := filepath.Dir(installPath)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return fmt.Errorf("failed to create install directory %s: %w", installDir, err)
	}

	// Step 6: Install binary
	cliprint.PrintStep("Installing...")
	fmt.Printf("  %s\n", dim(installPath))

	if err := replaceBinary(binaryPath, installPath); err != nil {
		return err
	}
	cliprint.PrintSuccess("Installed new binary")

	// Step 7: macOS quarantine removal
	if runtime.GOOS == "darwin" {
		_ = exec.Command("xattr", "-dr", "com.apple.quarantine", installPath).Run()
	}

	// Step 8: Show PATH warning if needed
	if pathWarning != "" {
		fmt.Println()
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("%s %s\n", yellow("⚠"), pathWarning)
	}

	return nil
}

// getInstallPath determines the best installation path for the binary
// Returns the install path and an optional warning message if PATH setup is needed
func getInstallPath(goos string) (string, string) {
	// First, try to use the current binary location if it exists and is writable
	currentBinary, err := os.Executable()
	if err == nil {
		// Resolve symlinks
		resolved, err := filepath.EvalSymlinks(currentBinary)
		if err == nil {
			currentBinary = resolved
		}

		// Check if current binary exists and is writable
		if _, err := os.Stat(currentBinary); err == nil {
			// File exists, check if we can write to it
			if isWritable(currentBinary) {
				return currentBinary, ""
			}
		}
	}

	// Fall back to standard user binary location
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Last resort fallback
		if goos == "windows" {
			return filepath.Join(os.Getenv("LOCALAPPDATA"), "Programs", "planton", "planton.exe"), ""
		}
		return "/usr/local/bin/planton", ""
	}

	if goos == "windows" {
		installPath := filepath.Join(homeDir, "AppData", "Local", "Programs", "planton", "planton.exe")
		return installPath, fmt.Sprintf("Add %s to your PATH if not already configured.", filepath.Dir(installPath))
	}

	// macOS and Linux: use ~/.local/bin (XDG standard)
	installDir := filepath.Join(homeDir, ".local", "bin")
	installPath := filepath.Join(installDir, "planton")

	// Check if ~/.local/bin is in PATH
	pathEnv := os.Getenv("PATH")
	if !strings.Contains(pathEnv, installDir) {
		warning := fmt.Sprintf("Add %s to your PATH:\n  echo 'export PATH=\"$HOME/.local/bin:$PATH\"' >> ~/.zshrc && source ~/.zshrc", installDir)
		if goos == "linux" {
			warning = fmt.Sprintf("Add %s to your PATH:\n  echo 'export PATH=\"$HOME/.local/bin:$PATH\"' >> ~/.bashrc && source ~/.bashrc", installDir)
		}
		return installPath, warning
	}

	return installPath, ""
}

// isWritable checks if a file is writable by attempting to open it for writing
func isWritable(path string) bool {
	file, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// downloadToTemp downloads a file from URL to a temporary file and returns the path
func downloadToTemp(url string) (string, error) {
	// Create temp file with appropriate extension
	var tempFile *os.File
	var err error
	if strings.HasSuffix(url, ".zip") {
		tempFile, err = os.CreateTemp("", "planton-upgrade-*.zip")
	} else {
		tempFile, err = os.CreateTemp("", "planton-upgrade-*.tar.gz")
	}
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Download file
	client := &http.Client{Timeout: httpTimeout * 6} // 60 seconds for download
	resp, err := client.Get(url)
	if err != nil {
		tempFile.Close()
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		tempFile.Close()
		os.Remove(tempPath)
		return "", fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	// Write to temp file
	_, err = io.Copy(tempFile, resp.Body)
	tempFile.Close()
	if err != nil {
		os.Remove(tempPath)
		return "", fmt.Errorf("failed to write downloaded file: %w", err)
	}

	return tempPath, nil
}

// extractBinary extracts the binary from the downloaded archive
func extractBinary(archivePath, destDir, goos string) (string, error) {
	if goos == "windows" {
		return extractFromZip(archivePath, destDir)
	}
	return extractFromTarGz(archivePath, destDir)
}

// extractFromTarGz extracts the binary from a .tar.gz archive
func extractFromTarGz(archivePath, destDir string) (string, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open archive: %w", err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)

	var binaryPath string
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar entry: %w", err)
		}

		// Look for the planton binary
		if header.Typeflag == tar.TypeReg && header.Name == "planton" {
			binaryPath = filepath.Join(destDir, header.Name)
			outFile, err := os.OpenFile(binaryPath, os.O_CREATE|os.O_WRONLY, os.FileMode(header.Mode))
			if err != nil {
				return "", fmt.Errorf("failed to create binary file: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return "", fmt.Errorf("failed to write binary file: %w", err)
			}
			outFile.Close()
			break
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("binary not found in archive")
	}

	return binaryPath, nil
}

// extractFromZip extracts the binary from a .zip archive (Windows)
func extractFromZip(archivePath, destDir string) (string, error) {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return "", fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer reader.Close()

	var binaryPath string
	for _, file := range reader.File {
		// Look for the planton.exe binary
		if file.Name == "planton.exe" {
			binaryPath = filepath.Join(destDir, file.Name)

			rc, err := file.Open()
			if err != nil {
				return "", fmt.Errorf("failed to open file in archive: %w", err)
			}

			outFile, err := os.OpenFile(binaryPath, os.O_CREATE|os.O_WRONLY, file.Mode())
			if err != nil {
				rc.Close()
				return "", fmt.Errorf("failed to create binary file: %w", err)
			}

			if _, err := io.Copy(outFile, rc); err != nil {
				rc.Close()
				outFile.Close()
				return "", fmt.Errorf("failed to write binary file: %w", err)
			}
			rc.Close()
			outFile.Close()
			break
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("binary not found in archive")
	}

	return binaryPath, nil
}

// verifyChecksum downloads the checksum file and verifies the downloaded archive
func verifyChecksum(archivePath, checksumURL, version, goos, goarch string) error {
	// Download checksum file
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(checksumURL)
	if err != nil {
		return fmt.Errorf("failed to download checksums: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch checksums: HTTP %d", resp.StatusCode)
	}

	// Build expected archive name
	versionNum := strings.TrimPrefix(version, "v")
	var archiveName string
	if goos == "windows" {
		archiveName = fmt.Sprintf("cli_%s_%s_%s.zip", versionNum, goos, goarch)
	} else {
		archiveName = fmt.Sprintf("cli_%s_%s_%s.tar.gz", versionNum, goos, goarch)
	}

	// Parse checksum file to find our archive's checksum
	var expectedChecksum string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		// Format: "checksum  filename" or "checksum filename"
		parts := strings.Fields(line)
		if len(parts) >= 2 && parts[1] == archiveName {
			expectedChecksum = parts[0]
			break
		}
	}

	if expectedChecksum == "" {
		return fmt.Errorf("checksum not found for %s", archiveName)
	}

	// Calculate actual checksum of downloaded archive
	file, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("failed to open downloaded archive: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return fmt.Errorf("failed to calculate checksum: %w", err)
	}

	actualChecksum := hex.EncodeToString(hash.Sum(nil))

	if actualChecksum != expectedChecksum {
		return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return nil
}

// replaceBinary replaces the current binary with the new one
func replaceBinary(newBinary, currentBinary string) error {
	// Make new binary executable
	if err := os.Chmod(newBinary, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	// Try to replace the binary directly
	if err := os.Rename(newBinary, currentBinary); err != nil {
		// If rename fails (e.g., cross-device link), try copy
		if err := copyFile(newBinary, currentBinary); err != nil {
			// Check if it's a permission error
			if os.IsPermission(err) {
				return &PermissionError{
					Path:    currentBinary,
					OrigErr: err,
				}
			}
			return fmt.Errorf("failed to install new binary: %w", err)
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// PermissionError represents a permission error when trying to replace the binary
type PermissionError struct {
	Path    string
	OrigErr error
}

func (e *PermissionError) Error() string {
	return fmt.Sprintf("permission denied: cannot write to %s", e.Path)
}
