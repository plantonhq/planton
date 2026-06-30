package manifest

import (
	"fmt"

	"github.com/plantonhq/planton/internal/cli/ui"
)

// ClipboardEmptyError indicates clipboard has no content.
type ClipboardEmptyError struct{}

func (e *ClipboardEmptyError) Error() string {
	return "clipboard is empty"
}

// ClipboardInvalidYAMLError indicates clipboard content is not valid YAML.
type ClipboardInvalidYAMLError struct {
	Raw        []byte
	ParseError error
}

func (e *ClipboardInvalidYAMLError) Error() string {
	return fmt.Sprintf("clipboard content is not valid YAML: %v", e.ParseError)
}

// ClipboardFileNotFoundError indicates clipboard contains a file path but file doesn't exist.
type ClipboardFileNotFoundError struct {
	FilePath string
}

func (e *ClipboardFileNotFoundError) Error() string {
	return fmt.Sprintf("file not found: %s", e.FilePath)
}

// ClipboardNotStackInputError indicates clipboard content is valid YAML
// but not a stack input (missing 'target' field).
type ClipboardNotStackInputError struct {
	Raw []byte
}

func (e *ClipboardNotStackInputError) Error() string {
	return "clipboard content is not a stack input (missing 'target' field)"
}

// IsClipboardError returns true if the error is a clipboard-related error.
func IsClipboardError(err error) bool {
	switch err.(type) {
	case *ClipboardEmptyError, *ClipboardInvalidYAMLError, *ClipboardFileNotFoundError, *ClipboardNotStackInputError:
		return true
	default:
		return false
	}
}

// HandleClipboardError checks if the error is a clipboard error and displays
// a beautiful error message. Returns true if it was a clipboard error (and was handled),
// false otherwise.
func HandleClipboardError(err error) bool {
	if err == nil {
		return false
	}

	switch e := err.(type) {
	case *ClipboardEmptyError:
		ui.ClipboardEmpty()
		return true
	case *ClipboardInvalidYAMLError:
		ui.ClipboardInvalidYAML(e.Raw, e.ParseError)
		return true
	case *ClipboardFileNotFoundError:
		ui.ClipboardFileNotFound(e.FilePath)
		return true
	case *ClipboardNotStackInputError:
		ui.ClipboardNotStackInput(e.Raw)
		return true
	default:
		return false
	}
}
