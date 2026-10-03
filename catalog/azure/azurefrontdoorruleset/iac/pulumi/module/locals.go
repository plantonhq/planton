package module

import (
	azurefrontdoorrulesetv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorruleset/v1alpha1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFrontDoorRuleSet *azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleSet
	ProfileId             string
}

// operatorStrings maps the shared operator enum to ARM's exact
// (case-sensitive) operator names. Which subset a condition accepts is
// enforced by the spec's per-field vocabulary; the wire name is the
// same everywhere.
var operatorStrings = map[azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator]string{
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_ANY:                   "Any",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_EQUAL:                 "Equal",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_CONTAINS:              "Contains",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_BEGINS_WITH:           "BeginsWith",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_ENDS_WITH:             "EndsWith",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_GREATER_THAN:          "GreaterThan",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_GREATER_THAN_OR_EQUAL: "GreaterThanOrEqual",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_LESS_THAN:             "LessThan",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_LESS_THAN_OR_EQUAL:    "LessThanOrEqual",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_REG_EX:                "RegEx",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_WILDCARD:              "Wildcard",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_GEO_MATCH:             "GeoMatch",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleOperator_IP_MATCH:              "IPMatch",
}

// transformStrings maps the transform enum to ARM's values.
var transformStrings = map[azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleTransform]string{
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleTransform_LOWERCASE:    "Lowercase",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleTransform_UPPERCASE:    "Uppercase",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleTransform_TRIM:         "Trim",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleTransform_URL_DECODE:   "UrlDecode",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleTransform_URL_ENCODE:   "UrlEncode",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleTransform_REMOVE_NULLS: "RemoveNulls",
}

// behaviorOnMatchStrings maps the post-match behavior enum to ARM's
// values (unspecified deploys Continue, Azure's default).
var behaviorOnMatchStrings = map[azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleBehaviorOnMatch]string{
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleBehaviorOnMatch_CONTINUE: "Continue",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleBehaviorOnMatch_STOP:     "Stop",
}

// redirectTypeStrings maps the redirect-status enum to ARM's values.
var redirectTypeStrings = map[azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleRedirectType]string{
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleRedirectType_MOVED:              "Moved",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleRedirectType_FOUND:              "Found",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleRedirectType_TEMPORARY_REDIRECT: "TemporaryRedirect",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleRedirectType_PERMANENT_REDIRECT: "PermanentRedirect",
}

// The shared forwarding-protocol enum maps to DIFFERENT wire
// vocabularies on its two consumers -- the redirect action speaks
// Http/Https/MatchRequest while the route-configuration override
// speaks HttpOnly/HttpsOnly/MatchRequest. Same semantics, two ARM
// dialects; each action maps its own.
var redirectProtocolStrings = map[azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleForwardingProtocol]string{
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleForwardingProtocol_MATCH_REQUEST: "MatchRequest",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleForwardingProtocol_HTTP_ONLY:     "Http",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleForwardingProtocol_HTTPS_ONLY:    "Https",
}

var overrideForwardingProtocolStrings = map[azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleForwardingProtocol]string{
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleForwardingProtocol_MATCH_REQUEST: "MatchRequest",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleForwardingProtocol_HTTP_ONLY:     "HttpOnly",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleForwardingProtocol_HTTPS_ONLY:    "HttpsOnly",
}

// headerActionStrings maps the header-action enum to ARM's values.
var headerActionStrings = map[azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleHeaderActionType]string{
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleHeaderActionType_APPEND:    "Append",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleHeaderActionType_OVERWRITE: "Overwrite",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleHeaderActionType_DELETE:    "Delete",
}

// cacheBehaviorStrings maps the override cache-behavior enum to ARM's
// values.
var cacheBehaviorStrings = map[azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleCacheBehavior]string{
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleCacheBehavior_HONOR_ORIGIN:               "HonorOrigin",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleCacheBehavior_OVERRIDE_ALWAYS:            "OverrideAlways",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleCacheBehavior_OVERRIDE_IF_ORIGIN_MISSING: "OverrideIfOriginMissing",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleCacheBehavior_DISABLED:                   "Disabled",
}

// queryStringCachingBehaviorStrings maps the override caching-key enum
// to ARM's values.
var queryStringCachingBehaviorStrings = map[azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleQueryStringCachingBehavior]string{
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleQueryStringCachingBehavior_IGNORE_QUERY_STRING:             "IgnoreQueryString",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleQueryStringCachingBehavior_USE_QUERY_STRING:                "UseQueryString",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleQueryStringCachingBehavior_IGNORE_SPECIFIED_QUERY_STRINGS:  "IgnoreSpecifiedQueryStrings",
	azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleQueryStringCachingBehavior_INCLUDE_SPECIFIED_QUERY_STRINGS: "IncludeSpecifiedQueryStrings",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurefrontdoorrulesetv1alpha1.AzureFrontDoorRuleSetIacInput) *Locals {
	locals := &Locals{}

	locals.AzureFrontDoorRuleSet = iacInput.Target
	locals.ProfileId = iacInput.Target.Spec.ProfileId.GetValue()

	// No Azure tags: ARM does not support tags on Front Door rule sets
	// or rules, so the platform's identity tags live on the profile.

	return locals
}
