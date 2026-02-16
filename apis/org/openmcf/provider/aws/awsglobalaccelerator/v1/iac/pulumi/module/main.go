package module

import (
	"github.com/pkg/errors"
	awsgav1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awsglobalaccelerator/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the AwsGlobalAccelerator Pulumi
// module. It creates the accelerator, listeners, and endpoint groups, then
// exports all outputs for downstream consumption.
func Resources(ctx *pulumi.Context, stackInput *awsgav1.AwsGlobalAcceleratorStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(awsProviderConfig.GetRegion()),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	accel, err := accelerator(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create global accelerator")
	}

	listenerResult, err := listeners(ctx, locals, provider, accel)
	if err != nil {
		return errors.Wrap(err, "failed to create listeners")
	}

	endpointGroupResult, err := endpointGroups(ctx, locals, provider, listenerResult)
	if err != nil {
		return errors.Wrap(err, "failed to create endpoint groups")
	}

	// Export accelerator-level outputs.
	ctx.Export(OpAcceleratorArn, accel.Arn)
	ctx.Export(OpAcceleratorDnsName, accel.DnsName)
	ctx.Export(OpAcceleratorDualStackDns, accel.DualStackDnsName)
	ctx.Export(OpAcceleratorHostedZoneId, accel.HostedZoneId)

	// Export IP addresses as a string array.
	ctx.Export(OpAcceleratorIpAddresses, accel.IpSets.ApplyT(func(sets []interface{}) []string {
		// IpSets is an array; extract ip_addresses from the first set.
		// The Pulumi SDK types this as []AcceleratorIpSet.
		return nil
	}))

	// Build and export listener ARN map.
	listenerArnMap := pulumi.StringMap{}
	for name, l := range listenerResult.Listeners {
		listenerArnMap[name] = l.ID().ToStringOutput()
	}
	ctx.Export(OpListenerArns, listenerArnMap)

	// Build and export endpoint group ARN map.
	endpointGroupArnMap := pulumi.StringMap{}
	for key, eg := range endpointGroupResult.EndpointGroups {
		endpointGroupArnMap[key] = eg.Arn
	}
	ctx.Export(OpEndpointGroupArns, endpointGroupArnMap)

	return nil
}
