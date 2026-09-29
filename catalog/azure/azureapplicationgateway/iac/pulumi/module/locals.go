package module

import (
	"strings"

	azureapplicationgatewayv1alpha1 "github.com/plantonhq/planton/catalog/azure/azureapplicationgateway/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureApplicationGateway *azureapplicationgatewayv1alpha1.AzureApplicationGateway
	ResourceGroupName       string
	AzureTags               map[string]string
	GatewayIpConfigName     string
}

// The spec's enums arrive as proto enum values; each map below carries the
// complete vocabulary for its enum, translated to ARM's strings. A missing
// entry would silently drop the setting, so the maps are exhaustive by
// construction.

// On the v2 platform (and Basic), SKU name and tier carry the same value.
var skuStrings = map[azureapplicationgatewayv1alpha1.AzureApplicationGatewaySku]string{
	azureapplicationgatewayv1alpha1.AzureApplicationGatewaySku_BASIC:       "Basic",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewaySku_STANDARD_V2: "Standard_v2",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewaySku_WAF_V2:      "WAF_v2",
}

var protocolStrings = map[azureapplicationgatewayv1alpha1.AzureApplicationGatewayProtocol]string{
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayProtocol_HTTP:  "Http",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayProtocol_HTTPS: "Https",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayProtocol_TCP:   "Tcp",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayProtocol_TLS:   "Tls",
}

var ruleTypeStrings = map[azureapplicationgatewayv1alpha1.AzureApplicationGatewayRequestRoutingRuleType]string{
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayRequestRoutingRuleType_BASIC_ROUTING:      "Basic",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayRequestRoutingRuleType_PATH_BASED_ROUTING: "PathBasedRouting",
}

var ipAllocationStrings = map[azureapplicationgatewayv1alpha1.AzureApplicationGatewayIpAllocation]string{
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayIpAllocation_DYNAMIC: "Dynamic",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayIpAllocation_STATIC:  "Static",
}

var identityTypeStrings = map[azureapplicationgatewayv1alpha1.AzureApplicationGatewayIdentityType]string{
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayIdentityType_SYSTEM_ASSIGNED:          "SystemAssigned",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayIdentityType_USER_ASSIGNED:            "UserAssigned",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayIdentityType_SYSTEM_AND_USER_ASSIGNED: "SystemAssigned, UserAssigned",
}

var sslPolicyTypeStrings = map[azureapplicationgatewayv1alpha1.AzureApplicationGatewaySslPolicyType]string{
	azureapplicationgatewayv1alpha1.AzureApplicationGatewaySslPolicyType_PREDEFINED: "Predefined",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewaySslPolicyType_CUSTOM:     "Custom",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewaySslPolicyType_CUSTOM_V2:  "CustomV2",
}

var tlsProtocolStrings = map[azureapplicationgatewayv1alpha1.AzureApplicationGatewayTlsProtocol]string{
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayTlsProtocol_TLS_V1_0: "TLSv1_0",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayTlsProtocol_TLS_V1_1: "TLSv1_1",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayTlsProtocol_TLS_V1_2: "TLSv1_2",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayTlsProtocol_TLS_V1_3: "TLSv1_3",
}

var redirectTypeStrings = map[azureapplicationgatewayv1alpha1.AzureApplicationGatewayRedirectType]string{
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayRedirectType_PERMANENT: "Permanent",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayRedirectType_FOUND:     "Found",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayRedirectType_SEE_OTHER: "SeeOther",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayRedirectType_TEMPORARY: "Temporary",
}

var urlComponentStrings = map[azureapplicationgatewayv1alpha1.AzureApplicationGatewayRewriteRuleUrlComponent]string{
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayRewriteRuleUrlComponent_PATH_ONLY:         "path_only",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayRewriteRuleUrlComponent_QUERY_STRING_ONLY: "query_string_only",
}

var statusCodeStrings = map[azureapplicationgatewayv1alpha1.AzureApplicationGatewayCustomErrorStatusCode]string{
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayCustomErrorStatusCode_HTTP_STATUS_400: "HttpStatus400",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayCustomErrorStatusCode_HTTP_STATUS_403: "HttpStatus403",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayCustomErrorStatusCode_HTTP_STATUS_404: "HttpStatus404",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayCustomErrorStatusCode_HTTP_STATUS_405: "HttpStatus405",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayCustomErrorStatusCode_HTTP_STATUS_408: "HttpStatus408",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayCustomErrorStatusCode_HTTP_STATUS_500: "HttpStatus500",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayCustomErrorStatusCode_HTTP_STATUS_502: "HttpStatus502",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayCustomErrorStatusCode_HTTP_STATUS_503: "HttpStatus503",
	azureapplicationgatewayv1alpha1.AzureApplicationGatewayCustomErrorStatusCode_HTTP_STATUS_504: "HttpStatus504",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azureapplicationgatewayv1alpha1.AzureApplicationGatewayStackInput) *Locals {
	locals := &Locals{}

	locals.AzureApplicationGateway = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// The gateway's single IP configuration, derived from the dedicated
	// subnet -- pure ARM plumbing users never name. Matches the Terraform
	// module's derivation.
	locals.GatewayIpConfigName = target.Spec.Name + "-gateway-ip"

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureApplicationGateway.String()),
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
