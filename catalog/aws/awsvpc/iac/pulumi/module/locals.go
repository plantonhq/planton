package module

import (
	"strconv"

	awsvpcv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsvpc/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsVpc  *awsvpcv1alpha1.AwsVpc
	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsvpcv1alpha1.AwsVpcIacInput) *Locals {
	locals := &Locals{}
	locals.AwsVpc = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsVpc.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
