package module

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	awsroute53zonev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsroute53zone/v1"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws"
	"github.com/pulumi/pulumi-aws-native/sdk/go/aws/route53"
	awsclassic "github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	awsclassicroute53 "github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context, stackInput *awsroute53zonev1.AwsRoute53ZoneStackInput) error {
	locals := initializeLocals(ctx, stackInput)
	awsRoute53Zone := locals.AwsRoute53Zone

	awsProviderConfig := stackInput.ProviderConfig
	var provider *aws.Provider
	var classicProvider *awsclassic.Provider
	var err error

	// If the user didn't provide AWS credentials, create a default provider.
	// Otherwise, inject custom credentials for the region, access key, etc.
	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx,
			"native-provider",
			&aws.ProviderArgs{
				Region: pulumi.String(locals.AwsRoute53Zone.Spec.Region),
			})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
		classicProvider, err = awsclassic.NewProvider(ctx,
			"classic-provider",
			&awsclassic.ProviderArgs{
				Region: pulumi.String(locals.AwsRoute53Zone.Spec.Region),
			})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS classic provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx,
			"native-provider",
			&aws.ProviderArgs{
				AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
				SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
				Region:    pulumi.String(locals.AwsRoute53Zone.Spec.Region),
			})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
		classicProvider, err = awsclassic.NewProvider(ctx,
			"classic-provider",
			&awsclassic.ProviderArgs{
				AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
				SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
				Region:    pulumi.String(locals.AwsRoute53Zone.Spec.Region),
				Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
			})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS classic provider with custom credentials")
		}
	}

	// Replace dots with hyphens to create valid managed-zone name
	managedZoneName := strings.ReplaceAll(awsRoute53Zone.Metadata.Name, ".", "-")

	// Create hosted zone based on zone type (public or private)
	var createdHostedZone *route53.HostedZone

	if awsRoute53Zone.Spec.IsPrivate {
		// Create private hosted zone with VPC associations
		if len(awsRoute53Zone.Spec.VpcAssociations) == 0 {
			return errors.New("private hosted zone requires at least one VPC association")
		}

		// Build VPC array for the first VPC (primary association)
		vpcs := route53.HostedZoneVpcArray{}
		for _, vpcAssoc := range awsRoute53Zone.Spec.VpcAssociations {
			vpcs = append(vpcs, &route53.HostedZoneVpcArgs{
				VpcId:     pulumi.String(vpcAssoc.VpcId.GetValue()),
				VpcRegion: pulumi.String(vpcAssoc.VpcRegion),
			})
		}

		createdHostedZone, err = route53.NewHostedZone(ctx,
			managedZoneName,
			&route53.HostedZoneArgs{
				Name: pulumi.String(awsRoute53Zone.Metadata.Name),
				Vpcs: vpcs,
			}, pulumi.Provider(provider))

		if err != nil {
			return errors.Wrapf(err, "failed to create private hosted-zone for %s domain",
				awsRoute53Zone.Metadata.Name)
		}
	} else {
		// Create public hosted zone
		createdHostedZone, err = route53.NewHostedZone(ctx,
			managedZoneName,
			&route53.HostedZoneArgs{
				Name: pulumi.String(awsRoute53Zone.Metadata.Name),
			}, pulumi.Provider(provider))

		if err != nil {
			return errors.Wrapf(err, "failed to create public hosted-zone for %s domain",
				awsRoute53Zone.Metadata.Name)
		}
	}

	// Enable DNSSEC if requested
	if awsRoute53Zone.Spec.EnableDnssec {
		_, err = awsclassicroute53.NewHostedZoneDnsSec(ctx,
			fmt.Sprintf("%s-dnssec", managedZoneName),
			&awsclassicroute53.HostedZoneDnsSecArgs{
				HostedZoneId: createdHostedZone.ID(),
			}, pulumi.Provider(classicProvider))

		if err != nil {
			return errors.Wrap(err, "failed to enable DNSSEC for hosted zone")
		}
	}

	// Enable query logging if requested
	if awsRoute53Zone.Spec.EnableQueryLogging {
		if awsRoute53Zone.Spec.QueryLogGroupName == "" {
			return errors.New("query_log_group_name is required when enable_query_logging is true")
		}

		_, err = awsclassicroute53.NewQueryLog(ctx,
			fmt.Sprintf("%s-query-log", managedZoneName),
			&awsclassicroute53.QueryLogArgs{
			CloudwatchLogGroupArn: pulumi.Sprintf("arn:aws:logs:%s:%s:log-group:%s",
				locals.AwsRoute53Zone.Spec.Region,
					"AWS::AccountId", // Pulumi will resolve this
					awsRoute53Zone.Spec.QueryLogGroupName),
				ZoneId: createdHostedZone.ID(),
			}, pulumi.Provider(classicProvider))

		if err != nil {
			return errors.Wrap(err, "failed to enable query logging for hosted zone")
		}
	}

	// Export important information about created hosted-zone as outputs
	ctx.Export(OpZoneName, createdHostedZone.Name)
	ctx.Export(OpZoneId, createdHostedZone.ID())
	ctx.Export(OpNameservers, createdHostedZone.NameServers)

	// Create DNS records
	for index, dnsRecord := range awsRoute53Zone.Spec.Records {
		err := createDnsRecord(ctx, classicProvider, createdHostedZone.ID(), index, dnsRecord)
		if err != nil {
			return errors.Wrapf(err, "failed to create DNS record %d: %s", index, dnsRecord.Name)
		}
	}

	return nil
}

