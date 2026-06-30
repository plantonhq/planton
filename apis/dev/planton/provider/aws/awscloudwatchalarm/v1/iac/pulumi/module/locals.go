package module

import (
	"strconv"

	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"

	awscloudwatchalarmv1 "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awscloudwatchalarm/v1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsCloudwatchAlarm *awscloudwatchalarmv1.AwsCloudwatchAlarm
	AwsTags            map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *awscloudwatchalarmv1.AwsCloudwatchAlarmStackInput) *Locals {
	locals := &Locals{}
	locals.AwsCloudwatchAlarm = stackInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsCloudwatchAlarm.Metadata.Org,
		awstagkeys.Environment:  locals.AwsCloudwatchAlarm.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsCloudwatchAlarm.String(),
		awstagkeys.ResourceId:   locals.AwsCloudwatchAlarm.Metadata.Id,
	}

	return locals
}
