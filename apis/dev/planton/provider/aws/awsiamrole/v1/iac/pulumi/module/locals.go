package module

import (
	"strconv"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"

	iamrolev1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsiamrole/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsIamRole *iamrolev1.AwsIamRole
	AwsTags    map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *iamrolev1.AwsIamRoleStackInput) *Locals {
	locals := &Locals{}
	locals.AwsIamRole = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsIamRole.Metadata.Org,
		awstagkeys.Environment:  locals.AwsIamRole.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsIamRole.String(),
		awstagkeys.ResourceId:   locals.AwsIamRole.Metadata.Id,
	}

	return locals
}
