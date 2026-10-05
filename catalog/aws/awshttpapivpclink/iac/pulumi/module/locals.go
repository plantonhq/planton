package module

import (
	"strconv"

	awshttpapivpclinkv1alpha1 "github.com/plantonhq/planton/catalog/aws/awshttpapivpclink/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set applied to the VPC link.
type Locals struct {
	AwsHttpApiVpcLink *awshttpapivpclinkv1alpha1.AwsHttpApiVpcLink
	AwsTags           map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awshttpapivpclinkv1alpha1.AwsHttpApiVpcLinkIacInput) *Locals {
	locals := &Locals{}
	locals.AwsHttpApiVpcLink = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsHttpApiVpcLink.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsHttpApiVpcLink.Metadata.Org,
		awstagkeys.Environment:  locals.AwsHttpApiVpcLink.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsHttpApiVpcLink.String(),
		awstagkeys.ResourceId:   locals.AwsHttpApiVpcLink.Metadata.Id,
	}

	return locals
}
