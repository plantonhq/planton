package module

import (
	"strconv"

	awsecstaskdefinitionv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsecstaskdefinition/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsEcsTaskDefinition *awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinition

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, iacInput *awsecstaskdefinitionv1alpha1.AwsEcsTaskDefinitionIacInput) *Locals {
	locals := &Locals{}
	locals.AwsEcsTaskDefinition = iacInput.Target

	metadata := iacInput.Target.Metadata
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsEcsTaskDefinition.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
