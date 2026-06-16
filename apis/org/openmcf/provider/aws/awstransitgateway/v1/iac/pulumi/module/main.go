package module

import (
	"github.com/pkg/errors"
	awstgwv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awstransitgateway/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the AwsTransitGateway Pulumi
// module. It creates the Transit Gateway, attaches VPCs, and exports
// all outputs for downstream consumption.
func Resources(ctx *pulumi.Context, stackInput *awstgwv1.AwsTransitGatewayStackInput) error {
	locals := initializeLocals(ctx, stackInput)

	// Build the AWS provider from the stack input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, stackInput.ProviderConfig, locals.TransitGateway.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
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
