package module

import (
	"strconv"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"

	awscloudwatchloggroupv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awscloudwatchloggroup/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsCloudwatchLogGroup *awscloudwatchloggroupv1.AwsCloudwatchLogGroup
	AwsTags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awscloudwatchloggroupv1.AwsCloudwatchLogGroupStackInput) *Locals {
	locals := &Locals{}
	locals.AwsCloudwatchLogGroup = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsCloudwatchLogGroup.Metadata.Org,
		awstagkeys.Environment:  locals.AwsCloudwatchLogGroup.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsCloudwatchLogGroup.String(),
		awstagkeys.ResourceId:   locals.AwsCloudwatchLogGroup.Metadata.Id,
	}

	return locals
}
