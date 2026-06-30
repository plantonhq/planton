package module

import (
	"strconv"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"

	awsiamuserv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsiamuser/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsIamUser *awsiamuserv1.AwsIamUser
	AwsTags    map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awsiamuserv1.AwsIamUserStackInput) *Locals {
	locals := &Locals{}
	locals.AwsIamUser = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsIamUser.Metadata.Org,
		awstagkeys.Environment:  locals.AwsIamUser.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsIamUser.String(),
		awstagkeys.ResourceId:   locals.AwsIamUser.Metadata.Id,
	}

	return locals
}
