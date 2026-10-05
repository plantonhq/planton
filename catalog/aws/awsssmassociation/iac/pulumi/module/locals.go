package module

import (
	"strconv"

	awsssmassociationv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsssmassociation/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsssmassociationv1alpha1.AwsSsmAssociation
	Spec   *awsssmassociationv1alpha1.AwsSsmAssociationSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsssmassociationv1alpha1.AwsSsmAssociationIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsSsmAssociation.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
