package module

import (
	"strconv"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"

	awskmskeyv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awskmskey/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsKmsKey *awskmskeyv1.AwsKmsKey
	AwsTags   map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awskmskeyv1.AwsKmsKeyStackInput) *Locals {
	locals := &Locals{}
	locals.AwsKmsKey = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsKmsKey.Metadata.Org,
		awstagkeys.Environment:  locals.AwsKmsKey.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsKmsKey.String(),
		awstagkeys.ResourceId:   locals.AwsKmsKey.Metadata.Id,
	}

	return locals
}
