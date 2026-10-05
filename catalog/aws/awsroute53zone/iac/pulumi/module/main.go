package module

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
	awsroute53zonev1alpha1 "github.com/plantonhq/planton/catalog/aws/awsroute53zone/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources creates the Route 53 hosted zone plus its zone-scoped companions
// (DNSSEC key-signing key + signing toggle, query-logging config).
//
// Individual DNS records are NOT created here — each record is its own
// AwsRoute53DnsRecord resource composing onto this zone's zone_id output.
func Resources(ctx *pulumi.Context, iacInput *awsroute53zonev1alpha1.AwsRoute53ZoneIacInput) error {
	locals := initializeLocals(ctx, iacInput)
	spec := locals.AwsRoute53Zone.Spec

	// Build the AWS provider from the IaC input via the shared builder, which
	// resolves the right credential mechanism (static keys, keyless web
	// identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	zoneArgs := &route53.ZoneArgs{
		// metadata.name IS the domain (ForceNew — a zone cannot be renamed).
		Name:         pulumi.String(locals.ZoneName),
		ForceDestroy: pulumi.Bool(spec.ForceDestroy),
		Tags:         pulumi.ToStringMap(locals.AwsTags),
	}
	if spec.Comment != "" {
		zoneArgs.Comment = pulumi.String(spec.Comment)
	}
	// Reusable delegation sets are public-zone-only and conflict with the vpc
	// block — both couplings are CEL-enforced in the spec.
	if spec.DelegationSetId != "" {
		zoneArgs.DelegationSetId = pulumi.String(spec.DelegationSetId)
	}
	// Tri-state pass-through: AWS keeps the feature's current state when the
	// argument is absent and requires an EXPLICIT false to switch it back
	// off, so the spec's presence travels as-is.
	if spec.EnableAcceleratedRecovery != nil {
		zoneArgs.EnableAcceleratedRecovery = pulumi.Bool(spec.GetEnableAcceleratedRecovery())
	}

	// A private zone is defined by its VPC set. AWS creates the zone attached
	// to the first VPC and associates the rest; the provider manages the whole
	// set declaratively. vpc_region defaults to the zone's region so
	// single-region graphs never have to repeat it.
	if spec.IsPrivate {
		vpcs := route53.ZoneVpcArray{}
		for _, association := range spec.VpcAssociations {
			vpcRegion := association.VpcRegion
			if vpcRegion == "" {
				vpcRegion = spec.Region
			}
			vpcs = append(vpcs, &route53.ZoneVpcArgs{
				VpcId:     pulumi.String(association.VpcId.GetValue()),
				VpcRegion: pulumi.String(vpcRegion),
			})
		}
		zoneArgs.Vpcs = vpcs
	}

	createdZone, err := route53.NewZone(ctx, locals.ResourceName, zoneArgs, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrapf(err, "failed to create hosted zone for %s", locals.ZoneName)
	}

	// DNSSEC signing: the key-signing key must exist before the zone's signing
	// status flips to SIGNING — an explicit dependency, not just ordering.
	// The KMS key requirements (us-east-1, ECC_NIST_P256, the
	// dnssec-route53.amazonaws.com key policy) are documented on the spec.
	// The DNSSEC chain-of-trust outputs are empty unless signing is enabled.
	dsRecord := pulumi.String("").ToStringOutput()
	dnskeyRecord := pulumi.String("").ToStringOutput()
	keySigningKeyTag := pulumi.String("").ToStringOutput()
	if spec.Dnssec != nil {
		kskName := spec.Dnssec.KeySigningKeyName
		if kskName == "" {
			kskName = locals.ResourceName + "-ksk"
		}
		// ACTIVE unless the spec deactivates the key (the diagnostics
		// lever); spec default and provider default agree, so this is
		// always sent.
		kskStatus := spec.Dnssec.KeySigningKeyStatus
		if kskStatus == "" {
			kskStatus = "ACTIVE"
		}
		createdKsk, err := route53.NewKeySigningKey(ctx,
			fmt.Sprintf("%s-ksk", locals.ResourceName),
			&route53.KeySigningKeyArgs{
				HostedZoneId:            createdZone.ZoneId,
				KeyManagementServiceArn: pulumi.String(spec.Dnssec.KmsKeyArn.GetValue()),
				Name:                    pulumi.String(kskName),
				Status:                  pulumi.String(kskStatus),
			}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to create DNSSEC key-signing key")
		}

		_, err = route53.NewHostedZoneDnsSec(ctx,
			fmt.Sprintf("%s-dnssec", locals.ResourceName),
			&route53.HostedZoneDnsSecArgs{
				HostedZoneId:  createdZone.ZoneId,
				SigningStatus: pulumi.String("SIGNING"),
			}, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{createdKsk}))
		if err != nil {
			return errors.Wrap(err, "failed to enable DNSSEC signing")
		}

		dsRecord = createdKsk.DsRecord
		dnskeyRecord = createdKsk.DnskeyRecord
		keySigningKeyTag = createdKsk.KeyTag.ApplyT(func(tag int) string {
			return strconv.Itoa(tag)
		}).(pulumi.StringOutput)
	}

	// Query logging: the log group must live in us-east-1 and carry a
	// CloudWatch Logs resource policy allowing route53.amazonaws.com — both
	// account-side prerequisites documented on the spec (the resource policy
	// is account-scoped and deliberately NOT created per zone).
	if spec.QueryLogging != nil {
		_, err = route53.NewQueryLog(ctx,
			fmt.Sprintf("%s-query-log", locals.ResourceName),
			&route53.QueryLogArgs{
				CloudwatchLogGroupArn: pulumi.String(spec.QueryLogging.CloudwatchLogGroupArn.GetValue()),
				ZoneId:                createdZone.ZoneId,
			}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrap(err, "failed to enable query logging")
		}
	}

	ctx.Export(OpZoneId, createdZone.ZoneId)
	ctx.Export(OpZoneName, createdZone.Name)
	ctx.Export(OpNameservers, createdZone.NameServers)
	ctx.Export(OpPrimaryNameServer, createdZone.PrimaryNameServer)
	ctx.Export(OpZoneArn, createdZone.Arn)
	ctx.Export(OpDsRecord, dsRecord)
	ctx.Export(OpDnskeyRecord, dnskeyRecord)
	ctx.Export(OpKeySigningKeyTag, keySigningKeyTag)

	return nil
}
