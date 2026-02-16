package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/glue"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func catalogDatabase(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	args := &glue.CatalogDatabaseArgs{
		Name: pulumi.StringPtr(locals.DatabaseName),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	if spec.LocationUri != "" {
		args.LocationUri = pulumi.StringPtr(spec.LocationUri)
	}

	db, err := glue.NewCatalogDatabase(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create Glue Catalog Database")
	}

	ctx.Export(OpDatabaseName, db.Name)
	ctx.Export(OpDatabaseArn, db.Arn)
	ctx.Export(OpCatalogId, db.CatalogId)

	return nil
}
