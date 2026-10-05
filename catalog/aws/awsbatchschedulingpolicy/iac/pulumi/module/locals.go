package module

import (
	"strconv"

	awsbatchschedulingpolicyv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsbatchschedulingpolicy/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	AwsBatchSchedulingPolicy *awsbatchschedulingpolicyv1alpha1.AwsBatchSchedulingPolicy
	AwsTags                  map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *awsbatchschedulingpolicyv1alpha1.AwsBatchSchedulingPolicyIacInput) *Locals {
	locals := &Locals{}
	locals.AwsBatchSchedulingPolicy = iacInput.Target

	// Resource-identity tags follow the catalog convention.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsBatchSchedulingPolicy.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsBatchSchedulingPolicy.Metadata.Org,
		awstagkeys.Environment:  locals.AwsBatchSchedulingPolicy.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsBatchSchedulingPolicy.String(),
		awstagkeys.ResourceId:   locals.AwsBatchSchedulingPolicy.Metadata.Id,
	}

	return locals
}
