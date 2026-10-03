package module

import (
	"strconv"

	awssagemakermodelv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssagemakermodel/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awssagemakermodelv1alpha1.AwsSagemakerModel
	Spec   *awssagemakermodelv1alpha1.AwsSagemakerModelSpec

	ModelName string
	AwsTags   map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awssagemakermodelv1alpha1.AwsSagemakerModelIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// The component's name IS the model name (charset-compatible).
	locals.ModelName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSagemakerModel.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
