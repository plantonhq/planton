package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awsnlbv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsnlb/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the NLB resource definition from the IaC input and a map of
// AWS tags to apply to all created resources.
type Locals struct {
	Nlb     *awsnlbv1alpha1.AwsNlb
	AwsTags map[string]string
}

// initializeLocals reads the IaC input and builds the Locals instance,
// analogous to a Terraform locals block.
func initializeLocals(ctx *pulumi.Context, iacInput *awsnlbv1alpha1.AwsNlbIacInput) *Locals {
	locals := &Locals{}

	locals.Nlb = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.Nlb.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Nlb.Metadata.Org,
		awstagkeys.Environment:  locals.Nlb.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsNlb.String(),
		awstagkeys.ResourceId:   locals.Nlb.Metadata.Id,
	}

	return locals
}
