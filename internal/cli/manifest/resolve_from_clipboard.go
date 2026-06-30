package manifest

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/plantonhq/planton/internal/cli/iacflags"
	"github.com/plantonhq/planton/internal/cli/workspace"
	"github.com/plantonhq/planton/pkg/clipboard"
	"github.com/plantonhq/planton/pkg/iac/stackinput"
	"github.com/plantonhq/planton/pkg/ulidgen"
	"github.com/spf13/cobra"
)

// resolveFromClipboard checks for clipboard flags (--clipboard, --clip, --cb, -c) and reads content from clipboard.
// Returns empty string if flag not provided.
//
// Smart detection order:
//  1. If clipboard content is a file path (and file exists), read from that file
//  2. If clipboard content has a "target" field at root, treat as stack input
//  3. Otherwise, treat as raw manifest YAML
//
// Returns structured errors (ClipboardEmptyError, ClipboardInvalidYAMLError, ClipboardFileNotFoundError)
// that can be handled by command handlers for beautiful display.
func resolveFromClipboard(cmd *cobra.Command) (manifestPath string, isTemp bool, err error) {
	useClipboard, err := iacflags.IsClipboardFlagSet(cmd)
	if err != nil {
		return "", false, errors.Wrap(err, "failed to get clipboard flag")
	}
	if !useClipboard {
		return "", false, nil
	}

	raw, err := clipboard.Read()
	if err != nil {
		// Check if clipboard is empty
		if strings.Contains(err.Error(), "empty") {
			return "", false, &ClipboardEmptyError{}
		}
		return "", false, err
	}

	// Parse and validate clipboard content
	content := ParseClipboardContent(raw)

	// Case 1: File path detected
	if content.IsFilePath {
		if !content.FileExists {
			return "", false, &ClipboardFileNotFoundError{FilePath: content.FilePath}
		}
		// Return the file path directly (not a temp file)
		return content.FilePath, false, nil
	}

	// Case 2: Invalid YAML
	if !content.IsValidYAML {
		return "", false, &ClipboardInvalidYAMLError{
			Raw:        raw,
			ParseError: content.ParseError,
		}
	}

	// Case 3: Stack input (has "target" field)
	if content.IsStackInput {
		manifestPath, err = stackinput.ExtractManifestFromBytes(raw)
		if err != nil {
			return "", false, errors.Wrap(err, "failed to extract manifest from stack input in clipboard")
		}
		return manifestPath, true, nil
	}

	// Case 4: Raw manifest YAML
	manifestPath, err = writeClipboardContent(raw)
	if err != nil {
		return "", false, err
	}

	return manifestPath, true, nil
}

// writeClipboardContent writes content to a file in the downloads directory.
// Follows the same pattern as extract_manifest.go for consistent temp file handling.
func writeClipboardContent(content []byte) (string, error) {
	downloadDir, err := workspace.GetManifestDownloadDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to get manifest download directory")
	}

	fileName := ulidgen.NewGenerator().Generate().String() + "-clipboard-manifest.yaml"
	manifestPath := filepath.Join(downloadDir, fileName)

	if err := os.WriteFile(manifestPath, content, 0600); err != nil {
		return "", errors.Wrapf(err, "failed to write manifest to %s", manifestPath)
	}

	return manifestPath, nil
}
