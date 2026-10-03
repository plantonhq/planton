package module

import (
	"strconv"

	awsgluecatalogdatabase "github.com/plantonhq/planton/catalog/aws/awsgluecatalogdatabase/v1alpha1"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/plantonhq/planton/shared/catalogkind"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the IaC input.
type Locals struct {
	Target       *awsgluecatalogdatabase.AwsGlueCatalogDatabase
	Spec         *awsgluecatalogdatabase.AwsGlueCatalogDatabaseSpec
	AwsTags      map[string]string
	DatabaseName string
}

func initializeLocals(ctx *pulumi.Context, in *awsgluecatalogdatabase.AwsGlueCatalogDatabaseIacInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	locals.DatabaseName = in.Target.Metadata.Name

	// Canonical six-key resource-identity map, matching the Terraform module
	// key-for-key (Name + the planton.ai identity keys).
	locals.AwsTags = map[string]string{
		awstagkeys.Name:         locals.Target.Metadata.Name,
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: catalogkind.CatalogKind_AwsGlueCatalogDatabase.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
