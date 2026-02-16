package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2transitgateway"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// transitGatewayResult holds the outputs from TGW creation needed by
// downstream resources (VPC attachments, output exports).
type transitGatewayResult struct {
	TransitGateway *ec2transitgateway.TransitGateway
}

// transitGateway creates the AWS Transit Gateway resource from the spec's
// core configuration and feature toggles.
func transitGateway(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) (*transitGatewayResult, error) {
	spec := locals.TransitGateway.Spec

	args := &ec2transitgateway.TransitGatewayArgs{
		AutoAcceptSharedAttachments:     pulumi.StringPtr(enableDisable(spec.AutoAcceptSharedAttachments)),
		DefaultRouteTableAssociation:    pulumi.StringPtr(enableDisable(spec.DefaultRouteTableAssociation)),
		DefaultRouteTablePropagation:    pulumi.StringPtr(enableDisable(spec.DefaultRouteTablePropagation)),
		DnsSupport:                      pulumi.StringPtr(enableDisable(spec.DnsSupport)),
		VpnEcmpSupport:                  pulumi.StringPtr(enableDisable(spec.VpnEcmpSupport)),
		SecurityGroupReferencingSupport: pulumi.StringPtr(enableDisable(spec.SecurityGroupReferencingSupport)),
		MulticastSupport:                pulumi.StringPtr(enableDisable(spec.MulticastSupport)),
		Tags:                            pulumi.ToStringMap(locals.AwsTags),
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	if spec.AmazonSideAsn != 0 {
		args.AmazonSideAsn = pulumi.IntPtr(int(spec.AmazonSideAsn))
	}

	if len(spec.TransitGatewayCidrBlocks) > 0 {
		args.TransitGatewayCidrBlocks = pulumi.ToStringArray(spec.TransitGatewayCidrBlocks)
	}

	createdTgw, err := ec2transitgateway.NewTransitGateway(
		ctx,
		locals.TransitGateway.Metadata.Name,
		args,
		pulumi.Provider(provider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create transit gateway")
	}

	return &transitGatewayResult{TransitGateway: createdTgw}, nil
}
