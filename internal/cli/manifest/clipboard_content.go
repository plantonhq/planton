package manifest

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/plantonhq/planton/pkg/iac/stackinput"
	"gopkg.in/yaml.v3"
)

// ClipboardContent holds parsed clipboard data and validation results.
type ClipboardContent struct {
	Raw          []byte
	IsFilePath   bool
	FilePath     string
	FileExists   bool
	IsValidYAML  bool
	ParseError   error
	IsStackInput bool
}

// ParseClipboardContent analyzes clipboard content and determines its type.
// It checks if the content is a file path, valid YAML, or stack input.
func ParseClipboardContent(raw []byte) *ClipboardContent {
	content := &ClipboardContent{
		Raw: raw,
	}

	// Trim whitespace for analysis
	trimmed := strings.TrimSpace(string(raw))

	// Check if it looks like a file path
	if isLikelyFilePath(trimmed) {
		content.IsFilePath = true
		content.FilePath = expandPath(trimmed)
		content.FileExists = fileExists(content.FilePath)
		return content
	}

	// Try to parse as YAML
	var parsed map[string]interface{}
	if err := yaml.Unmarshal(raw, &parsed); err != nil {
		content.IsValidYAML = false
		content.ParseError = err
		return content
	}

	content.IsValidYAML = true

	// Check if it's a stack input (has "target" key)
	content.IsStackInput = stackinput.IsStackInput(raw)

	return content
}

// isLikelyFilePath checks if content looks like a file path rather than YAML.
// For a path to be considered a file path, it must:
// - Be a single line
// - Look like a path (starts with /, ./, ../, or ~/)
// - Have a .yaml or .yml extension (required for paths that don't exist)
func isLikelyFilePath(content string) bool {
	// Must be a single line
	if strings.Contains(content, "\n") {
		return false
	}

	// Must not contain spaces (file paths typically don't)
	if strings.Contains(content, " ") {
		return false
	}

	// Check if it has YAML extension
	hasYAMLExtension := strings.HasSuffix(strings.ToLower(content), ".yaml") ||
		strings.HasSuffix(strings.ToLower(content), ".yml")

	// Check for common path prefixes
	isPathLike := strings.HasPrefix(content, "/") ||
		strings.HasPrefix(content, "./") ||
		strings.HasPrefix(content, "../") ||
		strings.HasPrefix(content, "~/")

	if isPathLike {
		// For path-like content, require YAML extension OR the file must exist
		if hasYAMLExtension {
			return true
		}
		// Check if the file exists (even without .yaml extension)
		expanded := expandPath(content)
		if fileExists(expanded) {
			return true
		}
		// Path-like but doesn't have .yaml extension and file doesn't exist
		// Treat as potential YAML content instead
		return false
	}

	// For content that doesn't look like a path, only treat as path if:
	// - It has .yaml/.yml extension AND doesn't look like YAML
	if hasYAMLExtension && !strings.Contains(content, ":") {
		return true
	}

	return false
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}
