package module

import (
	"strconv"

	awslbtargetgroupv1alpha1 "github.com/plantonhq/planton/catalog/aws/awslbtargetgroup/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsLbTargetGroup *awslbtargetgroupv1alpha1.AwsLbTargetGroup
	AwsTags          map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awslbtargetgroupv1alpha1.AwsLbTargetGroupIacInput) *Locals {
	locals := &Locals{}
	locals.AwsLbTargetGroup = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsLbTargetGroup.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
