package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awscloudwatchloggroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudwatchloggroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsCloudwatchLogGroup *awscloudwatchloggroupv1alpha1.AwsCloudwatchLogGroup
	AwsTags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awscloudwatchloggroupv1alpha1.AwsCloudwatchLogGroupIacInput) *Locals {
	locals := &Locals{}
	locals.AwsCloudwatchLogGroup = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsCloudwatchLogGroup.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsCloudwatchLogGroup.Metadata.Org,
		awstagkeys.Environment:  locals.AwsCloudwatchLogGroup.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsCloudwatchLogGroup.String(),
		awstagkeys.ResourceId:   locals.AwsCloudwatchLogGroup.Metadata.Id,
	}

	return locals
}
