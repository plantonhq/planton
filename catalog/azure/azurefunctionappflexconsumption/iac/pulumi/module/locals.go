package module

import (
	"strings"

	azurefunctionappflexconsumptionv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefunctionappflexconsumption/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFunctionAppFlexConsumption *azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumption
	ResourceGroupName               string
	AzureTags                       map[string]string
}

// Spec enums -> Azure's wire values. Spelled out row by row so a
// vocabulary drift fails loudly at preview time instead of deploying a
// wrong value.

var storageAuthenticationTypeStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionStorageAuthenticationType]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionStorageAuthenticationType_STORAGE_ACCOUNT_CONNECTION_STRING: "StorageAccountConnectionString",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionStorageAuthenticationType_SYSTEM_ASSIGNED_IDENTITY:          "SystemAssignedIdentity",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionStorageAuthenticationType_USER_ASSIGNED_IDENTITY:            "UserAssignedIdentity",
}

var runtimeNameStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionRuntimeName]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionRuntimeName_NODE:            "node",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionRuntimeName_DOTNET_ISOLATED: "dotnet-isolated",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionRuntimeName_JAVA:            "java",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionRuntimeName_POWERSHELL:      "powershell",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionRuntimeName_PYTHON:          "python",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionRuntimeName_CUSTOM_HANDLER:  "custom",
}

var clientCertificateModeStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionClientCertificateMode]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionClientCertificateMode_REQUIRED:                  "Required",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionClientCertificateMode_OPTIONAL:                  "Optional",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionClientCertificateMode_OPTIONAL_INTERACTIVE_USER: "OptionalInteractiveUser",
}

var tlsVersionStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionTlsVersion]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionTlsVersion_TLS_1_0: "1.0",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionTlsVersion_TLS_1_1: "1.1",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionTlsVersion_TLS_1_2: "1.2",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionTlsVersion_TLS_1_3: "1.3",
}

var loadBalancingModeStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionLoadBalancingMode]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionLoadBalancingMode_LEAST_REQUESTS:         "LeastRequests",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionLoadBalancingMode_WEIGHTED_ROUND_ROBIN:   "WeightedRoundRobin",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionLoadBalancingMode_LEAST_RESPONSE_TIME:    "LeastResponseTime",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionLoadBalancingMode_WEIGHTED_TOTAL_TRAFFIC: "WeightedTotalTraffic",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionLoadBalancingMode_REQUEST_HASH:           "RequestHash",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionLoadBalancingMode_PER_SITE_ROUND_ROBIN:   "PerSiteRoundRobin",
}

var managedPipelineModeStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionManagedPipelineMode]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionManagedPipelineMode_INTEGRATED: "Integrated",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionManagedPipelineMode_CLASSIC:    "Classic",
}

var ipRestrictionActionStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIpRestrictionAction]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIpRestrictionAction_ALLOW: "Allow",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIpRestrictionAction_DENY:  "Deny",
}

var connectionStringTypeStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType_MYSQL:            "MySQL",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType_SQL_SERVER:       "SQLServer",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType_SQL_AZURE:        "SQLAzure",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType_CUSTOM:           "Custom",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType_NOTIFICATION_HUB: "NotificationHub",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType_SERVICE_BUS:      "ServiceBus",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType_EVENT_HUB:        "EventHub",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType_API_HUB:          "APIHub",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType_DOC_DB:           "DocDb",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType_REDIS_CACHE:      "RedisCache",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionConnectionStringType_POSTGRESQL:       "PostgreSQL",
}

var identityTypeStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIdentityType]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

var unauthenticatedActionStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionUnauthenticatedAction]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionUnauthenticatedAction_REDIRECT_TO_LOGIN_PAGE: "RedirectToLoginPage",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionUnauthenticatedAction_ALLOW_ANONYMOUS:        "AllowAnonymous",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionUnauthenticatedAction_RETURN_401:             "Return401",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionUnauthenticatedAction_RETURN_403:             "Return403",
}

var forwardProxyConventionStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionForwardProxyConvention]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionForwardProxyConvention_FORWARD_PROXY_NO_PROXY: "NoProxy",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionForwardProxyConvention_FORWARD_PROXY_STANDARD: "Standard",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionForwardProxyConvention_FORWARD_PROXY_CUSTOM:   "Custom",
}

var cookieExpirationConventionStrings = map[azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionCookieExpirationConvention]string{
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionCookieExpirationConvention_FIXED_TIME:                "FixedTime",
	azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionCookieExpirationConvention_IDENTITY_PROVIDER_DERIVED: "IdentityProviderDerived",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurefunctionappflexconsumptionv1alpha1.AzureFunctionAppFlexConsumptionIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFunctionAppFlexConsumption = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureFunctionAppFlexConsumption.String()),
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
