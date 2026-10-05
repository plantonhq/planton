package module

import (
	"strings"

	azurewebapplicationfirewallpolicyv1alpha1 "github.com/plantonhq/planton/catalog/azure/azurewebapplicationfirewallpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/azure/azuretagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AzureWebApplicationFirewallPolicy *azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicy
	ResourceGroupName                 string
	AzureTags                         map[string]string
}

// The spec's enums arrive as proto enum values; each map below carries the
// complete vocabulary for its enum, translated to ARM's strings. A missing
// entry would silently drop the setting, so the maps are exhaustive by
// construction.

var ruleTypeStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyCustomRuleType]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyCustomRuleType_MATCH_RULE:      "MatchRule",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyCustomRuleType_RATE_LIMIT_RULE: "RateLimitRule",
}

var customRuleActionStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyCustomRuleAction]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyCustomRuleAction_ALLOW:        "Allow",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyCustomRuleAction_BLOCK:        "Block",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyCustomRuleAction_LOG:          "Log",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyCustomRuleAction_JS_CHALLENGE: "JSChallenge",
}

var rateLimitDurationStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyRateLimitDuration]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyRateLimitDuration_ONE_MIN:   "OneMin",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyRateLimitDuration_FIVE_MINS: "FiveMins",
}

var groupRateLimitByStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyGroupRateLimitBy]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyGroupRateLimitBy_CLIENT_ADDR:             "ClientAddr",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyGroupRateLimitBy_CLIENT_ADDR_XFF_HEADER:  "ClientAddrXFFHeader",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyGroupRateLimitBy_GEO_LOCATION:            "GeoLocation",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyGroupRateLimitBy_GEO_LOCATION_XFF_HEADER: "GeoLocationXFFHeader",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyGroupRateLimitBy_NONE:                    "None",
}

var matchVariableNameStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchVariableName]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchVariableName_REMOTE_ADDR:     "RemoteAddr",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchVariableName_REQUEST_METHOD:  "RequestMethod",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchVariableName_QUERY_STRING:    "QueryString",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchVariableName_POST_ARGS:       "PostArgs",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchVariableName_REQUEST_URI:     "RequestUri",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchVariableName_REQUEST_HEADERS: "RequestHeaders",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchVariableName_REQUEST_BODY:    "RequestBody",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchVariableName_REQUEST_COOKIES: "RequestCookies",
}

var matchOperatorStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_ANY:                   "Any",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_IP_MATCH:              "IPMatch",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_GEO_MATCH:             "GeoMatch",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_EQUAL:                 "Equal",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_CONTAINS:              "Contains",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_LESS_THAN:             "LessThan",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_GREATER_THAN:          "GreaterThan",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_LESS_THAN_OR_EQUAL:    "LessThanOrEqual",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_GREATER_THAN_OR_EQUAL: "GreaterThanOrEqual",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_BEGINS_WITH:           "BeginsWith",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_ENDS_WITH:             "EndsWith",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMatchOperator_REGEX:                 "Regex",
}

var transformStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyTransform]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyTransform_HTML_ENTITY_DECODE: "HtmlEntityDecode",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyTransform_LOWERCASE:          "Lowercase",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyTransform_REMOVE_NULLS:       "RemoveNulls",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyTransform_TRIM:               "Trim",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyTransform_URL_DECODE:         "UrlDecode",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyTransform_URL_ENCODE:         "UrlEncode",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyTransform_UPPERCASE:          "Uppercase",
}

var managedRuleSetTypeStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyManagedRuleSetType]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyManagedRuleSetType_OWASP:                          "OWASP",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyManagedRuleSetType_MICROSOFT_BOT_MANAGER_RULE_SET: "Microsoft_BotManagerRuleSet",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyManagedRuleSetType_MICROSOFT_DEFAULT_RULE_SET:     "Microsoft_DefaultRuleSet",
}

var ruleOverrideActionStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyRuleOverrideAction]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyRuleOverrideAction_OVERRIDE_ALLOW:           "Allow",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyRuleOverrideAction_OVERRIDE_ANOMALY_SCORING: "AnomalyScoring",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyRuleOverrideAction_OVERRIDE_BLOCK:           "Block",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyRuleOverrideAction_OVERRIDE_JS_CHALLENGE:    "JSChallenge",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyRuleOverrideAction_OVERRIDE_LOG:             "Log",
}

var exclusionMatchVariableStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyExclusionMatchVariable]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_ARG_KEYS:      "RequestArgKeys",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_ARG_NAMES:     "RequestArgNames",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_ARG_VALUES:    "RequestArgValues",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_COOKIE_KEYS:   "RequestCookieKeys",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_COOKIE_NAMES:  "RequestCookieNames",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_COOKIE_VALUES: "RequestCookieValues",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_HEADER_KEYS:   "RequestHeaderKeys",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_HEADER_NAMES:  "RequestHeaderNames",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyExclusionMatchVariable_REQUEST_HEADER_VALUES: "RequestHeaderValues",
}

var selectorMatchOperatorStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicySelectorMatchOperator]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_EQUALS:      "Equals",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_CONTAINS:    "Contains",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_STARTS_WITH: "StartsWith",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_ENDS_WITH:   "EndsWith",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicySelectorMatchOperator_SELECTOR_EQUALS_ANY:  "EqualsAny",
}

var scrubbingMatchVariableStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyScrubbingMatchVariable]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_ARG_NAMES:      "RequestArgNames",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_COOKIE_NAMES:   "RequestCookieNames",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_HEADER_NAMES:   "RequestHeaderNames",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_IP_ADDRESS:     "RequestIPAddress",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_JSON_ARG_NAMES: "RequestJSONArgNames",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyScrubbingMatchVariable_SCRUB_REQUEST_POST_ARG_NAMES: "RequestPostArgNames",
}

var modeStrings = map[azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMode]string{
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMode_PREVENTION: "Prevention",
	azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyMode_DETECTION:  "Detection",
}

func initializeLocals(ctx *pulumi.Context, iacInput *azurewebapplicationfirewallpolicyv1alpha1.AzureWebApplicationFirewallPolicyIacInput) *Locals {
	locals := &Locals{}

	locals.AzureWebApplicationFirewallPolicy = iacInput.Target
	target := iacInput.Target

	locals.ResourceGroupName = target.Spec.ResourceGroup.GetValue()

	locals.AzureTags = map[string]string{
		azuretagkeys.Resource:     "true",
		azuretagkeys.ResourceName: target.Metadata.Name,
		azuretagkeys.ResourceKind: strings.ToLower(catalogkind.CatalogKind_AzureWebApplicationFirewallPolicy.String()),
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
