package module

import (
	"github.com/pkg/errors"
	awstgwv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awstransitgateway/v1"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the AwsTransitGateway Pulumi
// module. It creates the Transit Gateway, attaches VPCs, and exports
// all outputs for downstream consumption.
func Resources(ctx *pulumi.Context, stackInput *awstgwv1.AwsTransitGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	var provider *aws.Provider
	var err error
	awsProviderConfig := stackInput.ProviderConfig

	if awsProviderConfig == nil {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			Region: pulumi.String(locals.TransitGateway.Spec.Region),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create default AWS provider")
		}
	} else {
		provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
			AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
			SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
			Region:    pulumi.String(locals.TransitGateway.Spec.Region),
			Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
		})
		if err != nil {
			return errors.Wrap(err, "failed to create AWS provider with custom credentials")
		}
	}

	tgwResult, err := transitGateway(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create transit gateway")
	}

	attachmentResult, err := vpcAttachments(ctx, locals, provider, tgwResult.TransitGateway)
	if err != nil {
		return errors.Wrap(err, "failed to create vpc attachments")
	}

	// Export TGW-level outputs.
	ctx.Export(OpTransitGatewayId, tgwResult.TransitGateway.ID())
	ctx.Export(OpTransitGatewayArn, tgwResult.TransitGateway.Arn)
	ctx.Export(OpOwnerId, tgwResult.TransitGateway.OwnerId)
	ctx.Export(OpAssociationDefaultRouteTableId, tgwResult.TransitGateway.AssociationDefaultRouteTableId)
	ctx.Export(OpPropagationDefaultRouteTableId, tgwResult.TransitGateway.PropagationDefaultRouteTableId)

	// Build and export the VPC attachment ID map.
	attachmentIdMap := pulumi.StringMap{}
	for name, att := range attachmentResult.Attachments {
		attachmentIdMap[name] = att.ID().ToStringOutput()
	}
	ctx.Export(OpVpcAttachmentIds, attachmentIdMap)

	return nil
}
