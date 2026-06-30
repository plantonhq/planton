package provisioner

import (
	"fmt"
	"strings"

	"github.com/plantonhq/planton/pkg/iac/provisionerlabels"
	"github.com/plantonhq/planton/pkg/reflection/metadatareflect"
	"google.golang.org/protobuf/proto"
)

// ProvisionerType represents the IaC provisioner type
type ProvisionerType int

const (
	ProvisionerTypeUnspecified ProvisionerType = iota
	ProvisionerTypePulumi
	ProvisionerTypeTofu
	ProvisionerTypeTerraform
)

// String returns the string representation of the provisioner type
func (p ProvisionerType) String() string {
	switch p {
	case ProvisionerTypePulumi:
		return "pulumi"
	case ProvisionerTypeTofu:
		return "tofu"
	case ProvisionerTypeTerraform:
		return "terraform"
	default:
		return "unspecified"
	}
}

// ExtractFromManifest extracts the provisioner type from manifest labels
// Returns:
//   - ProvisionerType and nil error if label exists and is valid
//   - ProvisionerTypeUnspecified and nil error if label is missing (needs user prompt)
//   - ProvisionerTypeUnspecified and error if label value is invalid
func ExtractFromManifest(manifest proto.Message) (ProvisionerType, error) {
	labels := metadatareflect.ExtractLabels(manifest)
	if labels == nil {
		return ProvisionerTypeUnspecified, nil
	}

	provisioner, ok := labels[provisionerlabels.ProvisionerLabelKey]
	if !ok || provisioner == "" {
		// Label not present - return unspecified (caller should prompt user)
		return ProvisionerTypeUnspecified, nil
	}

	// Case-insensitive matching
	provisionerLower := strings.ToLower(strings.TrimSpace(provisioner))

	switch provisionerLower {
	case "pulumi":
		return ProvisionerTypePulumi, nil
	case "tofu":
		return ProvisionerTypeTofu, nil
	case "terraform":
		return ProvisionerTypeTerraform, nil
	default:
		return ProvisionerTypeUnspecified, fmt.Errorf("invalid provisioner value '%s': must be one of 'pulumi', 'tofu', or 'terraform'", provisioner)
	}
}

// FromString converts a string to ProvisionerType (case-insensitive)
func FromString(s string) (ProvisionerType, error) {
	sLower := strings.ToLower(strings.TrimSpace(s))
	switch sLower {
	case "pulumi":
		return ProvisionerTypePulumi, nil
	case "tofu":
		return ProvisionerTypeTofu, nil
	case "terraform":
		return ProvisionerTypeTerraform, nil
	default:
		return ProvisionerTypeUnspecified, fmt.Errorf("invalid provisioner '%s': must be one of 'pulumi', 'tofu', or 'terraform'", s)
	}
}
