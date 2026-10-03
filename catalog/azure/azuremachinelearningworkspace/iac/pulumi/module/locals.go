package module

import (
	"strings"

	azuremachinelearningworkspacev1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningworkspace/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMachineLearningWorkspace *azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspace

	// ResourceGroupName is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal name.
	ResourceGroupName string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

// identityTypeWire maps the spec's identity flavors to the provider's
// comma-joined wire values.
var identityTypeWire = map[azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceIdentityType]string{
	azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceIdentityType_USER_ASSIGNED:            "UserAssigned",
	azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

// kindWire maps the spec's workspace flavors to the provider's wire
// values. Unspecified is absent -- the property is omitted so the
// provider applies its default, "Default".
var kindWire = map[azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceKind]string{
	azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceKind_DEFAULT:       "Default",
	azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceKind_FEATURE_STORE: "FeatureStore",
}

// isolationModeWire maps the spec's isolation modes to the provider's
// wire values. Unspecified is absent -- the property is omitted and the
// value is read back.
var isolationModeWire = map[azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceIsolationMode]string{
	azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceIsolationMode_DISABLED:                     "Disabled",
	azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceIsolationMode_ALLOW_INTERNET_OUTBOUND:      "AllowInternetOutbound",
	azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceIsolationMode_ALLOW_ONLY_APPROVED_OUTBOUND: "AllowOnlyApprovedOutbound",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azuremachinelearningworkspacev1alpha1.AzureMachineLearningWorkspaceIacInput) *Locals {
	locals := &Locals{}

	locals.AzureMachineLearningWorkspace = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureMachineLearningWorkspace.String()),
	}

	if target.Metadata.Id != "" {
		locals.AzureTags[azuretagkeys.ResourceId] = target.Metadata.Id
	}

	if target.Metadata.Org != "" {
		locals.AzureTags[azuretagkeys.Organization] = target.Metadata.Org
	}

	if target.Metadata.Env != "" {
		locals.AzureTags[azuretagkeys.Environment] = target.Metadata.Env
	}

	for k, v := range target.Spec.Tags {
		locals.AzureTags[k] = v
	}

	return locals
}
