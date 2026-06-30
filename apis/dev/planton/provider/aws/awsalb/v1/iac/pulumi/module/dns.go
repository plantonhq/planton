package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/lb"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/route53"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// dns configures Route53 alias records for the ALB if dns_config.enabled is true.
func dns(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	albResource *lb.LoadBalancer,
) error {
	if locals.AwsAlb.Spec.Dns.Route53ZoneId.GetValue() == "" {
		return errors.New("dns_config.enabled is true but route53_zone_id is not provided")
	}

	if len(locals.AwsAlb.Spec.Dns.Hostnames) == 0 {
		return errors.New("dns_config.enabled is true but no hostnames provided")
	}

	for i, hostname := range locals.AwsAlb.Spec.Dns.Hostnames {
		recordName := fmt.Sprintf("%s-dns-%d", locals.AwsAlb.Metadata.Name, i)

		_, err := route53.NewRecord(ctx, recordName, &route53.RecordArgs{
			// UPSERT instead of CREATE so an alias record left by a prior partial
			// apply (or already pointing at this ALB) is adopted, not collided with.
			AllowOverwrite: pulumi.Bool(true),
			Name:           pulumi.String(hostname),
			Type:           pulumi.String("A"),
			ZoneId:         pulumi.String(locals.AwsAlb.Spec.Dns.Route53ZoneId.GetValue()),
			Aliases: route53.RecordAliasArray{
				route53.RecordAliasArgs{
					Name:                 albResource.DnsName,
					ZoneId:               albResource.ZoneId,
					EvaluateTargetHealth: pulumi.Bool(true),
				},
			},
		}, pulumi.Provider(provider))
		if err != nil {
			return errors.Wrapf(err, "failed to create Route53 record for %s", hostname)
		}
	}

	return nil
}
