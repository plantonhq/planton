package module

import (
	"strconv"

	"github.com/plantonhq/openmcf/apis/org/openmcf/shared/cloudresourcekind"

	awscloudwatchalarmv1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/aws/awscloudwatchalarm/v1"
	"github.com/plantonhq/openmcf/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
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
