package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/plantonhq/planton/pkg/iac/provisioner"
)

// PromptForProvisioner prompts the user to select a provisioner interactively
// Returns the selected provisioner type, defaulting to Pulumi if user presses Enter
func PromptForProvisioner() (provisioner.ProvisionerType, error) {
	fmt.Print("Select provisioner [Pulumi]/tofu/terraform: ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return provisioner.ProvisionerTypeUnspecified, fmt.Errorf("failed to read input: %w", err)
	}

	// Trim whitespace and newline
	input = strings.TrimSpace(input)

	// Default to Pulumi if empty
	if input == "" {
		return provisioner.ProvisionerTypePulumi, nil
	}

	// Convert string to provisioner type (case-insensitive)
	provType, err := provisioner.FromString(input)
	if err != nil {
		return provisioner.ProvisionerTypeUnspecified, err
	}

	return provType, nil
}
