package module

import (
	"strconv"

	"github.com/plantonhq/planton/shared/catalogkind"

	iamrolev1 "github.com/plantonhq/planton/catalog/aws/awsiamrole/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	AwsIamRole *iamrolev1.AwsIamRole
	AwsTags    map[string]string
}

func initializeLocals(ctx *pulumi.Context, iacInput *iamrolev1.AwsIamRoleIacInput) *Locals {
	locals := &Locals{}
	locals.AwsIamRole = iacInput.Target

	// Resource-identity tags match the Terraform module key-for-key
	// (Name plus the planton.ai identity keys).
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.AwsIamRole.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.AwsIamRole.Metadata.Org,
		awstagkeys.Environment:  locals.AwsIamRole.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsIamRole.String(),
		awstagkeys.ResourceId:   locals.AwsIamRole.Metadata.Id,
	}

	return locals
}
