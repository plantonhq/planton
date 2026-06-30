package module

import (
	"fmt"
	"strings"

	ociapplicationloadbalancerv1 "github.com/plantonhq/planton/apis/dev/planton/provider/oci/ociapplicationloadbalancer/v1"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci"
	"github.com/pulumi/pulumi-oci/sdk/v4/go/oci/loadbalancer"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

var actionMap = map[ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_Action]string{
	ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_add_http_request_header:           "ADD_HTTP_REQUEST_HEADER",
	ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_add_http_response_header:          "ADD_HTTP_RESPONSE_HEADER",
	ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_extend_http_request_header_value:  "EXTEND_HTTP_REQUEST_HEADER_VALUE",
	ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_extend_http_response_header_value: "EXTEND_HTTP_RESPONSE_HEADER_VALUE",
	ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_remove_http_request_header:        "REMOVE_HTTP_REQUEST_HEADER",
	ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_remove_http_response_header:       "REMOVE_HTTP_RESPONSE_HEADER",
	ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_redirect:                          "REDIRECT",
	ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_allow:                             "ALLOW",
	ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_control_access_using_http_methods: "CONTROL_ACCESS_USING_HTTP_METHODS",
	ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_http_header:                       "HTTP_HEADER",
	ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem_ip_based_max_connections:          "IP_BASED_MAX_CONNECTIONS",
}

func createRuleSets(ctx *pulumi.Context, locals *Locals, provider *oci.Provider, lb *loadbalancer.LoadBalancer) ([]*loadbalancer.RuleSet, error) {
	spec := locals.OciApplicationLoadBalancer.Spec
	var created []*loadbalancer.RuleSet

	for _, rsSpec := range spec.RuleSets {
		items := buildRuleSetItems(rsSpec.Items)

		rs, err := loadbalancer.NewRuleSet(ctx, rsSpec.Name, &loadbalancer.RuleSetArgs{
			LoadBalancerId: lb.ID(),
			Name:           pulumi.String(rsSpec.Name),
			Items:          items,
		}, pulumiOciOpt(provider), pulumi.Parent(lb))
		if err != nil {
			return nil, fmt.Errorf("failed to create rule set %s: %w", rsSpec.Name, err)
		}
		created = append(created, rs)
	}

	return created, nil
}

func buildRuleSetItems(items []*ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem) loadbalancer.RuleSetItemArray {
	result := make(loadbalancer.RuleSetItemArray, len(items))
	for i, item := range items {
		result[i] = buildRuleSetItem(item)
	}
	return result
}

func buildRuleSetItem(item *ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItem) *loadbalancer.RuleSetItemArgs {
	action := actionMap[item.Action]
	if action == "" {
		action = strings.ToUpper(item.Action.String())
	}

	args := &loadbalancer.RuleSetItemArgs{
		Action: pulumi.String(action),
	}

	if item.Header != "" {
		args.Header = pulumi.StringPtr(item.Header)
	}
	if item.Value != "" {
		args.Value = pulumi.StringPtr(item.Value)
	}
	if item.Prefix != "" {
		args.Prefix = pulumi.StringPtr(item.Prefix)
	}
	if item.Suffix != "" {
		args.Suffix = pulumi.StringPtr(item.Suffix)
	}
	if item.Description != "" {
		args.Description = pulumi.StringPtr(item.Description)
	}
	if item.ResponseCode > 0 {
		args.ResponseCode = pulumi.IntPtr(int(item.ResponseCode))
	}
	if item.StatusCode > 0 {
		args.StatusCode = pulumi.IntPtr(int(item.StatusCode))
	}

	if item.RedirectUri != nil {
		args.RedirectUri = buildRedirectUri(item.RedirectUri)
	}

	if len(item.Conditions) > 0 {
		args.Conditions = buildConditions(item.Conditions)
	}

	if len(item.AllowedMethods) > 0 {
		args.AllowedMethods = pulumi.ToStringArray(item.AllowedMethods)
	}

	if item.AreInvalidCharactersAllowed {
		args.AreInvalidCharactersAllowed = pulumi.BoolPtr(true)
	}
	if item.HttpLargeHeaderSizeInKb > 0 {
		args.HttpLargeHeaderSizeInKb = pulumi.IntPtr(int(item.HttpLargeHeaderSizeInKb))
	}
	if item.DefaultMaxConnections > 0 {
		args.DefaultMaxConnections = pulumi.IntPtr(int(item.DefaultMaxConnections))
	}

	if len(item.IpMaxConnections) > 0 {
		ipConns := make(loadbalancer.RuleSetItemIpMaxConnectionArray, len(item.IpMaxConnections))
		for j, ipc := range item.IpMaxConnections {
			ipConns[j] = &loadbalancer.RuleSetItemIpMaxConnectionArgs{
				IpAddresses:    pulumi.ToStringArray(ipc.IpAddresses),
				MaxConnections: pulumi.IntPtr(int(ipc.MaxConnections)),
			}
		}
		args.IpMaxConnections = ipConns
	}

	return args
}

func buildRedirectUri(ru *ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RedirectUri) *loadbalancer.RuleSetItemRedirectUriArgs {
	args := &loadbalancer.RuleSetItemRedirectUriArgs{}
	if ru.Protocol != "" {
		args.Protocol = pulumi.StringPtr(ru.Protocol)
	}
	if ru.Host != "" {
		args.Host = pulumi.StringPtr(ru.Host)
	}
	if ru.Port > 0 {
		args.Port = pulumi.IntPtr(int(ru.Port))
	}
	if ru.Path != "" {
		args.Path = pulumi.StringPtr(ru.Path)
	}
	if ru.Query != "" {
		args.Query = pulumi.StringPtr(ru.Query)
	}
	return args
}

func buildConditions(conditions []*ociapplicationloadbalancerv1.OciApplicationLoadBalancerSpec_RuleSetItemCondition) loadbalancer.RuleSetItemConditionArray {
	result := make(loadbalancer.RuleSetItemConditionArray, len(conditions))
	for i, c := range conditions {
		args := &loadbalancer.RuleSetItemConditionArgs{
			AttributeName:  pulumi.String(c.AttributeName),
			AttributeValue: pulumi.String(c.AttributeValue),
		}
		if c.Operator != "" {
			args.Operator = pulumi.StringPtr(c.Operator)
		}
		result[i] = args
	}
	return result
}
