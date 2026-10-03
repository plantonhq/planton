package module

import (
	"strings"

	azuremssqlserverv1alpha1 "github.com/plantonhq/planton/catalog/azure/azuremssqlserver/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureMssqlServer  *azuremssqlserverv1alpha1.AzureMssqlServer
	ResourceGroupName string
	AzureTags         map[string]string
}

// connectionPolicyStrings maps the spec's connection-policy enum to ARM's
// values. Unspecified is not sent at all, letting Azure apply its Default
// policy -- mirroring the Terraform module's null.
var connectionPolicyStrings = map[azuremssqlserverv1alpha1.AzureMssqlServerConnectionPolicy]string{
	azuremssqlserverv1alpha1.AzureMssqlServerConnectionPolicy_DEFAULT:  "Default",
	azuremssqlserverv1alpha1.AzureMssqlServerConnectionPolicy_PROXY:    "Proxy",
	azuremssqlserverv1alpha1.AzureMssqlServerConnectionPolicy_REDIRECT: "Redirect",
}

// identityTypeStrings maps the spec's identity-type enum to ARM's values.
var identityTypeStrings = map[azuremssqlserverv1alpha1.AzureMssqlServerIdentityType]string{
	azuremssqlserverv1alpha1.AzureMssqlServerIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azuremssqlserverv1alpha1.AzureMssqlServerIdentityType_USER_ASSIGNED:            "UserAssigned",
	azuremssqlserverv1alpha1.AzureMssqlServerIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

// alertPolicyStateStrings maps the Defender policy-state enum to ARM's
// values.
var alertPolicyStateStrings = map[azuremssqlserverv1alpha1.AzureMssqlServerSecurityAlertPolicyState]string{
	azuremssqlserverv1alpha1.AzureMssqlServerSecurityAlertPolicyState_ENABLED:  "Enabled",
	azuremssqlserverv1alpha1.AzureMssqlServerSecurityAlertPolicyState_DISABLED: "Disabled",
}

// alertTypeStrings maps the Defender detector enum to ARM's Snake_Pascal
// wire vocabulary.
var alertTypeStrings = map[azuremssqlserverv1alpha1.AzureMssqlServerSecurityAlertType]string{
	azuremssqlserverv1alpha1.AzureMssqlServerSecurityAlertType_SQL_INJECTION:               "Sql_Injection",
	azuremssqlserverv1alpha1.AzureMssqlServerSecurityAlertType_SQL_INJECTION_VULNERABILITY: "Sql_Injection_Vulnerability",
	azuremssqlserverv1alpha1.AzureMssqlServerSecurityAlertType_ACCESS_ANOMALY:              "Access_Anomaly",
	azuremssqlserverv1alpha1.AzureMssqlServerSecurityAlertType_DATA_EXFILTRATION:           "Data_Exfiltration",
	azuremssqlserverv1alpha1.AzureMssqlServerSecurityAlertType_UNSAFE_ACTION:               "Unsafe_Action",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azuremssqlserverv1alpha1.AzureMssqlServerIacInput) *Locals {
	locals := &Locals{}

	locals.AzureMssqlServer = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureMssqlServer.String()),
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
