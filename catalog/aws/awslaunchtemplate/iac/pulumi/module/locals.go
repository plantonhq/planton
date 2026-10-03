package module

import (
	"strconv"

	awslaunchtemplatev1alpha1 "github.com/plantonhq/planton/catalog/aws/awslaunchtemplate/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsLaunchTemplate *awslaunchtemplatev1alpha1.AwsLaunchTemplate
	AwsTags           map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awslaunchtemplatev1alpha1.AwsLaunchTemplateIacInput) *Locals {
	locals := &Locals{}
	locals.AwsLaunchTemplate = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsLaunchTemplate.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
