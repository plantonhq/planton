package prompt

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/plantonhq/planton/pkg/iac/provisioner"
)

// PromptForProvisioner asks the user to choose among the engines a kind runs on, defaulting to
// the first when they press Enter. With no engines given it offers all three, Pulumi first, as
// the CLI always has. An answer outside the offered engines is refused, so a choice the kind
// cannot run is never accepted here to fail later.
func PromptForProvisioner(allowed ...provisioner.ProvisionerType) (provisioner.ProvisionerType, error) {
	if len(allowed) == 0 {
		allowed = []provisioner.ProvisionerType{
			provisioner.ProvisionerTypePulumi,
			provisioner.ProvisionerTypeTofu,
			provisioner.ProvisionerTypeTerraform,
		}
	}
	choices := make([]string, 0, len(allowed))
	for i, p := range allowed {
		if i == 0 {
			choices = append(choices, "["+p.DisplayName()+"]")
			continue
		}
		choices = append(choices, p.String())
	}
	fmt.Printf("Select provisioner %s: ", strings.Join(choices, "/"))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return provisioner.ProvisionerTypeUnspecified, fmt.Errorf("failed to read input: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		return allowed[0], nil
	}

	provType, err := provisioner.FromString(input)
	if err != nil {
		return provisioner.ProvisionerTypeUnspecified, err
	}
	for _, p := range allowed {
		if p == provType {
			return provType, nil
		}
	}
	offered := make([]string, 0, len(allowed))
	for _, p := range allowed {
		offered = append(offered, p.String())
	}
	return provisioner.ProvisionerTypeUnspecified,
		fmt.Errorf("%s cannot deploy this kind: choose %s", provType.DisplayName(), strings.Join(offered, " or "))
}
