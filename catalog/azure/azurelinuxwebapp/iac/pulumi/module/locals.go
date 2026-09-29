package module

import (
	"strings"

	azurelinuxwebappv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurelinuxwebapp/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureLinuxWebApp  *azurelinuxwebappv1alpha1.AzureLinuxWebApp
	ResourceGroupName string
	AzureTags         map[string]string
}

// Spec enums -> Azure's wire values. Spelled out row by row so a
// vocabulary drift fails loudly at preview time instead of deploying a
// wrong value.

var clientCertificateModeStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppClientCertificateMode]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppClientCertificateMode_REQUIRED:                  "Required",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppClientCertificateMode_OPTIONAL:                  "Optional",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppClientCertificateMode_OPTIONAL_INTERACTIVE_USER: "OptionalInteractiveUser",
}

var tlsVersionStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppTlsVersion]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppTlsVersion_TLS_1_0: "1.0",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppTlsVersion_TLS_1_1: "1.1",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppTlsVersion_TLS_1_2: "1.2",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppTlsVersion_TLS_1_3: "1.3",
}

var ftpsStateStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppFtpsState]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppFtpsState_ALL_ALLOWED: "AllAllowed",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppFtpsState_FTPS_ONLY:   "FtpsOnly",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppFtpsState_DISABLED:    "Disabled",
}

var loadBalancingModeStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppLoadBalancingMode]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppLoadBalancingMode_LEAST_REQUESTS:         "LeastRequests",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppLoadBalancingMode_WEIGHTED_ROUND_ROBIN:   "WeightedRoundRobin",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppLoadBalancingMode_LEAST_RESPONSE_TIME:    "LeastResponseTime",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppLoadBalancingMode_WEIGHTED_TOTAL_TRAFFIC: "WeightedTotalTraffic",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppLoadBalancingMode_REQUEST_HASH:           "RequestHash",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppLoadBalancingMode_PER_SITE_ROUND_ROBIN:   "PerSiteRoundRobin",
}

var managedPipelineModeStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppManagedPipelineMode]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppManagedPipelineMode_INTEGRATED: "Integrated",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppManagedPipelineMode_CLASSIC:    "Classic",
}

var ipRestrictionActionStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppIpRestrictionAction]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppIpRestrictionAction_ALLOW: "Allow",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppIpRestrictionAction_DENY:  "Deny",
}

var javaServerStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppJavaServer]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppJavaServer_JAVA_SE:  "JAVA",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppJavaServer_TOMCAT:   "TOMCAT",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppJavaServer_JBOSSEAP: "JBOSSEAP",
}

var connectionStringTypeStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType_MYSQL:            "MySQL",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType_SQL_SERVER:       "SQLServer",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType_SQL_AZURE:        "SQLAzure",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType_CUSTOM:           "Custom",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType_NOTIFICATION_HUB: "NotificationHub",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType_SERVICE_BUS:      "ServiceBus",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType_EVENT_HUB:        "EventHub",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType_API_HUB:          "APIHub",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType_DOC_DB:           "DocDb",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType_REDIS_CACHE:      "RedisCache",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppConnectionStringType_POSTGRESQL:       "PostgreSQL",
}

var storageMountTypeStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppStorageMountType]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppStorageMountType_AZURE_FILES: "AzureFiles",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppStorageMountType_AZURE_BLOB:  "AzureBlob",
}

var logLevelStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppLogLevel]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppLogLevel_OFF:         "Off",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppLogLevel_ERROR:       "Error",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppLogLevel_WARNING:     "Warning",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppLogLevel_INFORMATION: "Information",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppLogLevel_VERBOSE:     "Verbose",
}

var identityTypeStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppIdentityType]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppIdentityType_USER_ASSIGNED:            "UserAssigned",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

var backupFrequencyUnitStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppBackupFrequencyUnit]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppBackupFrequencyUnit_DAY:  "Day",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppBackupFrequencyUnit_HOUR: "Hour",
}

var unauthenticatedActionStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppUnauthenticatedAction]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppUnauthenticatedAction_REDIRECT_TO_LOGIN_PAGE: "RedirectToLoginPage",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppUnauthenticatedAction_ALLOW_ANONYMOUS:        "AllowAnonymous",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppUnauthenticatedAction_RETURN_401:             "Return401",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppUnauthenticatedAction_RETURN_403:             "Return403",
}

var forwardProxyConventionStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppForwardProxyConvention]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppForwardProxyConvention_FORWARD_PROXY_NO_PROXY: "NoProxy",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppForwardProxyConvention_FORWARD_PROXY_STANDARD: "Standard",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppForwardProxyConvention_FORWARD_PROXY_CUSTOM:   "Custom",
}

var cookieExpirationConventionStrings = map[azurelinuxwebappv1alpha1.AzureLinuxWebAppCookieExpirationConvention]string{
	azurelinuxwebappv1alpha1.AzureLinuxWebAppCookieExpirationConvention_FIXED_TIME:                "FixedTime",
	azurelinuxwebappv1alpha1.AzureLinuxWebAppCookieExpirationConvention_IDENTITY_PROVIDER_DERIVED: "IdentityProviderDerived",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurelinuxwebappv1alpha1.AzureLinuxWebAppStackInput) *Locals {
	locals := &Locals{}

	locals.AzureLinuxWebApp = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureLinuxWebApp.String()),
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
