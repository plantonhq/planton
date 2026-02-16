package module

import (
	"encoding/json"

	awswafwebaclv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awswafwebacl/v1"
)

// buildRulesJSON constructs the AWS WAFv2 API JSON representation of all rules.
// The JSON uses PascalCase keys matching the AWS API format, which is what the
// Pulumi/TF rule_json field expects.
//
// For first-class statement types (managed_rule_group, rate_based, geo_match,
// ip_set_reference), we build the JSON from the typed proto fields.
// For custom_statement, we pass the Struct map through directly.
func buildRulesJSON(spec *awswafwebaclv1.AwsWafWebAclSpec) (string, error) {
	var rules []map[string]interface{}

	for _, rule := range spec.Rules {
		ruleMap := map[string]interface{}{
			"Name":     rule.Name,
			"Priority": rule.Priority,
		}

		// Build statement.
		statement, err := buildStatement(rule)
		if err != nil {
			return "", err
		}
		ruleMap["Statement"] = statement

		// Build action or override_action.
		if rule.Action != "" {
			ruleMap["Action"] = buildAction(rule.Action, rule.CustomResponse, rule.CustomRequestHeaders)
		}
		if rule.OverrideAction != "" {
			ruleMap["OverrideAction"] = buildOverrideAction(rule.OverrideAction)
		}

		// Build visibility config with smart defaults (metric_name = rule name).
		ruleMap["VisibilityConfig"] = buildRuleVisibilityConfig(rule)

		// Rule labels.
		if len(rule.RuleLabels) > 0 {
			var labels []map[string]interface{}
			for _, label := range rule.RuleLabels {
				labels = append(labels, map[string]interface{}{"Key": label})
			}
			ruleMap["RuleLabels"] = labels
		}

		rules = append(rules, ruleMap)
	}

	bytes, err := json.Marshal(rules)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// buildStatement converts a rule's oneof statement to the AWS API JSON format.
func buildStatement(rule *awswafwebaclv1.AwsWafWebAclRule) (map[string]interface{}, error) {
	switch stmt := rule.Statement.(type) {
	case *awswafwebaclv1.AwsWafWebAclRule_ManagedRuleGroup:
		return buildManagedRuleGroupStatement(stmt.ManagedRuleGroup), nil

	case *awswafwebaclv1.AwsWafWebAclRule_RateBased:
		return buildRateBasedStatement(stmt.RateBased), nil

	case *awswafwebaclv1.AwsWafWebAclRule_GeoMatch:
		return buildGeoMatchStatement(stmt.GeoMatch), nil

	case *awswafwebaclv1.AwsWafWebAclRule_IpSetReference:
		return buildIpSetReferenceStatement(stmt.IpSetReference), nil

	case *awswafwebaclv1.AwsWafWebAclRule_CustomStatement:
		// Pass the Struct map through directly. The user writes PascalCase
		// keys matching the AWS WAFv2 API format.
		return stmt.CustomStatement.AsMap(), nil

	default:
		return map[string]interface{}{}, nil
	}
}

// buildManagedRuleGroupStatement constructs the AWS API JSON for a managed
// rule group reference.
func buildManagedRuleGroupStatement(mg *awswafwebaclv1.AwsWafWebAclManagedRuleGroupStatement) map[string]interface{} {
	stmt := map[string]interface{}{
		"Name":       mg.Name,
		"VendorName": mg.VendorName,
	}

	if mg.Version != "" {
		stmt["Version"] = mg.Version
	}

	// Rule action overrides.
	if len(mg.RuleActionOverrides) > 0 {
		var overrides []map[string]interface{}
		for _, override := range mg.RuleActionOverrides {
			overrides = append(overrides, map[string]interface{}{
				"Name":        override.Name,
				"ActionToUse": buildSimpleAction(override.Action),
			})
		}
		stmt["RuleActionOverrides"] = overrides
	}

	// Scope-down statement (Struct passed through directly).
	if mg.ScopeDownStatement != nil {
		stmt["ScopeDownStatement"] = mg.ScopeDownStatement.AsMap()
	}

	return map[string]interface{}{"ManagedRuleGroupStatement": stmt}
}

// buildRateBasedStatement constructs the AWS API JSON for a rate-based rule.
func buildRateBasedStatement(rb *awswafwebaclv1.AwsWafWebAclRateBasedStatement) map[string]interface{} {
	stmt := map[string]interface{}{
		"Limit": rb.Limit,
	}

	// Aggregate key type defaults to IP.
	aggregateKeyType := "IP"
	if rb.AggregateKeyType != "" {
		aggregateKeyType = rb.AggregateKeyType
	}
	stmt["AggregateKeyType"] = aggregateKeyType

	// Evaluation window.
	if rb.EvaluationWindowSec > 0 {
		stmt["EvaluationWindowSec"] = rb.EvaluationWindowSec
	}

	// Forwarded IP config.
	if rb.ForwardedIpConfig != nil {
		stmt["ForwardedIPConfig"] = buildForwardedIpConfig(rb.ForwardedIpConfig)
	}

	// Scope-down statement.
	if rb.ScopeDownStatement != nil {
		stmt["ScopeDownStatement"] = rb.ScopeDownStatement.AsMap()
	}

	return map[string]interface{}{"RateBasedStatement": stmt}
}

// buildGeoMatchStatement constructs the AWS API JSON for a geo match rule.
func buildGeoMatchStatement(gm *awswafwebaclv1.AwsWafWebAclGeoMatchStatement) map[string]interface{} {
	stmt := map[string]interface{}{
		"CountryCodes": gm.CountryCodes,
	}

	if gm.ForwardedIpConfig != nil {
		stmt["ForwardedIPConfig"] = buildForwardedIpConfig(gm.ForwardedIpConfig)
	}

	return map[string]interface{}{"GeoMatchStatement": stmt}
}

// buildIpSetReferenceStatement constructs the AWS API JSON for an IP set
// reference rule.
func buildIpSetReferenceStatement(ip *awswafwebaclv1.AwsWafWebAclIpSetReferenceStatement) map[string]interface{} {
	stmt := map[string]interface{}{
		"ARN": ip.Arn,
	}

	if ip.ForwardedIpConfig != nil {
		fwdConfig := map[string]interface{}{
			"HeaderName":       ip.ForwardedIpConfig.HeaderName,
			"FallbackBehavior": ip.ForwardedIpConfig.FallbackBehavior,
		}
		position := "FIRST"
		if ip.ForwardedIpConfig.Position != "" {
			position = ip.ForwardedIpConfig.Position
		}
		fwdConfig["Position"] = position
		stmt["IPSetForwardedIPConfig"] = fwdConfig
	}

	return map[string]interface{}{"IPSetReferenceStatement": stmt}
}

// buildForwardedIpConfig constructs the AWS API JSON for forwarded IP config
// used by geo_match and rate_based statements.
func buildForwardedIpConfig(fwd *awswafwebaclv1.AwsWafWebAclForwardedIpConfig) map[string]interface{} {
	return map[string]interface{}{
		"HeaderName":       fwd.HeaderName,
		"FallbackBehavior": fwd.FallbackBehavior,
	}
}

// buildAction constructs the AWS API JSON for a rule action (allow, block,
// count, captcha, challenge) with optional custom response/headers.
func buildAction(
	actionType string,
	customResponse *awswafwebaclv1.AwsWafWebAclCustomResponse,
	customHeaders []*awswafwebaclv1.AwsWafWebAclCustomHeader,
) map[string]interface{} {
	actionContent := map[string]interface{}{}

	switch actionType {
	case "block":
		blockContent := map[string]interface{}{}
		if customResponse != nil {
			cr := map[string]interface{}{
				"ResponseCode": customResponse.ResponseCode,
			}
			if customResponse.CustomResponseBodyKey != "" {
				cr["CustomResponseBodyKey"] = customResponse.CustomResponseBodyKey
			}
			if len(customResponse.ResponseHeaders) > 0 {
				var headers []map[string]interface{}
				for _, h := range customResponse.ResponseHeaders {
					headers = append(headers, map[string]interface{}{
						"Name":  h.Name,
						"Value": h.Value,
					})
				}
				cr["ResponseHeaders"] = headers
			}
			blockContent["CustomResponse"] = cr
		}
		actionContent["Block"] = blockContent

	case "allow":
		allowContent := map[string]interface{}{}
		if len(customHeaders) > 0 {
			allowContent["CustomRequestHandling"] = buildCustomRequestHandling(customHeaders)
		}
		actionContent["Allow"] = allowContent

	case "count":
		countContent := map[string]interface{}{}
		if len(customHeaders) > 0 {
			countContent["CustomRequestHandling"] = buildCustomRequestHandling(customHeaders)
		}
		actionContent["Count"] = countContent

	case "captcha":
		captchaContent := map[string]interface{}{}
		if len(customHeaders) > 0 {
			captchaContent["CustomRequestHandling"] = buildCustomRequestHandling(customHeaders)
		}
		actionContent["Captcha"] = captchaContent

	case "challenge":
		challengeContent := map[string]interface{}{}
		if len(customHeaders) > 0 {
			challengeContent["CustomRequestHandling"] = buildCustomRequestHandling(customHeaders)
		}
		actionContent["Challenge"] = challengeContent
	}

	return actionContent
}

// buildSimpleAction constructs a simple action JSON object (no custom response/headers).
// Used for rule action overrides within managed rule groups.
func buildSimpleAction(actionType string) map[string]interface{} {
	switch actionType {
	case "block":
		return map[string]interface{}{"Block": map[string]interface{}{}}
	case "allow":
		return map[string]interface{}{"Allow": map[string]interface{}{}}
	case "count":
		return map[string]interface{}{"Count": map[string]interface{}{}}
	case "captcha":
		return map[string]interface{}{"Captcha": map[string]interface{}{}}
	case "challenge":
		return map[string]interface{}{"Challenge": map[string]interface{}{}}
	default:
		return map[string]interface{}{"Count": map[string]interface{}{}}
	}
}

// buildOverrideAction constructs the AWS API JSON for an override action
// (used with managed rule group rules).
func buildOverrideAction(overrideType string) map[string]interface{} {
	if overrideType == "count" {
		return map[string]interface{}{"Count": map[string]interface{}{}}
	}
	// "none" means use the rule group's own actions.
	return map[string]interface{}{"None": map[string]interface{}{}}
}

// buildCustomRequestHandling constructs the CustomRequestHandling JSON object.
func buildCustomRequestHandling(headers []*awswafwebaclv1.AwsWafWebAclCustomHeader) map[string]interface{} {
	var insertHeaders []map[string]interface{}
	for _, h := range headers {
		insertHeaders = append(insertHeaders, map[string]interface{}{
			"Name":  h.Name,
			"Value": h.Value,
		})
	}
	return map[string]interface{}{"InsertHeaders": insertHeaders}
}

// buildRuleVisibilityConfig constructs the visibility config for a single rule,
// applying smart defaults when the user omits it.
func buildRuleVisibilityConfig(rule *awswafwebaclv1.AwsWafWebAclRule) map[string]interface{} {
	metricsEnabled := true
	sampledEnabled := true
	metricName := rule.Name

	if rule.VisibilityConfig != nil {
		metricsEnabled = rule.VisibilityConfig.CloudwatchMetricsEnabled
		sampledEnabled = rule.VisibilityConfig.SampledRequestsEnabled
		if rule.VisibilityConfig.MetricName != "" {
			metricName = rule.VisibilityConfig.MetricName
		}
	}

	return map[string]interface{}{
		"CloudWatchMetricsEnabled": metricsEnabled,
		"SampledRequestsEnabled":   sampledEnabled,
		"MetricName":               metricName,
	}
}
