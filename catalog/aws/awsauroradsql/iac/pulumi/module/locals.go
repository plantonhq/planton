package module

import (
	"strconv"

	awsauroradsqlv1alpha1 "github.com/plantonhq/planton/catalog/aws/awsauroradsql/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target *awsauroradsqlv1alpha1.AwsAuroraDsql
	Spec   *awsauroradsqlv1alpha1.AwsAuroraDsqlSpec

	AwsTags map[string]string
}

func initializeLocals(_ *pulumi.Context, in *awsauroradsqlv1alpha1.AwsAuroraDsqlIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec

	metadata := in.Target.Metadata

	// Resource-identity tags match the Terraform module key-for-key.
	// DSQL generates its own cluster identifier, so the Name tag is
	// how humans find this cluster in the console.
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: metadata.Org,
		awstagkeys.Environment:  metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsAuroraDsql.String(),
		awstagkeys.ResourceId:   metadata.Id,
	}

	return locals
}
