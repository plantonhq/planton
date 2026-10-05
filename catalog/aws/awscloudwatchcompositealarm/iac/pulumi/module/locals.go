package module

import (
	"strconv"

	awscloudwatchcompositealarmv1alpha1 "github.com/plantonhq/planton/catalog/aws/awscloudwatchcompositealarm/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals mirrors Terraform-style locals: the target resource and the identity
// tag set applied to the composite alarm.
type Locals struct {
	AwsCloudwatchCompositeAlarm *awscloudwatchcompositealarmv1alpha1.AwsCloudwatchCompositeAlarm
	AwsTags                     map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awscloudwatchcompositealarmv1alpha1.AwsCloudwatchCompositeAlarmIacInput) *Locals {
	locals := &Locals{}
	locals.AwsCloudwatchCompositeAlarm = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsCloudwatchCompositeAlarm.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsCloudwatchCompositeAlarm.Metadata.Org,
		awstagkeys.Environment:  locals.AwsCloudwatchCompositeAlarm.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsCloudwatchCompositeAlarm.String(),
		awstagkeys.ResourceId:   locals.AwsCloudwatchCompositeAlarm.Metadata.Id,
	}

	return locals
}
