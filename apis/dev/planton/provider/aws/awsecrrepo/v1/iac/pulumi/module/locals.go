package module

import (
	"strconv"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"

	awsecrrepov1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsecrrepo/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the AWS ECR Repo resource definition from the stack input
// and a map of AWS tags to apply to resources.
type Locals struct {
	AwsEcrRepo *awsecrrepov1.AwsEcrRepo
	AwsTags    map[string]string
}

// initializeLocals is similar to Terraform "locals" usage. It reads
// values from AwsEcrRepoStackInput to build a Locals instance.
func initializeLocals(ctx *pulumi.Context, stackInput *awsecrrepov1.AwsEcrRepoStackInput) *Locals {
	locals := &Locals{}

	locals.AwsEcrRepo = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsEcrRepo.Metadata.Org,
		awstagkeys.Environment:  locals.AwsEcrRepo.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsEcrRepo.String(),
		awstagkeys.ResourceId:   locals.AwsEcrRepo.Metadata.Id,
	}

	return locals
}
