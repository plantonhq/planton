package module

import (
	"strconv"

	awssagemakerimagev1alpha1 "github.com/plantonhq/planton/catalog/aws/awssagemakerimage/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awssagemakerimagev1alpha1.AwsSagemakerImage
	Spec   *awssagemakerimagev1alpha1.AwsSagemakerImageSpec

	ImageName string
	AwsTags   map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awssagemakerimagev1alpha1.AwsSagemakerImageIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// The component's name IS the image name.
	locals.ImageName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSagemakerImage.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
