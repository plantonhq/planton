package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	awswafwebaclv1alpha1 "github.com/plantonhq/planton/catalog/aws/awswafwebacl/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds the Web ACL resource definition from the IaC input and a map
// of AWS tags to apply to all created resources.
type Locals struct {
	WebAcl  *awswafwebaclv1alpha1.AwsWafWebAcl
	AwsTags map[string]string
}

// initializeLocals reads the IaC input and builds the Locals instance,
// analogous to a Terraform locals block.
func initializeLocals(ctx *pulumi.Context, iacInput *awswafwebaclv1alpha1.AwsWafWebAclIacInput) *Locals {
	locals := &Locals{}

	locals.WebAcl = iacInput.Target

	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.WebAcl.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.WebAcl.Metadata.Org,
		awstagkeys.Environment:  locals.WebAcl.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsWafWebAcl.String(),
		awstagkeys.ResourceId:   locals.WebAcl.Metadata.Id,
	}

	return locals
}