// createDnsRecord creates a DNS record with support for basic records, alias records, and routing policies
func createDnsRecord(
	ctx *pulumi.Context,
	provider *awsclassic.Provider,
	zoneId pulumi.IDOutput,
	index int,
	dnsRecord *awsroute53zonev1.Route53DnsRecord,
) error {
	recordName := fmt.Sprintf("dns-record-%d", index)

	// Determine TTL (not applicable for alias records)
	ttlSeconds := dnsRecord.TtlSeconds
	if ttlSeconds == 0 && dnsRecord.AliasTarget == nil {
		ttlSeconds = 300 // Default TTL: 300 seconds (5 minutes)
	}

	// Build record args based on record type (basic, alias, or with routing policy)
	recordArgs := &awsclassicroute53.RecordArgs{
		ZoneId: zoneId,
		Name:   pulumi.String(dnsRecord.Name),
		Type:   pulumi.String(dnsRecord.RecordType.String()),
	}

	// Set identifier if using routing policies
	if dnsRecord.SetIdentifier != "" {
		recordArgs.SetIdentifier = pulumi.String(dnsRecord.SetIdentifier)
	}

	// Add health check if specified
	if dnsRecord.HealthCheckId != "" {
		recordArgs.HealthCheckId = pulumi.String(dnsRecord.HealthCheckId)
	}

	// Handle alias records
	if dnsRecord.AliasTarget != nil {
		recordArgs.Aliases = awsclassicroute53.RecordAliasArray{
			&awsclassicroute53.RecordAliasArgs{
				Name:                 pulumi.String(dnsRecord.AliasTarget.DnsName),
				ZoneId:               pulumi.String(dnsRecord.AliasTarget.HostedZoneId),
				EvaluateTargetHealth: pulumi.Bool(dnsRecord.AliasTarget.EvaluateTargetHealth),
			},
		}
	} else {
		// Basic record with values
		recordArgs.Ttl = pulumi.IntPtr(int(ttlSeconds))
		recordArgs.Records = pulumi.ToStringArray(dnsRecord.Values)
	}

	// Add routing policy if specified
	if dnsRecord.RoutingPolicy != nil {
		err := applyRoutingPolicy(recordArgs, dnsRecord.RoutingPolicy)
		if err != nil {
			return err
		}
	}

	// Create the record
	_, err := awsclassicroute53.NewRecord(ctx, recordName, recordArgs, pulumi.Provider(provider))
	return err
}

// applyRoutingPolicy applies the specified routing policy to the record args
func applyRoutingPolicy(
	recordArgs *awsclassicroute53.RecordArgs,
	policy *awsroute53zonev1.Route53RoutingPolicy,
) error {
	switch p := policy.Policy.(type) {
	case *awsroute53zonev1.Route53RoutingPolicy_Weighted:
		// Weighted routing
		recordArgs.WeightedRoutingPolicies = awsclassicroute53.RecordWeightedRoutingPolicyArray{
			&awsclassicroute53.RecordWeightedRoutingPolicyArgs{
				Weight: pulumi.Int(int(p.Weighted.Weight)),
			},
		}

	case *awsroute53zonev1.Route53RoutingPolicy_Latency:
		// Latency-based routing
		recordArgs.LatencyRoutingPolicies = awsclassicroute53.RecordLatencyRoutingPolicyArray{
			&awsclassicroute53.RecordLatencyRoutingPolicyArgs{
				Region: pulumi.String(p.Latency.Region),
			},
		}

	case *awsroute53zonev1.Route53RoutingPolicy_Failover:
		// Failover routing
		failoverType := "PRIMARY"
		if p.Failover.Type == awsroute53zonev1.Route53FailoverRoutingPolicy_SECONDARY {
			failoverType = "SECONDARY"
		}
		recordArgs.FailoverRoutingPolicies = awsclassicroute53.RecordFailoverRoutingPolicyArray{
			&awsclassicroute53.RecordFailoverRoutingPolicyArgs{
				Type: pulumi.String(failoverType),
			},
		}

	case *awsroute53zonev1.Route53RoutingPolicy_Geolocation:
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
