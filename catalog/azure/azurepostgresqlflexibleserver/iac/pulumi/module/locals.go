package module

import (
	"strings"

	azurepostgresqlflexibleserverv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurepostgresqlflexibleserver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzurePostgresqlFlexibleServer *azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServer
	ResourceGroupName             string
	AzureTags                     map[string]string
}

// createModeStrings maps the spec's create-mode enum to ARM's values. An
// unspecified mode means a fresh (DEFAULT) server and is not sent at all,
// mirroring the Terraform module's null -- so the two engines produce the
// same ARM payload.
var createModeStrings = map[azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerCreateMode]string{
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerCreateMode_DEFAULT:               "Default",
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerCreateMode_POINT_IN_TIME_RESTORE: "PointInTimeRestore",
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerCreateMode_REPLICA:               "Replica",
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerCreateMode_GEO_RESTORE:           "GeoRestore",
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerCreateMode_REVIVE_DROPPED:        "ReviveDropped",
}

// haModeStrings maps the spec's high-availability mode enum to ARM's values.
var haModeStrings = map[azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerHighAvailabilityMode]string{
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerHighAvailabilityMode_ZONE_REDUNDANT: "ZoneRedundant",
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerHighAvailabilityMode_SAME_ZONE:      "SameZone",
}

// identityTypeStrings maps the spec's identity-type enum to ARM's values.
var identityTypeStrings = map[azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerIdentityType]string{
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

// principalTypeStrings maps the spec's Entra principal-type enum to ARM's
// values.
var principalTypeStrings = map[azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerAadPrincipalType]string{
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerAadPrincipalType_USER:              "User",
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerAadPrincipalType_GROUP:             "Group",
	azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerAadPrincipalType_SERVICE_PRINCIPAL: "ServicePrincipal",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurepostgresqlflexibleserverv1alpha1.AzurePostgresqlFlexibleServerIacInput) *Locals {
	locals := &Locals{}

	locals.AzurePostgresqlFlexibleServer = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzurePostgresqlFlexibleServer.String()),
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
