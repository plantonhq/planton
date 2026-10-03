package module

import (
	"strconv"

	awssagemakermodelregistryv1alpha1 "github.com/plantonhq/planton/catalog/aws/awssagemakermodelregistry/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awssagemakermodelregistryv1alpha1.AwsSagemakerModelRegistry
	Spec   *awssagemakermodelregistryv1alpha1.AwsSagemakerModelRegistrySpec

	GroupName string
	AwsTags   map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awssagemakermodelregistryv1alpha1.AwsSagemakerModelRegistryIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// The component's name IS the group name.
	locals.GroupName = metadata.Name

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSagemakerModelRegistry.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
