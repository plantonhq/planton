package module

import (
	"strings"

	gcpprovider "github.com/plantonhq/planton/catalog/gcp"
	gcpdeploytargetv1alpha1 "github.com/plantonhq/planton/catalog/gcp/gcpdeploytarget/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/gcp/gcplabelkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	GcpProviderConfig *gcpprovider.GcpProviderConfig
	GcpDeployTarget   *gcpdeploytargetv1alpha1.GcpDeployTarget

	// GcpLabels is the spec's labels with the planton-ai_* attribution set
	// merged on top.
	GcpLabels map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *gcpdeploytargetv1alpha1.GcpDeployTargetIacInput) *Locals {
	locals := &Locals{}
	locals.GcpDeployTarget = iacInput.Target
	metadata := locals.GcpDeployTarget.Metadata

	attribution := map[string]string{
		gcplabelkeys.Resource:     "true",
		gcplabelkeys.ResourceName: metadata.Name,
		gcplabelkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_GcpDeployTarget.String()),
	}
	if metadata.Org != "" {
		attribution[gcplabelkeys.Organization] = metadata.Org
	}
	if metadata.Env != "" {
		attribution[gcplabelkeys.Environment] = metadata.Env
	}
	if metadata.Id != "" {
		attribution[gcplabelkeys.ResourceId] = metadata.Id
	}

	// User labels first so platform attribution labels win on key
	// conflicts -- identical merge order to the Terraform module.
	locals.GcpLabels = mergeLabels(locals.GcpDeployTarget.Spec.Labels, attribution)

	locals.GcpProviderConfig = iacInput.ProviderConfig
	return locals
}

// mergeLabels returns user labels with the attribution labels laid over
// them, the Terraform module's merge(user, attribution).
func mergeLabels(user, attribution map[string]string) map[string]string {
	merged := map[string]string{}
	for key, value := range user {
		merged[key] = value
	}
	for key, value := range attribution {
		merged[key] = value
	}
	return merged
}
