package module

import (
	"strconv"

	awsbudgetv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbudget/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsbudgetv1alpha1.AwsBudget
	Spec   *awsbudgetv1alpha1.AwsBudgetSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsbudgetv1alpha1.AwsBudgetIacInput) *Locals {
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
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBudget.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
