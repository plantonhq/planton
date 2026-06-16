package module

import (
	"github.com/pkg/errors"
	awsroute53dnsrecordv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsroute53dnsrecord/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	awsclassicroute53 "github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsroute53dnsrecordv1.AwsRoute53DnsRecordStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	dnsRecord := locals.AwsRoute53DnsRecord
	spec := dnsRecord.Spec

	// Build the AWS provider via the shared keyless builder, which resolves the right
	// credential mechanism (static keys, keyless web identity, or the ambient chain)
	// from the provider config -- the same convergent path every other AWS pulumi
	// module uses. Region comes from the resource's spec so the provider matches it.
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create aws classic provider")
	}

	// Extract zone_id from StringValueOrRef
	zoneId := ""
	if spec.ZoneId != nil {
		zoneId = spec.ZoneId.GetValue()
	}
	if zoneId == "" {
		return errors.New("zone_id is required but was empty")
	}

	// Build record args
	recordArgs := &awsclassicroute53.RecordArgs{
		ZoneId: pulumi.String(zoneId),
		Name:   pulumi.String(spec.Name),
		Type:   pulumi.String(spec.Type.String()),
	}

	// Track if this is an alias record by checking if alias_target has a dns_name
	isAlias := spec.AliasTarget != nil &&
		spec.AliasTarget.DnsName != nil &&
		spec.AliasTarget.DnsName.GetValue() != ""

	// Set identifier for routing policies
	if spec.SetIdentifier != "" {
		recordArgs.SetIdentifier = pulumi.String(spec.SetIdentifier)
	}

	// Add health check if specified
	if spec.HealthCheckId != "" {
		recordArgs.HealthCheckId = pulumi.String(spec.HealthCheckId)
	}

	// Handle alias records vs basic records
	if isAlias {
		// Extract dns_name and zone_id from alias target's StringValueOrRef fields
		aliasDnsName := spec.AliasTarget.DnsName.GetValue()
		aliasZoneId := ""
		if spec.AliasTarget.ZoneId != nil {
			aliasZoneId = spec.AliasTarget.ZoneId.GetValue()
		}

		if aliasDnsName == "" {
			return errors.New("alias_target.dns_name is required but was empty")
		}
		if aliasZoneId == "" {
			return errors.New("alias_target.zone_id is required but was empty")
		}

		recordArgs.Aliases = awsclassicroute53.RecordAliasArray{
			&awsclassicroute53.RecordAliasArgs{
				Name:                 pulumi.String(aliasDnsName),
				ZoneId:               pulumi.String(aliasZoneId),
				EvaluateTargetHealth: pulumi.Bool(spec.AliasTarget.EvaluateTargetHealth),
			},
		}
	} else {
		// Basic record with values and TTL
		ttl := spec.Ttl
		if ttl == 0 {
			ttl = 300 // Default TTL: 300 seconds (5 minutes)
		}
		recordArgs.Ttl = pulumi.IntPtr(int(ttl))
		recordArgs.Records = pulumi.ToStringArray(spec.Values)
	}

	// Apply routing policy if specified
	if spec.RoutingPolicy != nil {
		err = applyRoutingPolicy(recordArgs, spec.RoutingPolicy)
		if err != nil {
			return errors.Wrap(err, "failed to apply routing policy")
		}
	}

	// Create the DNS record
	createdRecord, err := awsclassicroute53.NewRecord(ctx,
		dnsRecord.Metadata.Name,
		recordArgs,
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrapf(err, "failed to create DNS record %s", spec.Name)
	}

	// Export outputs
	ctx.Export(OpFqdn, createdRecord.Fqdn)
	ctx.Export(OpRecordType, pulumi.String(spec.Type.String()))
	ctx.Export(OpZoneId, pulumi.String(zoneId))
	ctx.Export(OpIsAlias, pulumi.Bool(isAlias))
	ctx.Export(OpSetIdentifier, pulumi.String(spec.SetIdentifier))

	return nil
}

// applyRoutingPolicy applies the specified routing policy to the record args
func applyRoutingPolicy(
	recordArgs *awsclassicroute53.RecordArgs,
	policy *awsroute53dnsrecordv1.AwsRoute53RoutingPolicy,
) error {
	switch p := policy.Policy.(type) {
	case *awsroute53dnsrecordv1.AwsRoute53RoutingPolicy_Weighted:
		// Weighted routing
		recordArgs.WeightedRoutingPolicies = awsclassicroute53.RecordWeightedRoutingPolicyArray{
			&awsclassicroute53.RecordWeightedRoutingPolicyArgs{
				Weight: pulumi.Int(int(p.Weighted.Weight)),
			},
		}

	case *awsroute53dnsrecordv1.AwsRoute53RoutingPolicy_Latency:
		// Latency-based routing
		recordArgs.LatencyRoutingPolicies = awsclassicroute53.RecordLatencyRoutingPolicyArray{
			&awsclassicroute53.RecordLatencyRoutingPolicyArgs{
				Region: pulumi.String(p.Latency.Region),
			},
		}

	case *awsroute53dnsrecordv1.AwsRoute53RoutingPolicy_Failover:
		// Failover routing
		failoverType := "PRIMARY"
		if p.Failover.FailoverType == awsroute53dnsrecordv1.AwsRoute53FailoverPolicy_secondary {
			failoverType = "SECONDARY"
		}
		recordArgs.FailoverRoutingPolicies = awsclassicroute53.RecordFailoverRoutingPolicyArray{
			&awsclassicroute53.RecordFailoverRoutingPolicyArgs{
				Type: pulumi.String(failoverType),
			},
		}

	case *awsroute53dnsrecordv1.AwsRoute53RoutingPolicy_Geolocation:
		// Geolocation routing
		geolocationPolicy := &awsclassicroute53.RecordGeolocationRoutingPolicyArgs{}

		if p.Geolocation.Continent != "" {
			geolocationPolicy.Continent = pulumi.String(p.Geolocation.Continent)
		}
		if p.Geolocation.Country != "" {
			geolocationPolicy.Country = pulumi.String(p.Geolocation.Country)
		}
		if p.Geolocation.Subdivision != "" {
			geolocationPolicy.Subdivision = pulumi.String(p.Geolocation.Subdivision)
		}

		recordArgs.GeolocationRoutingPolicies = awsclassicroute53.RecordGeolocationRoutingPolicyArray{
			geolocationPolicy,
		}

	default:
		// Simple routing (default) - no additional configuration needed
	}

	return nil
}
