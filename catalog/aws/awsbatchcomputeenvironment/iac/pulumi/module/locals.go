package module

import (
	"strconv"

	awsbatchcomputeenvironmentv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbatchcomputeenvironment/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	AwsBatchComputeEnvironment *awsbatchcomputeenvironmentv1alpha1.AwsBatchComputeEnvironment
	AwsTags                    map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsbatchcomputeenvironmentv1alpha1.AwsBatchComputeEnvironmentIacInput) *Locals {
	locals := &Locals{}
	locals.AwsBatchComputeEnvironment = iacInput.Target

	// Resource-identity tags follow the catalog convention. These land on the
	// compute environment itself; tags for the EC2 instances Batch launches
	// are a separate concern carried by spec.compute_resources.resource_tags.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsBatchComputeEnvironment.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsBatchComputeEnvironment.Metadata.Org,
		awstagkeys.Environment:  locals.AwsBatchComputeEnvironment.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBatchComputeEnvironment.String(),
		awstagkeys.ResourceId:   locals.AwsBatchComputeEnvironment.Metadata.Id,
	}

	return locals
}
