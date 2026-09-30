package module

import (
	"strings"

	azurecontainerappjobv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerappjob/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureContainerAppJob *azurecontainerappjobv1alpha1.AzureContainerAppJob
	ResourceGroupName    string
	AzureTags            map[string]string
}

// probeTransportStrings maps the probe-transport enum to Azure's
// case-sensitive wire values.
var probeTransportStrings = map[azurecontainerappjobv1alpha1.AzureContainerAppJobProbeTransport]string{
	azurecontainerappjobv1alpha1.AzureContainerAppJobProbeTransport_TCP_SOCKET: "TCP",
	azurecontainerappjobv1alpha1.AzureContainerAppJobProbeTransport_HTTP_GET:   "HTTP",
	azurecontainerappjobv1alpha1.AzureContainerAppJobProbeTransport_HTTPS_GET:  "HTTPS",
}

// volumeStorageTypeStrings maps the volume storage-type enum to ARM's
// values. Unspecified deploys EmptyDir (resolved in volume construction).
var volumeStorageTypeStrings = map[azurecontainerappjobv1alpha1.AzureContainerAppJobVolumeStorageType]string{
	azurecontainerappjobv1alpha1.AzureContainerAppJobVolumeStorageType_EMPTY_DIR:      "EmptyDir",
	azurecontainerappjobv1alpha1.AzureContainerAppJobVolumeStorageType_AZURE_FILE:     "AzureFile",
	azurecontainerappjobv1alpha1.AzureContainerAppJobVolumeStorageType_NFS_AZURE_FILE: "NfsAzureFile",
	azurecontainerappjobv1alpha1.AzureContainerAppJobVolumeStorageType_SECRET:         "Secret",
}

// identityTypeStrings maps the identity-type enum to ARM's values.
var identityTypeStrings = map[azurecontainerappjobv1alpha1.AzureContainerAppJobIdentityType]string{
	azurecontainerappjobv1alpha1.AzureContainerAppJobIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurecontainerappjobv1alpha1.AzureContainerAppJobIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurecontainerappjobv1alpha1.AzureContainerAppJobIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurecontainerappjobv1alpha1.AzureContainerAppJobStackInput) *Locals {
	locals := &Locals{}

	locals.AzureContainerAppJob = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureContainerAppJob.String()),
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
