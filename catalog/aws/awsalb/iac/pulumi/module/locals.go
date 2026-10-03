package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awsalbv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsalb/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the AWS ALB resource definition from the IaC input
// and a map of AWS tags to apply to resources.
type Locals struct {
	AwsAlb  *awsalbv1alpha1.AwsAlb
	AwsTags map[string]string
}

// initializeLocals is analogous to Terraform "locals." It reads
// values from AwsAlbIacInput to build a Locals instance.
func initializeLocals(ctx *pulumi.Context, iacInput *awsalbv1alpha1.AwsAlbIacInput) *Locals {
	locals := &Locals{}

	locals.AwsAlb = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsAlb.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsAlb.Metadata.Org,
		awstagkeys.Environment:  locals.AwsAlb.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsAlb.String(),
		awstagkeys.ResourceId:   locals.AwsAlb.Metadata.Id,
	}

	return locals
}
