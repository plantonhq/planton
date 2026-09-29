package module

import (
	"strings"

	azurefrontdoorfirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurefrontdoorfirewallpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/cloudresourcekind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureFrontDoorFirewallPolicy *azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicy
	ResourceGroupName            string
	AzureTags                    map[string]string
	// SkuName is ARM's tier value, materialized from the spec enum with
	// the documented STANDARD default (stack inputs never carry proto
	// defaults).
	SkuName string
	// IsPremium gates the Premium-only policy settings (the
	// JS-challenge/CAPTCHA expirations, which Azure ALWAYS enables on
	// Premium and REJECTS on Standard).
	IsPremium bool
}

// skuStrings maps the spec's sku enum to ARM's tier values.
var skuStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicySku]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicySku_STANDARD: "Standard_AzureFrontDoor",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicySku_PREMIUM:  "Premium_AzureFrontDoor",
}

// modeStrings maps the enforcement-mode enum to ARM's values.
var modeStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMode]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMode_DETECTION:  "Detection",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMode_PREVENTION: "Prevention",
}

// customRuleTypeStrings maps the custom-rule type enum to ARM's values.
var customRuleTypeStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyCustomRuleType]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyCustomRuleType_MATCH_RULE:      "MatchRule",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyCustomRuleType_RATE_LIMIT_RULE: "RateLimitRule",
}

// customRuleActionStrings maps the custom-rule action enum to ARM's
// values (JSChallenge/CAPTCHA are Premium-only, spec-enforced).
var customRuleActionStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyCustomRuleAction]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyCustomRuleAction_ALLOW:        "Allow",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyCustomRuleAction_BLOCK:        "Block",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyCustomRuleAction_LOG:          "Log",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyCustomRuleAction_REDIRECT:     "Redirect",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyCustomRuleAction_JS_CHALLENGE: "JSChallenge",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyCustomRuleAction_CAPTCHA:      "CAPTCHA",
}

// matchVariableStrings maps the match-variable enum to ARM's values.
var matchVariableStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMatchVariable]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMatchVariable_COOKIES:        "Cookies",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMatchVariable_POST_ARGS:      "PostArgs",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMatchVariable_QUERY_STRING:   "QueryString",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMatchVariable_REMOTE_ADDR:    "RemoteAddr",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMatchVariable_REQUEST_BODY:   "RequestBody",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMatchVariable_REQUEST_HEADER: "RequestHeader",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMatchVariable_REQUEST_METHOD: "RequestMethod",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMatchVariable_REQUEST_URI:    "RequestUri",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyMatchVariable_SOCKET_ADDR:    "SocketAddr",
}

// operatorStrings maps the condition-operator enum to ARM's exact
// (case-sensitive) operator names.
var operatorStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_ANY:                   "Any",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_BEGINS_WITH:           "BeginsWith",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_CONTAINS:              "Contains",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_ENDS_WITH:             "EndsWith",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_EQUAL:                 "Equal",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_GEO_MATCH:             "GeoMatch",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_GREATER_THAN:          "GreaterThan",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_GREATER_THAN_OR_EQUAL: "GreaterThanOrEqual",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_IP_MATCH:              "IPMatch",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_LESS_THAN:             "LessThan",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_LESS_THAN_OR_EQUAL:    "LessThanOrEqual",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyOperator_REG_EX:                "RegEx",
}

// transformStrings maps the transform enum to ARM's values. Note the
// canonical casing is "UrlDecode"/"UrlEncode" (the SDK constants'
// STRING VALUES) -- the provider validates case-sensitively, so the
// SDK's URLDecode/URLEncode Go identifiers are NOT the wire values.
var transformStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyTransform]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyTransform_LOWERCASE:    "Lowercase",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyTransform_REMOVE_NULLS: "RemoveNulls",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyTransform_TRIM:         "Trim",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyTransform_UPPERCASE:    "Uppercase",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyTransform_URL_DECODE:   "UrlDecode",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyTransform_URL_ENCODE:   "UrlEncode",
}

