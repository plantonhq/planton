package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws"
	"github.com/pulumi/pulumi-aws/sdk/v7/go/aws/glue"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// catalogDatabase creates the Glue Data Catalog database and exports its
// outputs.
//
// A single metadata resource with three creation shapes: a regular database,
// a resource link to a database shared from another account/region
// (target_database), or a federated projection of an external source
// (federated_database). The spec's CEL keeps the shapes exclusive, so the
// module simply maps each optional message onto its provider input.
func catalogDatabase(ctx *pulumi.Context, locals *Locals, provider *aws.Provider) error {
	spec := locals.Spec

	args := &glue.CatalogDatabaseArgs{
		// The cloud name is set explicitly from metadata.name (never Pulumi
		// auto-naming) so both engines create the identical database. The
		// name is passed verbatim -- AWS rejects uppercase; the module never
		// silently transforms it.
		Name: pulumi.StringPtr(locals.DatabaseName),
		Tags: pulumi.ToStringMap(locals.AwsTags),
	}

	// Defaults to the deploying account's own catalog; set only for the
	// cross-account create-in-another-catalog governance pattern. ForceNew.
	if spec.CatalogId != "" {
		args.CatalogId = pulumi.StringPtr(spec.CatalogId)
	}

	if spec.Description != "" {
		args.Description = pulumi.StringPtr(spec.Description)
	}

	if spec.LocationUri != "" {
		args.LocationUri = pulumi.StringPtr(spec.LocationUri)
	}

	// Catalog metadata properties read by engines and governance tooling --
	// distinct from the AWS resource tags above.
	if len(spec.Parameters) > 0 {
		args.Parameters = pulumi.ToStringMap(spec.Parameters)
	}

	// Default Lake Formation grants applied to tables CREATED in this
	// database. An entry with an empty permissions list (and no principal)
	// is meaningful: it stops granting ALL to IAM_ALLOWED_PRINCIPALS on new
	// tables -- the hardening step when moving to Lake Formation-managed
	// permissions.
	if len(spec.CreateTableDefaultPermissions) > 0 {
		perms := glue.CatalogDatabaseCreateTableDefaultPermissionArray{}
		for _, entry := range spec.CreateTableDefaultPermissions {
			permArgs := &glue.CatalogDatabaseCreateTableDefaultPermissionArgs{
				Permissions: pulumi.ToStringArray(entry.Permissions),
			}
			if entry.Principal != "" {
				permArgs.Principal = &glue.CatalogDatabaseCreateTableDefaultPermissionPrincipalArgs{
					DataLakePrincipalIdentifier: pulumi.StringPtr(entry.Principal),
				}
			}
			perms = append(perms, permArgs)
		}
		args.CreateTableDefaultPermissions = perms
	}

	// Resource link: a local pointer to a database shared via AWS RAM /
	// Lake Formation. All coordinates are ForceNew.
	if td := spec.TargetDatabase; td != nil {
		targetArgs := &glue.CatalogDatabaseTargetDatabaseArgs{
			CatalogId:    pulumi.String(td.CatalogId),
			DatabaseName: pulumi.String(td.DatabaseName),
		}
		if td.Region != "" {
			targetArgs.Region = pulumi.StringPtr(td.Region)
		}
		args.TargetDatabase = targetArgs
	}

	// Federated database: projects an external source (e.g. a Redshift
	// datashare) into the catalog through a Glue connection.
	if fd := spec.FederatedDatabase; fd != nil {
		fedArgs := &glue.CatalogDatabaseFederatedDatabaseArgs{}
		if fd.Identifier != "" {
			fedArgs.Identifier = pulumi.StringPtr(fd.Identifier)
		}
		if fd.ConnectionName != "" {
			fedArgs.ConnectionName = pulumi.StringPtr(fd.ConnectionName)
		}
		args.FederatedDatabase = fedArgs
	}

	db, err := glue.NewCatalogDatabase(ctx, locals.Target.Metadata.Name, args, pulumi.Provider(provider))
	if err != nil {
		return errors.Wrap(err, "failed to create Glue Catalog Database")
	}

	// Outputs (contract: AwsGlueCatalogDatabaseOutputs).
	ctx.Export(OpDatabaseName, db.Name)
	ctx.Export(OpDatabaseArn, db.Arn)
	ctx.Export(OpCatalogId, db.CatalogId)

	return nil
}
