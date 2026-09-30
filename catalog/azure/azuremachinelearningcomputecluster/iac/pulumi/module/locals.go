package module

import (
	"strings"

	azuremachinelearningcomputeclusterv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremachinelearningcomputecluster/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMachineLearningComputeCluster *azuremachinelearningcomputeclusterv1alpha1.AzureMachineLearningComputeCluster

	// WorkspaceId is a StringValueOrRef field; the platform middleware
	// resolves valueFrom references before IaC modules run, so GetValue()
	// always returns the resolved literal ARM ID.
	WorkspaceId string

	// AzureTags is the metadata-derived tag map with the spec's user tags
	// merged over it (user tags win on key collision), mirroring the
	// Terraform module's merge order.
	AzureTags map[string]string
}

// vmPriorityWire maps the spec's VM priorities to the provider's wire
// values. Required on the provider -- the spec enum is required, so
// every deployable spec has an entry here.
var vmPriorityWire = map[azuremachinelearningcomputeclusterv1alpha1.AzureMachineLearningComputeClusterVmPriority]string{
	azuremachinelearningcomputeclusterv1alpha1.AzureMachineLearningComputeClusterVmPriority_DEDICATED:    "Dedicated",
	azuremachinelearningcomputeclusterv1alpha1.AzureMachineLearningComputeClusterVmPriority_LOW_PRIORITY: "LowPriority",
}

// identityTypeWire maps the spec's identity flavors to the provider's
// comma-joined wire values.
var identityTypeWire = map[azuremachinelearningcomputeclusterv1alpha1.AzureMachineLearningComputeClusterIdentityType]string{
	azuremachinelearningcomputeclusterv1alpha1.AzureMachineLearningComputeClusterIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azuremachinelearningcomputeclusterv1alpha1.AzureMachineLearningComputeClusterIdentityType_USER_ASSIGNED:            "UserAssigned",
	azuremachinelearningcomputeclusterv1alpha1.AzureMachineLearningComputeClusterIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azuremachinelearningcomputeclusterv1alpha1.AzureMachineLearningComputeClusterStackInput) *Locals {
	locals := &Locals{}

	locals.AzureMachineLearningComputeCluster = stackInput.Target
	target := stackInput.Target

	locals.WorkspaceId = target.Spec.WorkspaceId.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureMachineLearningComputeCluster.String()),
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
