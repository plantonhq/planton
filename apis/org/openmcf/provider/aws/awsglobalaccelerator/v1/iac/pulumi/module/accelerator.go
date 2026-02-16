package module

import (
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/globalaccelerator"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// accelerator creates the AWS Global Accelerator resource.
func accelerator(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*globalaccelerator.Accelerator, error) {
	spec := locals.GlobalAccelerator.Spec
	name := locals.GlobalAccelerator.Metadata.Name

	args := &globalaccelerator.AcceleratorArgs{
		Name: pulumi.String(name),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	if spec.Enabled != nil {
		args.Enabled = pulumi.Bool(spec.GetEnabled())
	}

	if spec.IpAddressType != nil {
		args.IpAddressType = pulumi.StringPtr(spec.GetIpAddressType())
	}

	if len(spec.IpAddresses) > 0 {
		args.IpAddresses = pulumi.ToStringArray(spec.IpAddresses)
	}

	if spec.FlowLogs != nil && spec.FlowLogs.Enabled {
		args.Attributes = globalaccelerator.AcceleratorAttributesArgs{
			FlowLogsEnabled:  pulumi.Bool(true),
			FlowLogsS3Bucket: pulumi.String(spec.FlowLogs.S3Bucket.GetValue()),
			FlowLogsS3Prefix: pulumi.String(spec.FlowLogs.S3Prefix),
		}.ToAcceleratorAttributesPtrOutput().Elem().ToAcceleratorAttributesPtrOutput()
	}

	accel, err := globalaccelerator.NewAccelerator(ctx, name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, err
	}

	return accel, nil
}
