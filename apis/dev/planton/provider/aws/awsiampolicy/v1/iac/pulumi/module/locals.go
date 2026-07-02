package module

import (
	"strconv"

	awsiampolicyv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsiampolicy/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsIamPolicy *awsiampolicyv1.AwsIamPolicy
	AwsTags      map[string]string
}

func initializeLocals(_ *pulumi.Context, stackInput *awsiampolicyv1.AwsIamPolicyStackInput) *Locals {
	locals := &Locals{}
	locals.AwsIamPolicy = stackInput.Target

	metadata := stackInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsIamPolicy.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
