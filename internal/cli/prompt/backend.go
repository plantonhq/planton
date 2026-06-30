package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/plantonhq/planton/internal/cli/ui"
	"github.com/plantonhq/planton/pkg/iac/tofu/backendconfig"
)

// PromptForMissingBackendConfig interactively prompts the user for missing required fields.
// It displays helpful context for each field and applies the user's input to the config.
// Returns the updated configuration or an error if the user doesn't provide a required value.
func PromptForMissingBackendConfig(
	config *backendconfig.TofuBackendConfig,
	missing []backendconfig.MissingField,
) (*backendconfig.TofuBackendConfig, error) {
	reader := bufio.NewReader(os.Stdin)

	for _, field := range missing {
		if !field.Required {
			continue
		}

		// Display field context
		ui.MissingFieldPrompt(field)

		// Prompt for input
		fmt.Printf("Enter %s: ", field.Name)

		input, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read input for %s: %w", field.Name, err)
		}

		value := strings.TrimSpace(input)
		if value == "" {
			return nil, fmt.Errorf("%s is required but was not provided", field.Name)
		}

		// Apply value to config
		applyFieldValue(config, field.Name, value)
	}

	// Recompute S3-compatible flag after prompts
	config.S3Compatible = config.IsS3Compatible()

	return config, nil
}

// applyFieldValue applies a user-provided value to the appropriate config field.
func applyFieldValue(config *backendconfig.TofuBackendConfig, fieldName, value string) {
	switch fieldName {
	case "bucket":
		config.BackendBucket = value
	case "key":
		config.BackendKey = value
	case "region":
		config.BackendRegion = value
	case "endpoint":
		config.BackendEndpoint = value
	case "type":
		config.BackendType = value
	}
}

// IsInteractive returns true if stdin is a terminal (interactive mode).
// In non-interactive mode (CI/CD), prompts should not be displayed.
func IsInteractive() bool {
	fileInfo, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	// Check if stdin is a character device (terminal)
	return (fileInfo.Mode() & os.ModeCharDevice) != 0
}
