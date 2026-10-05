package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/apigatewayv2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// domainName creates the API Gateway v2 custom domain and its API mappings.
// The domain binds an owned DNS name to an ACM certificate; mappings then
// publish APIs under the domain (optionally namespaced by a path key). DNS is
// composed downstream: the exported target_domain_name / hosted_zone_id feed
// a Route 53 alias record.
func domainName(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.AwsHttpApiDomain.Spec

	// API Gateway v2 domains accept exactly one endpoint type (REGIONAL) and
	// one security policy (TLS_1_2) -- AWS validates against single-value
	// enums -- so neither is a spec field. Modeling them would be two
	// decorative knobs a user could never meaningfully turn.
	configArgs := &apigatewayv2.DomainNameDomainNameConfigurationArgs{
		CertificateArn: pulumi.String(spec.CertificateArn.GetValue()),
		EndpointType:   pulumi.String("REGIONAL"),
		SecurityPolicy: pulumi.String("TLS_1_2"),
	}

	// ipv4 or dualstack; AWS applies its default when unset.
	if spec.IpAddressType != "" {
		configArgs.IpAddressType = pulumi.StringPtr(spec.IpAddressType)
	}

	// AWS-issued public certificate proving domain ownership -- required by
	// AWS when the TLS certificate is Private-CA-issued, or when mTLS is
	// combined with an ACM-imported certificate.
	if spec.OwnershipVerificationCertificateArn.GetValue() != "" {
		configArgs.OwnershipVerificationCertificateArn = pulumi.StringPtr(spec.OwnershipVerificationCertificateArn.GetValue())
	}

	args := &apigatewayv2.DomainNameArgs{
		// The domain's AWS identity is spec.domain_name itself (it IS the
		// resource); metadata.name only feeds the identity tags.
		DomainName:              pulumi.String(spec.DomainName),
		DomainNameConfiguration: configArgs,
		Tags:                    pulumi.ToStringMap(locals.AwsTags),
	}

	// How requests route: static api_mappings only (AWS default), routing
	// rules only, or rules first with mapping fallback. Spec CEL guarantees
	// the mode and the routing_rules list agree.
	if spec.RoutingMode != "" {
		args.RoutingMode = pulumi.StringPtr(spec.RoutingMode)
	}

	// Mutual TLS: clients must present a certificate chaining to a CA in the
	// S3-hosted truststore. Pinning truststore_version makes CA rotation an
	// explicit change instead of a silent side effect of overwriting the
	// object.
	if spec.MutualTls != nil {
		mtlsArgs := &apigatewayv2.DomainNameMutualTlsAuthenticationArgs{
			TruststoreUri: pulumi.String(spec.MutualTls.TruststoreUri),
		}
		if spec.MutualTls.TruststoreVersion != "" {
			mtlsArgs.TruststoreVersion = pulumi.StringPtr(spec.MutualTls.TruststoreVersion)
		}
		args.MutualTlsAuthentication = mtlsArgs
	}

	createdDomain, err := apigatewayv2.NewDomainName(ctx, locals.AwsHttpApiDomain.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create domain name")
	}

	// One mapping resource per entry, addressed by path key (the domain root
	// maps under the "root" alias -- the alias never reaches AWS). Referenced
	// API IDs arrive pre-resolved by the orchestrator.
	for _, mapping := range spec.ApiMappings {
		aliasKey := mapping.ApiMappingKey
		if aliasKey == "" {
			aliasKey = "root"
		}
		resourceName := fmt.Sprintf("%s-mapping-%s", locals.AwsHttpApiDomain.Metadata.Name, aliasKey)

		mappingArgs := &apigatewayv2.ApiMappingArgs{
			ApiId:      pulumi.String(mapping.ApiId.GetValue()),
			DomainName: createdDomain.ID(),
			Stage:      pulumi.String(mapping.Stage),
		}
		// Empty means the domain root; AWS stores the absence of a key,
		// not "".
		if mapping.ApiMappingKey != "" {
			mappingArgs.ApiMappingKey = pulumi.StringPtr(mapping.ApiMappingKey)
		}

		if _, err := apigatewayv2.NewApiMapping(ctx, resourceName, mappingArgs, pulumi.Provider(provider)); err != nil {
			return errors.Wrapf(err, "failed to create api mapping %q", aliasKey)
		}
	}

	// One routing rule per entry, addressed by priority (unique per spec
	// CEL, mirroring AWS's own uniqueness rule -- and stable: re-ordering
	// the spec list never churns rule resources). Rules match on base path
	// or header and invoke one API stage; the provider nests the target
	// under action/invoke_api -- the spec's flat api_id/stage/
	// strip_base_path fields render into that wrapper here.
	for _, rule := range spec.RoutingRules {
		resourceName := fmt.Sprintf("%s-rule-%d", locals.AwsHttpApiDomain.Metadata.Name, rule.Priority)

		// Each condition tests exactly one dimension (spec CEL mirrors the
		// provider's exactly-one-of); all conditions on a rule must match.
		conditionArgs := make(apigatewayv2.RoutingRuleConditionArray, 0, len(rule.Conditions))
		for _, condition := range rule.Conditions {
			conditionArg := apigatewayv2.RoutingRuleConditionArgs{}
			if len(condition.BasePaths) > 0 {
				conditionArg.MatchBasePaths = &apigatewayv2.RoutingRuleConditionMatchBasePathsArgs{
					AnyOfs: pulumi.ToStringArray(condition.BasePaths),
				}
			}
			if condition.Header != nil {
				conditionArg.MatchHeaders = &apigatewayv2.RoutingRuleConditionMatchHeadersArgs{
					AnyOf: &apigatewayv2.RoutingRuleConditionMatchHeadersAnyOfArgs{
						Header:    pulumi.String(condition.Header.Name),
						ValueGlob: pulumi.String(condition.Header.ValueGlob),
					},
				}
			}
			conditionArgs = append(conditionArgs, conditionArg)
		}

		ruleArgs := &apigatewayv2.RoutingRuleArgs{
			DomainName: createdDomain.ID(),
			Priority:   pulumi.Int(int(rule.Priority)),
			Actions: apigatewayv2.RoutingRuleActionArray{
				apigatewayv2.RoutingRuleActionArgs{
					InvokeApi: &apigatewayv2.RoutingRuleActionInvokeApiArgs{
						ApiId:         pulumi.String(rule.ApiId.GetValue()),
						Stage:         pulumi.String(rule.Stage),
						StripBasePath: pulumi.Bool(rule.StripBasePath),
					},
				},
			},
			Conditions: conditionArgs,
		}

		if _, err := apigatewayv2.NewRoutingRule(ctx, resourceName, ruleArgs, pulumi.Provider(provider)); err != nil {
			return errors.Wrapf(err, "failed to create routing rule priority %d", rule.Priority)
		}
	}

	// Export outputs matching AwsHttpApiDomainOutputs. The nested
	// configuration outputs (target domain + hosted zone) are the DNS
	// composition surface.
	ctx.Export(OpDomainName, createdDomain.ID())
	ctx.Export(OpDomainNameArn, createdDomain.Arn)
	ctx.Export(OpTargetDomainName, createdDomain.DomainNameConfiguration.TargetDomainName().Elem())
	ctx.Export(OpHostedZoneId, createdDomain.DomainNameConfiguration.HostedZoneId().Elem())

	return nil
}
