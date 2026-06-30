package module

import (
	"strconv"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"

	awsalbv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsalb/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the AWS ALB resource definition from the stack input
// and a map of AWS tags to apply to resources.
type Locals struct {
	AwsAlb  *awsalbv1.AwsAlb
	AwsTags map[string]string
}

// initializeLocals is analogous to Terraform "locals." It reads
// values from AwsAlbStackInput to build a Locals instance.
func initializeLocals(ctx *pulumi.Context, stackInput *awsalbv1.AwsAlbStackInput) *Locals {
	locals := &Locals{}

	locals.AwsAlb = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsAlb.Metadata.Org,
		awstagkeys.Environment:  locals.AwsAlb.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsAlb.String(),
		awstagkeys.ResourceId:   locals.AwsAlb.Metadata.Id,
	}

	return locals
}
