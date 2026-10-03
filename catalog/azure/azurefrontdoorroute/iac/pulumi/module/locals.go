package module

import (
	azurefrontdoorroutev1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorroute/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFrontDoorRoute *azurefrontdoorroutev1alpha1.AzureFrontDoorRoute
	EndpointId          string
	OriginGroupId       string
	OriginIds           []string
	RuleSetIds          []string
	CustomDomainIds     []string
}

// protocolStrings maps the client-facing protocol enum to ARM's values.
var protocolStrings = map[azurefrontdoorroutev1alpha1.AzureFrontDoorRouteProtocol]string{
	azurefrontdoorroutev1alpha1.AzureFrontDoorRouteProtocol_HTTP:  "Http",
	azurefrontdoorroutev1alpha1.AzureFrontDoorRouteProtocol_HTTPS: "Https",
}

// forwardingProtocolStrings maps the origin-leg protocol enum to ARM's
// values (unspecified deploys MatchRequest, Azure's default).
var forwardingProtocolStrings = map[azurefrontdoorroutev1alpha1.AzureFrontDoorRouteForwardingProtocol]string{
	azurefrontdoorroutev1alpha1.AzureFrontDoorRouteForwardingProtocol_MATCH_REQUEST: "MatchRequest",
	azurefrontdoorroutev1alpha1.AzureFrontDoorRouteForwardingProtocol_HTTP_ONLY:     "HttpOnly",
	azurefrontdoorroutev1alpha1.AzureFrontDoorRouteForwardingProtocol_HTTPS_ONLY:    "HttpsOnly",
}

// queryStringCachingBehaviorStrings maps the cache-key enum to ARM's
// values (unspecified deploys IgnoreQueryString, Azure's default).
var queryStringCachingBehaviorStrings = map[azurefrontdoorroutev1alpha1.AzureFrontDoorRouteQueryStringCachingBehavior]string{
	azurefrontdoorroutev1alpha1.AzureFrontDoorRouteQueryStringCachingBehavior_IGNORE_QUERY_STRING:             "IgnoreQueryString",
	azurefrontdoorroutev1alpha1.AzureFrontDoorRouteQueryStringCachingBehavior_USE_QUERY_STRING:                "UseQueryString",
	azurefrontdoorroutev1alpha1.AzureFrontDoorRouteQueryStringCachingBehavior_IGNORE_SPECIFIED_QUERY_STRINGS:  "IgnoreSpecifiedQueryStrings",
	azurefrontdoorroutev1alpha1.AzureFrontDoorRouteQueryStringCachingBehavior_INCLUDE_SPECIFIED_QUERY_STRINGS: "IncludeSpecifiedQueryStrings",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurefrontdoorroutev1alpha1.AzureFrontDoorRouteIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFrontDoorRoute = iacInput.Target
	spec := iacInput.Target.Spec

	locals.EndpointId = spec.EndpointId.GetValue()
	locals.OriginGroupId = spec.OriginGroupId.GetValue()

	for _, originId := range spec.OriginIds {
		locals.OriginIds = append(locals.OriginIds, originId.GetValue())
	}

	for _, ruleSetId := range spec.RuleSetIds {
		locals.RuleSetIds = append(locals.RuleSetIds, ruleSetId.GetValue())
	}

	for _, customDomainId := range spec.CustomDomainIds {
		locals.CustomDomainIds = append(locals.CustomDomainIds, customDomainId.GetValue())
	}

	// No Azure tags: ARM does not support tags on Front Door routes, so
	// the platform's identity tags live on the profile.

	return locals
}
