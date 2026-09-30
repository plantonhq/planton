package module

import (
	"strings"

	azurecontainerappenvironmentv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappenvironment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureContainerAppEnvironment *azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironment
	ResourceGroupName            string
	AzureTags                    map[string]string
}

// logsDestinationStrings maps the logs-destination enum to Azure's wire
// values. Unspecified is handled in main.go (workspace present implies
// log-analytics; otherwise the property is omitted for streaming-only).
var logsDestinationStrings = map[azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentLogsDestination]string{
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentLogsDestination_LOG_ANALYTICS: "log-analytics",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentLogsDestination_AZURE_MONITOR: "azure-monitor",
}

// publicNetworkAccessStrings maps the public-network-access enum to ARM's
// values. Unspecified is never sent -- Azure derives the value from the
// network configuration.
var publicNetworkAccessStrings = map[azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentPublicNetworkAccess]string{
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentPublicNetworkAccess_ENABLED:  "Enabled",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentPublicNetworkAccess_DISABLED: "Disabled",
}

// workloadProfileTypeStrings maps the profile-type enum to Azure's SKU
// spellings. Spelled out row by row so a vocabulary drift fails loudly at
// preview time instead of deploying a wrong profile.
var workloadProfileTypeStrings = map[azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType]string{
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_CONSUMPTION:               "Consumption",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_CONSUMPTION_GPU_NC8AS_T4:  "Consumption-GPU-NC8as-T4",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_CONSUMPTION_GPU_NC24_A100: "Consumption-GPU-NC24-A100",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_D4:                        "D4",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_D8:                        "D8",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_D16:                       "D16",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_D32:                       "D32",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_E4:                        "E4",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_E8:                        "E8",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_E16:                       "E16",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_E32:                       "E32",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_NC24_A100:                 "NC24-A100",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_NC48_A100:                 "NC48-A100",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentWorkloadProfileType_NC96_A100:                 "NC96-A100",
}

// identityTypeStrings maps the identity-type enum to ARM's values.
var identityTypeStrings = map[azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentIdentityType]string{
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurecontainerappenvironmentv1alpha1.AzureContainerAppEnvironmentStackInput) *Locals {
	locals := &Locals{}

	locals.AzureContainerAppEnvironment = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureContainerAppEnvironment.String()),
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

	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
