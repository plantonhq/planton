package module

import (
	"strconv"

	awstgwattachv1 "github.com/plantonhq/planton/catalog/aws/awstransitgatewayvpcattachment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set applied to the attachment.
type Locals struct {
	VpcAttachment *awstgwattachv1.AwsTransitGatewayVpcAttachment
	AwsTags       map[string]string
}

// initializeLocals reads the IaC input and builds the Locals instance.
func initializeLocals(ctx *pulumi.Context, iacInput *awstgwattachv1.AwsTransitGatewayVpcAttachmentIacInput) *Locals {
	locals := &Locals{}

	locals.VpcAttachment = iacInput.Target

	// Identity tags match the Terraform module key-for-key. The Name tag IS
	// the attachment's console identity -- attachments have no name
	// attribute of their own.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.VpcAttachment.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.VpcAttachment.Metadata.Org,
		awstagkeys.Environment:  locals.VpcAttachment.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsTransitGatewayVpcAttachment.String(),
		awstagkeys.ResourceId:   locals.VpcAttachment.Metadata.Id,
	}

	return locals
}

// enableDisable converts a plain boolean to the AWS-style "enable"/"disable"
// string the Transit Gateway API speaks.
func enableDisable(b bool) string {
	if b {
		return "enable"
	}
	return "disable"
}

// enableDisableTriState converts a proto `optional bool` to the AWS-style
// string, preserving the tri-state: nil (unset in the manifest) returns nil
// so the argument is omitted and the provider/AWS default applies, exactly
// like the Terraform module's null fall-through.
func enableDisableTriState(b *bool) pulumi.StringPtrInput {
	if b == nil {
		return nil
	}
	return pulumi.StringPtr(enableDisable(*b))
}
