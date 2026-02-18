package module

import (
	"github.com/pkg/errors"
	awsnlbv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsnetworkloadbalancer/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the AwsNetworkLoadBalancer Pulumi
// module. It creates the NLB, listeners with target groups, and optional DNS.
func Resources(ctx *pulumi.Context, stackInput *awsnlbv1.AwsNetworkLoadBalancerStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.Nlb.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.Nlb.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	nlbResource, err := nlb(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create Network Load Balancer")
	}

	if err := listeners(ctx, locals, provider, nlbResource); err != nil {
		return errors.Wrap(err, "failed to create listeners and target groups")
	}

	if locals.Nlb.Spec.Dns != nil && locals.Nlb.Spec.Dns.Enabled {
		if err := dns(ctx, locals, provider, nlbResource); err != nil {
			return errors.Wrap(err, "failed to configure DNS")
		}
	}

	return nil
}