// managedRuleSetActionStrings maps the rule-set action enum to ARM's
// values (the RULE_SET_ prefix exists only to keep the proto enum
// names collision-free within the kind).
var managedRuleSetActionStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleSetAction]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleSetAction_RULE_SET_BLOCK:    "Block",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleSetAction_RULE_SET_LOG:      "Log",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleSetAction_RULE_SET_REDIRECT: "Redirect",
}

// exclusionMatchVariableStrings maps the exclusion-collection enum to
// ARM's values.
var exclusionMatchVariableStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyExclusionMatchVariable]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyExclusionMatchVariable_EXCLUDE_QUERY_STRING_ARG_NAMES:      "QueryStringArgNames",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyExclusionMatchVariable_EXCLUDE_REQUEST_BODY_JSON_ARG_NAMES: "RequestBodyJsonArgNames",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyExclusionMatchVariable_EXCLUDE_REQUEST_BODY_POST_ARG_NAMES: "RequestBodyPostArgNames",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyExclusionMatchVariable_EXCLUDE_REQUEST_COOKIE_NAMES:        "RequestCookieNames",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyExclusionMatchVariable_EXCLUDE_REQUEST_HEADER_NAMES:        "RequestHeaderNames",
}

// selectorOperatorStrings maps the selector-operator enum to ARM's
// values (shared by managed-rule exclusions and log scrubbing).
var selectorOperatorStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicySelectorOperator]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_CONTAINS:    "Contains",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_ENDS_WITH:   "EndsWith",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_EQUALS:      "Equals",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_EQUALS_ANY:  "EqualsAny",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicySelectorOperator_SELECTOR_STARTS_WITH: "StartsWith",
}

// managedRuleOverrideActionStrings maps the per-rule override action
// enum to ARM's values (the OVERRIDE_ prefix exists only to keep the
// proto enum names collision-free within the kind).
var managedRuleOverrideActionStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleOverrideAction]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_ALLOW:           "Allow",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_ANOMALY_SCORING: "AnomalyScoring",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_BLOCK:           "Block",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_CAPTCHA:         "CAPTCHA",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_JS_CHALLENGE:    "JSChallenge",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_REDIRECT:        "Redirect",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyManagedRuleOverrideAction_OVERRIDE_LOG:             "Log",
}

// scrubbingMatchVariableStrings maps the scrubbing-variable enum to
// ARM's values (a superset of the profile's access-log vocabulary --
// the WAF logs carry more request parts).
var scrubbingMatchVariableStrings = map[azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyScrubbingMatchVariable]string{
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_QUERY_STRING_ARG_NAMES:      "QueryStringArgNames",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_BODY_JSON_ARG_NAMES: "RequestBodyJsonArgNames",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_BODY_POST_ARG_NAMES: "RequestBodyPostArgNames",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_COOKIE_NAMES:        "RequestCookieNames",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_HEADER_NAMES:        "RequestHeaderNames",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_IP_ADDRESS:          "RequestIPAddress",
	azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_URI:                 "RequestUri",
}

func initializeLocals(ctx *pulumi.Context, stackInput *azurefrontdoorfirewallpolicyv1alpha1.AzureFrontDoorFirewallPolicyStackInput) *Locals {
	locals := &Locals{}

	locals.AzureFrontDoorFirewallPolicy = stackInput.Target
	target := stackInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	// Materialize the tier default: unspecified deploys STANDARD (the
	// spec's documented default -- stack inputs never carry proto
	// defaults).
	locals.SkuName = skuStrings[target.Spec.Sku]
	if locals.SkuName == "" {
		locals.SkuName = "Standard_AzureFrontDoor"
	}
	locals.IsPremium = locals.SkuName == "Premium_AzureFrontDoor"

	// Metadata-derived tags first, then the user's spec tags merged over
	// them: user tags deliberately win so an org's governance conventions
	// (cost center, owner) can override the derived values where they
	// collide.
	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(cloudresourcekind.CloudResourceKind_AzureFrontDoorFirewallPolicy.String()),
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
