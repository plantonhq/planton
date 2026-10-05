package module

import (
	"github.com/pkg/errors"
	awstgwattachv1 "github.com/plantonhq/planton/catalog/aws/awstransitgatewayvpcattachment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/pulumiawsprovider"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Resources is the primary entry point for the AwsTransitGatewayVpcAttachment
// Pulumi module. It creates the attachment and exports its outputs -- the
// attachment ID is the join key Transit Gateway route tables associate,
// propagate, and route against.
func Resources(ctx *pulumi.Context, iacInput *awstgwattachv1.AwsTransitGatewayVpcAttachmentIacInput) error {
	locals := initializeLocals(ctx, iacInput)

	// Build the AWS provider from the IaC input via the shared builder, which resolves
	// the right credential mechanism (static keys, keyless web identity, or ambient chain).
	provider, err := pulumiawsprovider.Get(ctx, iacInput.ProviderConfig, locals.VpcAttachment.Spec.Region)
	if err != nil {
		return errors.Wrap(err, "failed to create AWS provider")
	}

	createdVpcAttachment, err := vpcAttachment(ctx, locals, provider)
	if err != nil {
		return errors.Wrap(err, "failed to create transit gateway vpc attachment")
	}

	ctx.Export(OpAttachmentId, createdVpcAttachment.ID())
	ctx.Export(OpAttachmentArn, createdVpcAttachment.Arn)
	ctx.Export(OpVpcOwnerId, createdVpcAttachment.VpcOwnerId)

	return nil
}
