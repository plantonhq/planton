package module

import (
	"strings"

	azurefunctionappv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefunctionapp/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFunctionApp  *azurefunctionappv1alpha1.AzureFunctionApp
	ResourceGroupName string
	AzureTags         map[string]string
}

// Spec enums -> Azure's wire values. Spelled out row by row so a
// vocabulary drift fails loudly at preview time instead of deploying a
// wrong value.

var clientCertificateModeStrings = map[azurefunctionappv1alpha1.AzureFunctionAppClientCertificateMode]string{
	azurefunctionappv1alpha1.AzureFunctionAppClientCertificateMode_REQUIRED:                  "Required",
	azurefunctionappv1alpha1.AzureFunctionAppClientCertificateMode_OPTIONAL:                  "Optional",
	azurefunctionappv1alpha1.AzureFunctionAppClientCertificateMode_OPTIONAL_INTERACTIVE_USER: "OptionalInteractiveUser",
}

var tlsVersionStrings = map[azurefunctionappv1alpha1.AzureFunctionAppTlsVersion]string{
	azurefunctionappv1alpha1.AzureFunctionAppTlsVersion_TLS_1_0: "1.0",
	azurefunctionappv1alpha1.AzureFunctionAppTlsVersion_TLS_1_1: "1.1",
	azurefunctionappv1alpha1.AzureFunctionAppTlsVersion_TLS_1_2: "1.2",
	azurefunctionappv1alpha1.AzureFunctionAppTlsVersion_TLS_1_3: "1.3",
}

var ftpsStateStrings = map[azurefunctionappv1alpha1.AzureFunctionAppFtpsState]string{
	azurefunctionappv1alpha1.AzureFunctionAppFtpsState_ALL_ALLOWED: "AllAllowed",
	azurefunctionappv1alpha1.AzureFunctionAppFtpsState_FTPS_ONLY:   "FtpsOnly",
	azurefunctionappv1alpha1.AzureFunctionAppFtpsState_DISABLED:    "Disabled",
}

var loadBalancingModeStrings = map[azurefunctionappv1alpha1.AzureFunctionAppLoadBalancingMode]string{
	azurefunctionappv1alpha1.AzureFunctionAppLoadBalancingMode_LEAST_REQUESTS:         "LeastRequests",
	azurefunctionappv1alpha1.AzureFunctionAppLoadBalancingMode_WEIGHTED_ROUND_ROBIN:   "WeightedRoundRobin",
	azurefunctionappv1alpha1.AzureFunctionAppLoadBalancingMode_LEAST_RESPONSE_TIME:    "LeastResponseTime",
	azurefunctionappv1alpha1.AzureFunctionAppLoadBalancingMode_WEIGHTED_TOTAL_TRAFFIC: "WeightedTotalTraffic",
	azurefunctionappv1alpha1.AzureFunctionAppLoadBalancingMode_REQUEST_HASH:           "RequestHash",
	azurefunctionappv1alpha1.AzureFunctionAppLoadBalancingMode_PER_SITE_ROUND_ROBIN:   "PerSiteRoundRobin",
}

var managedPipelineModeStrings = map[azurefunctionappv1alpha1.AzureFunctionAppManagedPipelineMode]string{
	azurefunctionappv1alpha1.AzureFunctionAppManagedPipelineMode_INTEGRATED: "Integrated",
	azurefunctionappv1alpha1.AzureFunctionAppManagedPipelineMode_CLASSIC:    "Classic",
}

var ipRestrictionActionStrings = map[azurefunctionappv1alpha1.AzureFunctionAppIpRestrictionAction]string{
	azurefunctionappv1alpha1.AzureFunctionAppIpRestrictionAction_ALLOW: "Allow",
	azurefunctionappv1alpha1.AzureFunctionAppIpRestrictionAction_DENY:  "Deny",
}

var connectionStringTypeStrings = map[azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType]string{
	azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType_MYSQL:            "MySQL",
	azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType_SQL_SERVER:       "SQLServer",
	azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType_SQL_AZURE:        "SQLAzure",
	azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType_CUSTOM:           "Custom",
	azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType_NOTIFICATION_HUB: "NotificationHub",
	azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType_SERVICE_BUS:      "ServiceBus",
	azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType_EVENT_HUB:        "EventHub",
	azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType_API_HUB:          "APIHub",
	azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType_DOC_DB:           "DocDb",
	azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType_REDIS_CACHE:      "RedisCache",
	azurefunctionappv1alpha1.AzureFunctionAppConnectionStringType_POSTGRESQL:       "PostgreSQL",
}

var storageMountTypeStrings = map[azurefunctionappv1alpha1.AzureFunctionAppStorageMountType]string{
	azurefunctionappv1alpha1.AzureFunctionAppStorageMountType_AZURE_FILES: "AzureFiles",
	azurefunctionappv1alpha1.AzureFunctionAppStorageMountType_AZURE_BLOB:  "AzureBlob",
}

var identityTypeStrings = map[azurefunctionappv1alpha1.AzureFunctionAppIdentityType]string{
	azurefunctionappv1alpha1.AzureFunctionAppIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurefunctionappv1alpha1.AzureFunctionAppIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurefunctionappv1alpha1.AzureFunctionAppIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

var backupFrequencyUnitStrings = map[azurefunctionappv1alpha1.AzureFunctionAppBackupFrequencyUnit]string{
	azurefunctionappv1alpha1.AzureFunctionAppBackupFrequencyUnit_DAY:  "Day",
	azurefunctionappv1alpha1.AzureFunctionAppBackupFrequencyUnit_HOUR: "Hour",
}

var unauthenticatedActionStrings = map[azurefunctionappv1alpha1.AzureFunctionAppUnauthenticatedAction]string{
	azurefunctionappv1alpha1.AzureFunctionAppUnauthenticatedAction_REDIRECT_TO_LOGIN_PAGE: "RedirectToLoginPage",
	azurefunctionappv1alpha1.AzureFunctionAppUnauthenticatedAction_ALLOW_ANONYMOUS:        "AllowAnonymous",
	azurefunctionappv1alpha1.AzureFunctionAppUnauthenticatedAction_RETURN_401:             "Return401",
	azurefunctionappv1alpha1.AzureFunctionAppUnauthenticatedAction_RETURN_403:             "Return403",
}

var forwardProxyConventionStrings = map[azurefunctionappv1alpha1.AzureFunctionAppForwardProxyConvention]string{
	azurefunctionappv1alpha1.AzureFunctionAppForwardProxyConvention_FORWARD_PROXY_NO_PROXY: "NoProxy",
	azurefunctionappv1alpha1.AzureFunctionAppForwardProxyConvention_FORWARD_PROXY_STANDARD: "Standard",
	azurefunctionappv1alpha1.AzureFunctionAppForwardProxyConvention_FORWARD_PROXY_CUSTOM:   "Custom",
}

var cookieExpirationConventionStrings = map[azurefunctionappv1alpha1.AzureFunctionAppCookieExpirationConvention]string{
	azurefunctionappv1alpha1.AzureFunctionAppCookieExpirationConvention_FIXED_TIME:                "FixedTime",
	azurefunctionappv1alpha1.AzureFunctionAppCookieExpirationConvention_IDENTITY_PROVIDER_DERIVED: "IdentityProviderDerived",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurefunctionappv1alpha1.AzureFunctionAppIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFunctionApp = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureFunctionApp.String()),
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
