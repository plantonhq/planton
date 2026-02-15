package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp"
	"github.com/pulumi/pulumi-gcp/sdk/v9/go/gcp/firestore"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func firestoreDatabase(ctx *pulumi.Context, locals *Locals, gcpProvider *gcp.Provider) error {
	spec := locals.GcpFirestoreDatabase.Spec

	args := &firestore.DatabaseArgs{
		LocationId: pulumi.String(spec.LocationId),
		Type:       pulumi.String(spec.Type),
		Name:       pulumi.StringPtr(spec.DatabaseName),
		Project:    pulumi.StringPtr(spec.ProjectId.GetValue()),
	}

	// Concurrency mode.
	if spec.ConcurrencyMode != "" {
		args.ConcurrencyMode = pulumi.StringPtr(spec.ConcurrencyMode)
	}

	// Point-in-time recovery.
	if spec.PointInTimeRecoveryEnablement != "" {
		args.PointInTimeRecoveryEnablement = pulumi.StringPtr(spec.PointInTimeRecoveryEnablement)
	}

	// Delete protection state.
	if spec.GetDeleteProtectionState() != "" {
		args.DeleteProtectionState = pulumi.StringPtr(spec.GetDeleteProtectionState())
	}

	// Database edition.
	if spec.DatabaseEdition != "" {
		args.DatabaseEdition = pulumi.StringPtr(spec.DatabaseEdition)
	}

	// CMEK encryption.
	if spec.KmsKeyName != nil && spec.KmsKeyName.GetValue() != "" {
		args.CmekConfig = &firestore.DatabaseCmekConfigArgs{
			KmsKeyName: pulumi.String(spec.KmsKeyName.GetValue()),
		}
	}

	// Always set deletion_policy to DELETE so the IaC tool manages the
	// full lifecycle. Without this, the default "ABANDON" would leave the
	// database behind when the stack is destroyed.
	args.DeletionPolicy = pulumi.StringPtr("DELETE")

	createdDatabase, err := firestore.NewDatabase(ctx, "firestore-database", args, pulumi.Provider(gcpProvider))
	if err != nil {
		return errors.Wrap(err, "failed to create firestore database")
	}

	// Export outputs.
	ctx.Export(OpDatabaseId, pulumi.Sprintf(
		"projects/%s/databases/%s",
		pulumi.String(spec.ProjectId.GetValue()),
		createdDatabase.Name,
	))
	ctx.Export(OpDatabaseName, createdDatabase.Name)
	ctx.Export(OpUid, createdDatabase.Uid)
	ctx.Export(OpCreateTime, createdDatabase.CreateTime)
	ctx.Export(OpEarliestVersionTime, createdDatabase.EarliestVersionTime)

	return nil
}
