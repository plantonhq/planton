package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/spanner"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func spannerDatabase(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpSpannerDatabase.Spec

	args := &spanner.DatabaseArgs{
		Instance: pulumi.String(spec.Instance.GetValue()),
		Name:     pulumi.StringPtr(spec.DatabaseName),
		Project:  pulumi.StringPtr(spec.ProjectId.GetValue()),
	}

	// Database dialect (immutable).
	if spec.DatabaseDialect != "" {
		args.DatabaseDialect = pulumi.StringPtr(spec.DatabaseDialect)
	}

	// Version retention period for point-in-time recovery.
	if spec.VersionRetentionPeriod != "" {
		args.VersionRetentionPeriod = pulumi.StringPtr(spec.VersionRetentionPeriod)
	}

	// Initial DDL statements.
	if len(spec.Ddl) > 0 {
		args.Ddls = pulumi.ToStringArray(spec.Ddl)
	}

	// GCP API-level drop protection.
	if spec.EnableDropProtection {
		args.EnableDropProtection = pulumi.BoolPtr(true)
	}

	// CMEK encryption.
	if spec.KmsKeyName != nil && spec.KmsKeyName.GetValue() != "" {
		args.EncryptionConfig = &spanner.DatabaseEncryptionConfigArgs{
			KmsKeyName: pulumi.StringPtr(spec.KmsKeyName.GetValue()),
		}
	}

	// Default time zone.
	if spec.DefaultTimeZone != "" {
		args.DefaultTimeZone = pulumi.StringPtr(spec.DefaultTimeZone)
	}

	createdDatabase, err := spanner.NewDatabase(ctx, "spanner-database", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create spanner database")
	}

	// Export outputs.
	// database_id: fully qualified path projects/{project}/instances/{instance}/databases/{name}.
	ctx.Export(OpDatabaseId, pulumi.Sprintf(
		"projects/%s/instances/%s/databases/%s",
		pulumi.String(spec.ProjectId.GetValue()),
		pulumi.String(spec.Instance.GetValue()),
		createdDatabase.Name,
	))
	ctx.Export(OpDatabaseName, createdDatabase.Name)
	ctx.Export(OpState, createdDatabase.State)

	return nil
}
