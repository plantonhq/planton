package module

import (
	"github.com/pkg/errors"
	"github.com/pulumi/pulumi-civo/sdk/v2/go/civo"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// database provisions the managed database instance and exports outputs.
func database(
	ctx *pulumi.Context,
	locals *Locals,
	civoProvider *civo.Provider,
) (*civo.Database, error) {

	// 1. Get engine slug directly from enum (values match Civo API strings).
	if locals.CivoDatabase.Spec.Engine == 0 {
		return nil, errors.Errorf("database engine is required")
	}
	engineSlug := locals.CivoDatabase.Spec.Engine.String()

	// 2. Build resource arguments from proto fields.
	databaseArgs := &civo.DatabaseArgs{
		Name:    pulumi.String(locals.CivoDatabase.Spec.DbInstanceName),
		Engine:  pulumi.String(engineSlug),
		Version: pulumi.String(locals.CivoDatabase.Spec.EngineVersion),
		Region:  pulumi.String(locals.CivoDatabase.Spec.Region.String()),
		Size:    pulumi.String(locals.CivoDatabase.Spec.SizeSlug),
		// Replica count in Civo = total nodes; primary + replicas.
		Nodes: pulumi.Int(int(locals.CivoDatabase.Spec.Replicas) + 1),
	}

	// Private network attachment
	if locals.CivoDatabase.Spec.NetworkId != nil && locals.CivoDatabase.Spec.NetworkId.GetValue() != "" {
		databaseArgs.NetworkId = pulumi.StringPtr(locals.CivoDatabase.Spec.NetworkId.GetValue())
	}

	// Optional: firewalls (first one only—Civo currently allows one).
	if len(locals.CivoDatabase.Spec.FirewallIds) > 0 &&
		locals.CivoDatabase.Spec.FirewallIds[0].GetValue() != "" {
		databaseArgs.FirewallId = pulumi.String(
			locals.CivoDatabase.Spec.FirewallIds[0].GetValue())
	}

	// Note: storage_gib and tags fields are defined in spec.proto but not currently
	// supported by the pulumi-civo provider v2 for Database resources.
	// These will be enabled when the provider is updated to support them.
	// See: https://github.com/civo/terraform-provider-civo/blob/main/civo/database/resource_database.go

	// 3. Provision the database.
	createdDatabase, err := civo.NewDatabase(
		ctx,
		"database",
		databaseArgs,
		pulumi.Provider(civoProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Civo database")
	}

	// 4. Export stack outputs so Planton can wire dependencies.
	ctx.Export(OpDatabaseId, createdDatabase.ID())
	ctx.Export(OpHost, createdDatabase.DnsEndpoint)
	ctx.Export(OpPort, createdDatabase.Port)
	ctx.Export(OpUsername, createdDatabase.Username)
	ctx.Export(OpPasswordSecretRef, createdDatabase.Password) // secret‑ref in future

	return createdDatabase, nil
}
