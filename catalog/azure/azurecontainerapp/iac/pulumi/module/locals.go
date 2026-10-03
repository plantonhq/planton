package module

import (
	"strings"

	azurecontainerappv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurecontainerapp/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureContainerApp *azurecontainerappv1alpha1.AzureContainerApp
	ResourceGroupName string
	AzureTags         map[string]string
}

// revisionModeStrings maps the revision-mode enum to Azure's wire values.
// Unspecified deploys Single (resolved in main.go).
var revisionModeStrings = map[azurecontainerappv1alpha1.AzureContainerAppRevisionMode]string{
	azurecontainerappv1alpha1.AzureContainerAppRevisionMode_SINGLE:   "Single",
	azurecontainerappv1alpha1.AzureContainerAppRevisionMode_MULTIPLE: "Multiple",
}

// probeTransportStrings maps the probe-transport enum to Azure's
// case-sensitive wire values.
var probeTransportStrings = map[azurecontainerappv1alpha1.AzureContainerAppProbeTransport]string{
	azurecontainerappv1alpha1.AzureContainerAppProbeTransport_TCP_SOCKET: "TCP",
	azurecontainerappv1alpha1.AzureContainerAppProbeTransport_HTTP_GET:   "HTTP",
	azurecontainerappv1alpha1.AzureContainerAppProbeTransport_HTTPS_GET:  "HTTPS",
}

// ingressTransportStrings maps the ingress-transport enum to Azure's wire
// values. Unspecified deploys auto (resolved in ingress construction).
var ingressTransportStrings = map[azurecontainerappv1alpha1.AzureContainerAppIngressTransport]string{
	azurecontainerappv1alpha1.AzureContainerAppIngressTransport_AUTO:  "auto",
	azurecontainerappv1alpha1.AzureContainerAppIngressTransport_HTTP:  "http",
	azurecontainerappv1alpha1.AzureContainerAppIngressTransport_HTTP2: "http2",
	azurecontainerappv1alpha1.AzureContainerAppIngressTransport_TCP:   "tcp",
}

// clientCertificateModeStrings maps the mTLS mode enum to ARM's values.
// Unspecified is never sent (Azure's default behavior applies). ARM's
// wire vocabulary here is lowercase (accept/require/ignore) -- unlike
// most ARM enums; the SDK's Go identifiers are capitalized but the
// string constants are not.
var clientCertificateModeStrings = map[azurecontainerappv1alpha1.AzureContainerAppIngressClientCertificateMode]string{
	azurecontainerappv1alpha1.AzureContainerAppIngressClientCertificateMode_ACCEPT:  "accept",
	azurecontainerappv1alpha1.AzureContainerAppIngressClientCertificateMode_REQUIRE: "require",
	azurecontainerappv1alpha1.AzureContainerAppIngressClientCertificateMode_IGNORE:  "ignore",
}

// ipRestrictionActionStrings maps the IP-restriction action enum to ARM's
// values.
var ipRestrictionActionStrings = map[azurecontainerappv1alpha1.AzureContainerAppIpRestrictionAction]string{
	azurecontainerappv1alpha1.AzureContainerAppIpRestrictionAction_ALLOW: "Allow",
	azurecontainerappv1alpha1.AzureContainerAppIpRestrictionAction_DENY:  "Deny",
}

// volumeStorageTypeStrings maps the volume storage-type enum to ARM's
// values. Unspecified deploys EmptyDir (resolved in volume construction).
var volumeStorageTypeStrings = map[azurecontainerappv1alpha1.AzureContainerAppVolumeStorageType]string{
	azurecontainerappv1alpha1.AzureContainerAppVolumeStorageType_EMPTY_DIR:      "EmptyDir",
	azurecontainerappv1alpha1.AzureContainerAppVolumeStorageType_AZURE_FILE:     "AzureFile",
	azurecontainerappv1alpha1.AzureContainerAppVolumeStorageType_NFS_AZURE_FILE: "NfsAzureFile",
	azurecontainerappv1alpha1.AzureContainerAppVolumeStorageType_SECRET:         "Secret",
}

// daprProtocolStrings maps the Dapr app-protocol enum to Azure's wire
// values. Unspecified deploys http (resolved in dapr construction).
var daprProtocolStrings = map[azurecontainerappv1alpha1.AzureContainerAppDaprProtocol]string{
	azurecontainerappv1alpha1.AzureContainerAppDaprProtocol_DAPR_HTTP: "http",
	azurecontainerappv1alpha1.AzureContainerAppDaprProtocol_DAPR_GRPC: "grpc",
}

// identityTypeStrings maps the identity-type enum to ARM's values.
var identityTypeStrings = map[azurecontainerappv1alpha1.AzureContainerAppIdentityType]string{
	azurecontainerappv1alpha1.AzureContainerAppIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurecontainerappv1alpha1.AzureContainerAppIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurecontainerappv1alpha1.AzureContainerAppIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurecontainerappv1alpha1.AzureContainerAppIacInput) *Locals {
	locals := &Locals{}

	locals.AzureContainerApp = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureContainerApp.String()),
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
