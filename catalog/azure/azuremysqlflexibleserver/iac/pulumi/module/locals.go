package module

import (
	"strings"

	azuremysqlflexibleserverv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremysqlflexibleserver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMysqlFlexibleServer *azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServer
	ResourceGroupName        string
	AzureTags                map[string]string
}

// createModeStrings maps the spec's create-mode enum to ARM's values. An
// unspecified mode means a fresh (DEFAULT) server and is not sent at all,
// mirroring the Terraform module's null -- so the two engines produce the
// same ARM payload.
var createModeStrings = map[azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerCreateMode]string{
	azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerCreateMode_DEFAULT:               "Default",
	azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerCreateMode_POINT_IN_TIME_RESTORE: "PointInTimeRestore",
	azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerCreateMode_REPLICA:               "Replica",
	azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerCreateMode_GEO_RESTORE:           "GeoRestore",
}

// haModeStrings maps the spec's high-availability mode enum to ARM's values.
var haModeStrings = map[azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerHighAvailabilityMode]string{
	azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerHighAvailabilityMode_ZONE_REDUNDANT: "ZoneRedundant",
	azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerHighAvailabilityMode_SAME_ZONE:      "SameZone",
}

// publicNetworkAccessStrings maps the spec's public-network-access enum to
// ARM's values. Unspecified is not sent at all, letting Azure derive the
// value (Enabled publicly, Disabled when VNet-injected) -- mirroring the
// Terraform module's null.
var publicNetworkAccessStrings = map[azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerPublicNetworkAccess]string{
	azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerPublicNetworkAccess_ENABLED:  "Enabled",
	azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerPublicNetworkAccess_DISABLED: "Disabled",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azuremysqlflexibleserverv1alpha1.AzureMysqlFlexibleServerStackInput) *Locals {
	locals := &Locals{}

	locals.AzureMysqlFlexibleServer = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureMysqlFlexibleServer.String()),
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

	// The user's spec tags merge over the metadata-derived tags -- user
	// tags deliberately win so an org's governance conventions can
	// override the derived values where they collide.
	for key, value := range target.Spec.Tags {
		locals.AzureTags[key] = value
	}

	return locals
}
