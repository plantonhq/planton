package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/wafv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// webAcl creates the WAFv2 Web ACL resource. Rules are passed via RuleJson
// to handle both typed first-class statements and custom_statement Struct
// escape hatches uniformly.
func webAcl(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*wafv2.WebAcl, error) {
	spec := locals.WebAcl.Spec

	// Default action: allow or block.
	var defaultAction *wafv2.WebAclDefaultActionArgs
	if spec.DefaultAction.Type == "block" {
		defaultAction = &wafv2.WebAclDefaultActionArgs{
			Block: &wafv2.WebAclDefaultActionBlockArgs{},
		}
	} else {
		defaultAction = &wafv2.WebAclDefaultActionArgs{
			Allow: &wafv2.WebAclDefaultActionAllowArgs{},
		}
	}

	// Visibility config with smart defaults.
	metricName := locals.WebAcl.Metadata.Name
	metricsEnabled := true
	sampledEnabled := true
	if spec.VisibilityConfig != nil {
		if spec.VisibilityConfig.MetricName != "" {
			metricName = spec.VisibilityConfig.MetricName
		}
		metricsEnabled = spec.VisibilityConfig.CloudwatchMetricsEnabled
		sampledEnabled = spec.VisibilityConfig.SampledRequestsEnabled
	}

	args := &wafv2.WebAclArgs{
		Name:          pulumi.String(locals.WebAcl.Metadata.Name),
		Scope:         pulumi.String(spec.Scope),
		DefaultAction: defaultAction,
		VisibilityConfig: &wafv2.WebAclVisibilityConfigArgs{
			CloudwatchMetricsEnabled: pulumi.Bool(metricsEnabled),
			SampledRequestsEnabled:   pulumi.Bool(sampledEnabled),
			MetricName:               pulumi.String(metricName),
		},
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// Description.
	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	// Token domains.
	if len(spec.TokenDomains) > 0 {
		args.TokenDomains = pulumi.ToStringArray(spec.TokenDomains)
	}

	// Custom response bodies.
	if len(spec.CustomResponseBodies) > 0 {
		bodies := wafv2.WebAclCustomResponseBodyArray{}
		for _, body := range spec.CustomResponseBodies {
			bodies = append(bodies, &wafv2.WebAclCustomResponseBodyArgs{
				Key:         pulumi.String(body.Key),
				Content:     pulumi.String(body.Content),
				ContentType: pulumi.String(body.ContentType),
			})
		}
		args.CustomResponseBodies = bodies
	}

	// Build rules as JSON for maximum flexibility.
	if len(spec.Rules) > 0 {
		rulesJSON, err := buildRulesJSON(spec)
		if err != nil {
			return nil, errors.Wrap(err, "failed to build rules JSON")
		}
		args.RuleJson = pulumi.StringPtr(rulesJSON)
	}

	createdAcl, err := wafv2.NewWebAcl(ctx, locals.WebAcl.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create WAFv2 Web ACL")
	}

	// Export outputs.
	ctx.Export(OpWebAclArn, createdAcl.Arn)
	ctx.Export(OpWebAclId, createdAcl.ID())
	ctx.Export(OpWebAclName, createdAcl.Name)
	ctx.Export(OpCapacity, createdAcl.Capacity)

	return createdAcl, nil
}
