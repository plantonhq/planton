package module

import (
	"strconv"

	awseksaddonv1alpha1 "github.com/plantonhq/planton/catalog/aws/awseksaddon/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsEksAddon *awseksaddonv1alpha1.AwsEksAddon

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awseksaddonv1alpha1.AwsEksAddonIacInput) *Locals {
	locals := &Locals{}
	locals.AwsEksAddon = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEksAddon.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
