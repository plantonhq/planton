package module

import (
	"github.com/pkg/errors"
	awsroute53dnsrecordv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsroute53dnsrecord/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	foreignkeyv1 "github.com/plantonhq/planton/shared/foreignkey/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates one DNS resource record set. The record's AWS identity is
// (zone, name, type, set_identifier); it is either a STANDARD record
// (values + ttl) or an ALIAS record (alias_target) — never both
// (CEL-enforced) — and carries at most one routing policy (a proto oneof).
func Resources(ctx *pulumi.Context, iacInput *awsroute53dnsrecordv1alpha1.AwsRoute53DnsRecordIacInput) error {
	locals := initializeLocals(ctx, iacInput)
	spec := locals.AwsRoute53DnsRecord.Spec

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web
	// identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	zoneId := spec.ZoneId.GetValue()
	if zoneId == "" {
		return errors.New("zone_id resolved to an empty value")
	}

	recordArgs := &route53.RecordArgs{
		ZoneId: pulumi.String(zoneId),
		Name:   pulumi.String(spec.Name),
		Type:   pulumi.String(spec.Type),
		// Overwrite an existing record set with the same name and type
		// instead of failing — the adoption path for records created outside
		// the graph.
		AllowOverwrite: pulumi.Bool(spec.AllowOverwrite),
	}

	if spec.SetIdentifier != "" {
		recordArgs.SetIdentifier = pulumi.String(spec.SetIdentifier)
	}

	// Health check gating this record's answers — valid with any non-simple
	// routing policy, most commonly the failover PRIMARY.
	if spec.HealthCheckId.GetValue() != "" {
		recordArgs.HealthCheckId = pulumi.String(spec.HealthCheckId.GetValue())
	}

	isAlias := spec.AliasTarget != nil
	if isAlias {
		// Alias records point at an AWS resource's DNS name + that service's
		// OWN hosted zone ID (not this record's zone). No TTL — the target's
		// TTL applies.
		recordArgs.Aliases = route53.RecordAliasArray{
			&route53.RecordAliasArgs{
				Name:                 pulumi.String(spec.AliasTarget.DnsName.GetValue()),
				ZoneId:               pulumi.String(spec.AliasTarget.ZoneId.GetValue()),
				EvaluateTargetHealth: pulumi.Bool(spec.AliasTarget.EvaluateTargetHealth),
			},
		}
	} else {
		// The CEL contract guarantees standard records carry a TTL; an
		// explicit 0 ("never cache") passes through faithfully.
		recordArgs.Ttl = pulumi.IntPtr(int(spec.GetTtl()))
		recordArgs.Records = pulumi.ToStringArray(stringValues(spec.Values))
	}

	if spec.RoutingPolicy != nil {
		applyRoutingPolicy(recordArgs, spec.RoutingPolicy)
	}

	createdRecord, err := route53.NewRecord(ctx,
		locals.AwsRoute53DnsRecord.Metadata.Name,
		recordArgs,
		pulumi.Provider(provider))
	if err != nil {
		return errors.Wrapf(err, "failed to create DNS record %s", spec.Name)
	}

	ctx.Export(OpFqdn, createdRecord.Fqdn)
	ctx.Export(OpRecordType, pulumi.String(spec.Type))
	ctx.Export(OpZoneId, pulumi.String(zoneId))
	ctx.Export(OpIsAlias, pulumi.Bool(isAlias))
	ctx.Export(OpSetIdentifier, pulumi.String(spec.SetIdentifier))

	return nil
}

// applyRoutingPolicy maps the spec's oneof onto the provider's seven mutually
// exclusive routing-policy arguments.
func applyRoutingPolicy(
	recordArgs *route53.RecordArgs,
	policy *awsroute53dnsrecordv1alpha1.AwsRoute53RoutingPolicy,
) {
	switch p := policy.Policy.(type) {
	case *awsroute53dnsrecordv1alpha1.AwsRoute53RoutingPolicy_Weighted:
		recordArgs.WeightedRoutingPolicies = route53.RecordWeightedRoutingPolicyArray{
			&route53.RecordWeightedRoutingPolicyArgs{
				Weight: pulumi.Int(int(p.Weighted.Weight)),
			},
		}

	case *awsroute53dnsrecordv1alpha1.AwsRoute53RoutingPolicy_Latency:
		recordArgs.LatencyRoutingPolicies = route53.RecordLatencyRoutingPolicyArray{
			&route53.RecordLatencyRoutingPolicyArgs{
				Region: pulumi.String(p.Latency.Region),
			},
		}

	case *awsroute53dnsrecordv1alpha1.AwsRoute53RoutingPolicy_Failover:
		recordArgs.FailoverRoutingPolicies = route53.RecordFailoverRoutingPolicyArray{
			&route53.RecordFailoverRoutingPolicyArgs{
				Type: pulumi.String(p.Failover.FailoverType),
			},
		}

	case *awsroute53dnsrecordv1alpha1.AwsRoute53RoutingPolicy_Geolocation:
		geolocationPolicy := &route53.RecordGeolocationRoutingPolicyArgs{}
		if p.Geolocation.Continent != "" {
			geolocationPolicy.Continent = pulumi.String(p.Geolocation.Continent)
		}
		if p.Geolocation.Country != "" {
			geolocationPolicy.Country = pulumi.String(p.Geolocation.Country)
		}
		if p.Geolocation.Subdivision != "" {
			geolocationPolicy.Subdivision = pulumi.String(p.Geolocation.Subdivision)
		}
		recordArgs.GeolocationRoutingPolicies = route53.RecordGeolocationRoutingPolicyArray{
			geolocationPolicy,
		}

	case *awsroute53dnsrecordv1alpha1.AwsRoute53RoutingPolicy_Geoproximity:
		// Exactly one location determinant (CEL-enforced): region,
		// coordinates, or Local Zone group; bias widens/narrows the catchment.
		geoproximityPolicy := &route53.RecordGeoproximityRoutingPolicyArgs{}
		if p.Geoproximity.AwsRegion != "" {
			geoproximityPolicy.AwsRegion = pulumi.String(p.Geoproximity.AwsRegion)
		}
		if p.Geoproximity.LocalZoneGroup != "" {
			geoproximityPolicy.LocalZoneGroup = pulumi.String(p.Geoproximity.LocalZoneGroup)
		}
		if p.Geoproximity.Bias != 0 {
			geoproximityPolicy.Bias = pulumi.Int(int(p.Geoproximity.Bias))
		}
		if p.Geoproximity.Coordinates != nil {
			geoproximityPolicy.Coordinates = route53.RecordGeoproximityRoutingPolicyCoordinateArray{
				&route53.RecordGeoproximityRoutingPolicyCoordinateArgs{
					Latitude:  pulumi.String(p.Geoproximity.Coordinates.Latitude),
					Longitude: pulumi.String(p.Geoproximity.Coordinates.Longitude),
				},
			}
		}
		recordArgs.GeoproximityRoutingPolicy = geoproximityPolicy

	case *awsroute53dnsrecordv1alpha1.AwsRoute53RoutingPolicy_Cidr:
		recordArgs.CidrRoutingPolicy = &route53.RecordCidrRoutingPolicyArgs{
			CollectionId: pulumi.String(p.Cidr.CollectionId),
			LocationName: pulumi.String(p.Cidr.LocationName),
		}

	case *awsroute53dnsrecordv1alpha1.AwsRoute53RoutingPolicy_MultivalueAnswer:
		// A bare flag in the provider; the spec models it as an empty policy
		// message so the oneof stays uniform.
		recordArgs.MultivalueAnswerRoutingPolicy = pulumi.Bool(true)
	}
}

// stringValues resolves each value-or-reference entry to the string the
// provider receives (references are resolved to values before the module runs).
func stringValues(entries []*foreignkeyv1.StringValueOrRef) []string {
	values := make([]string, 0, len(entries))
	for _, entry := range entries {
		values = append(values, entry.GetValue())
	}
	return values
}
