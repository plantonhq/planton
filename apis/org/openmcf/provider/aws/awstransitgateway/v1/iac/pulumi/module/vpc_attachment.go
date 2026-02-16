package module

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/ec2transitgateway"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// vpcAttachmentResult holds the map of attachment name to Pulumi resource,
// used by the main function to build output maps.
type vpcAttachmentResult struct {
	Attachments map[string]*ec2transitgateway.VpcAttachment
}

// vpcAttachments creates a Transit Gateway VPC attachment for each entry
// in spec.vpc_attachments. Each attachment connects one VPC to the TGW
// through the specified subnets.
func vpcAttachments(
	ctx *pulumi.Context,
	locals *Locals,
	provider *aws.Provider,
	tgw *ec2transitgateway.TransitGateway,
) (*vpcAttachmentResult, error) {
	result := &vpcAttachmentResult{
		Attachments: make(map[string]*ec2transitgateway.VpcAttachment),
	}

	for _, attachment := range locals.TransitGateway.Spec.VpcAttachments {
		subnetIds := make([]string, 0, len(attachment.SubnetIds))
		for _, s := range attachment.SubnetIds {
			subnetIds = append(subnetIds, s.GetValue())
		}

		resourceName := fmt.Sprintf("%s-%s", locals.TransitGateway.Metadata.Name, attachment.Name)

		args := &ec2transitgateway.VpcAttachmentArgs{
			TransitGatewayId:                          tgw.ID(),
			VpcId:                                     pulumi.String(attachment.VpcId.GetValue()),
			SubnetIds:                                 pulumi.ToStringArray(subnetIds),
			DnsSupport:                                pulumi.StringPtr(enableDisable(attachment.DnsSupport)),
			Ipv6Support:                               pulumi.StringPtr(enableDisable(attachment.Ipv6Support)),
			ApplianceModeSupport:                      pulumi.StringPtr(enableDisable(attachment.ApplianceModeSupport)),
			TransitGatewayDefaultRouteTableAssociation: pulumi.BoolPtr(attachment.DefaultRouteTableAssociation),
			TransitGatewayDefaultRouteTablePropagation: pulumi.BoolPtr(attachment.DefaultRouteTablePropagation),
			Tags:                                      pulumi.ToStringMap(locals.AwsTags),
		}

		createdAttachment, err := ec2transitgateway.NewVpcAttachment(
			ctx,
			resourceName,
			args,
			pulumi.Provider(provider),
			pulumi.DependsOn([]pulumi.Resource{tgw}),
		)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to create vpc attachment %q", attachment.Name)
		}

		result.Attachments[attachment.Name] = createdAttachment
	}

	return result, nil
}
