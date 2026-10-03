package module

import (
	"strconv"

	awsbatchjobdefinitionv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbatchjobdefinition/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	AwsBatchJobDefinition *awsbatchjobdefinitionv1alpha1.AwsBatchJobDefinition
	AwsTags               map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsbatchjobdefinitionv1alpha1.AwsBatchJobDefinitionIacInput) *Locals {
	locals := &Locals{}
	locals.AwsBatchJobDefinition = iacInput.Target

	// Resource-identity tags follow the catalog convention. With
	// spec.propagate_tags they also reach the ECS tasks jobs run as.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsBatchJobDefinition.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsBatchJobDefinition.Metadata.Org,
		awstagkeys.Environment:  locals.AwsBatchJobDefinition.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBatchJobDefinition.String(),
		awstagkeys.ResourceId:   locals.AwsBatchJobDefinition.Metadata.Id,
	}

	return locals
}
