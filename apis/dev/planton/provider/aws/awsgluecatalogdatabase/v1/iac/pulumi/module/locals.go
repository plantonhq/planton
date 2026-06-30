package module

import (
	"strconv"

	awsgluecatalogdatabase "github.com/plantonhq/planton/apis/dev/planton/provider/aws/awsgluecatalogdatabase/v1"
	"github.com/plantonhq/planton/apis/dev/planton/shared/cloudresourcekind"
	"github.com/plantonhq/planton/pkg/iac/pulumi/pulumimodule/provider/aws/awstagkeys"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Locals holds pre-computed values derived from the stack input.
type Locals struct {
	Target       *awsgluecatalogdatabase.AwsGlueCatalogDatabase
	Spec         *awsgluecatalogdatabase.AwsGlueCatalogDatabaseSpec
	AwsTags      map[string]string
	DatabaseName string
}

func initializeLocals(ctx *pulumi.Context, in *awsgluecatalogdatabase.AwsGlueCatalogDatabaseStackInput) *Locals {
	locals := &Locals{}
	locals.Target = in.Target
	locals.Spec = in.Target.Spec
	locals.DatabaseName = in.Target.Metadata.Name

	locals.AwsTags = map[string]string{
		awstagkeys.Resource:     strconv.FormatBool(true),
		awstagkeys.Organization: locals.Target.Metadata.Org,
		awstagkeys.Environment:  locals.Target.Metadata.Env,
		awstagkeys.ResourceKind: cloudresourcekind.CloudResourceKind_AwsGlueCatalogDatabase.String(),
		awstagkeys.ResourceId:   locals.Target.Metadata.Id,
	}

	return locals
}
