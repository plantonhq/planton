package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type EipResult struct {
	AllocationId pulumi.StringOutput
	PublicIp     pulumi.StringOutput
	Arn          pulumi.StringOutput
	PublicDns    pulumi.StringOutput
}

func eip(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*EipResult, error) {
	spec := locals.AwsElasticIp.Spec

	args := &ec2.EipArgs{
		Domain: pulumi.String("vpc"),
		Tags:   pulumi.ToStringMap(locals.AwsTags),
	}

	// BYOIP: allocate from a specific IPv4 address pool.
	if spec.PublicIpv4Pool != "" {
		args.PublicIpv4Pool = pulumi.StringPtr(spec.PublicIpv4Pool)
	}

	// BYOIP: request a specific IP address from the pool.
	if spec.Address != "" {
		args.Address = pulumi.StringPtr(spec.Address)
	}

	// Location scope for Local Zones and Wavelength zones.
	if spec.NetworkBorderGroup != "" {
		args.NetworkBorderGroup = pulumi.StringPtr(spec.NetworkBorderGroup)
	}

	createdEip, err := ec2.NewEip(ctx, locals.AwsElasticIp.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create elastic ip")
	}

	return &EipResult{
		AllocationId: createdEip.AllocationId,
		PublicIp:     createdEip.PublicIp,
		Arn:          createdEip.Arn,
		PublicDns:    createdEip.PublicDns,
	}, nil
}
