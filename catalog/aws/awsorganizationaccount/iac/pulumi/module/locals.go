package module

import (
	"strconv"

	awsorganizationaccountv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsorganizationaccount/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsorganizationaccountv1alpha1.AwsOrganizationAccount
	Spec   *awsorganizationaccountv1alpha1.AwsOrganizationAccountSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsorganizationaccountv1alpha1.AwsOrganizationAccountIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	// The account resource is the kind's one taggable surface (the
	// contact and region satellites are untaggable).
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsOrganizationAccount.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
