package module

import (
	"github.com/pkg/errors"
	cloudflared1databasev1alpha1 "github.com/plantonhq/planton/catalog/cloudflare/cloudflared1database/v1alpha1"
	"github.com/pulumi/pulumi-cloudflare/sdk/v6/go/cloudflare"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// database provisions the Cloudflare D1 database and exports its outputs.
func database(
	ctx *pulumi.Context,
	locals *Locals,
	cloudflareProvider *cloudflare.Provider,
) (*cloudflare.D1Database, error) {

	// 1.  Build arguments directly from proto fields—no extra structs.
	d1Args := &cloudflare.D1DatabaseArgs{
		AccountId: pulumi.String(locals.CloudflareD1Database.Spec.AccountId),
		Name:      pulumi.String(locals.CloudflareD1Database.Spec.DatabaseName),
	}

	// 2. Add optional primary location hint (region) if specified.
	if locals.CloudflareD1Database.Spec.Region != cloudflared1databasev1alpha1.CloudflareD1Region_cloudflare_d1_region_unspecified {
		regionStr := mapRegionToString(locals.CloudflareD1Database.Spec.Region)
		if regionStr != "" {
			d1Args.PrimaryLocationHint = pulumi.String(regionStr)
		}
	}

	// 3. Add optional data-residency jurisdiction (mutually exclusive with region).
	if locals.CloudflareD1Database.Spec.Jurisdiction != "" {
		d1Args.Jurisdiction = pulumi.String(locals.CloudflareD1Database.Spec.Jurisdiction)
	}

	// 4. Read replication: Cloudflare always reports it on read (mode
	// "disabled" when never configured) while the provider models the
	// attribute as Optional-not-Computed, so an omitted spec block coalesces
	// to the server default -- sending nothing leaves a refresh-vs-config
	// diff that never converges (measured live 2026-08-26 on the terraform
	// module; this engine carries the same latent class).
	readReplicationMode := "disabled"
	if locals.CloudflareD1Database.Spec.ReadReplication != nil {
		readReplicationMode = locals.CloudflareD1Database.Spec.ReadReplication.Mode.String()
	}
	d1Args.ReadReplication = &cloudflare.D1DatabaseReadReplicationArgs{
		Mode: pulumi.String(readReplicationMode),
	}

	// 5.  Create the resource. primary_location_hint is a create-time
	// placement hint Cloudflare never returns (no_refresh + RequiresReplace
	// upstream at v5.23.0): without the ignore, adopting an existing database
	// restores the hint as null and a post-create region edit plans a
	// REPLACE -- destroying the database to move a hint. Placement is decided
	// at birth; post-create edits are deliberately inert.
	createdD1Database, err := cloudflare.NewD1Database(
		ctx,
		"database",
		d1Args,
		pulumi.Provider(cloudflareProvider),
		pulumi.IgnoreChanges([]string{"primaryLocationHint"}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare d1 database")
	}

	// 5.  Export outputs. A Worker reaches D1 through its binding, so there
	// is no connection string to export (none exists on the v5 resource).
	ctx.Export(OpDatabaseId, createdD1Database.ID())
	ctx.Export(OpDatabaseName, createdD1Database.Name)
	ctx.Export(OpVersion, createdD1Database.Version)

	return createdD1Database, nil
}

// mapRegionToString converts the proto enum to the Cloudflare API region string.
func mapRegionToString(region cloudflared1databasev1alpha1.CloudflareD1Region) string {
	switch region {
	case cloudflared1databasev1alpha1.CloudflareD1Region_weur:
		return "weur"
	case cloudflared1databasev1alpha1.CloudflareD1Region_eeur:
		return "eeur"
	case cloudflared1databasev1alpha1.CloudflareD1Region_apac:
		return "apac"
	case cloudflared1databasev1alpha1.CloudflareD1Region_oc:
		return "oc"
	case cloudflared1databasev1alpha1.CloudflareD1Region_wnam:
		return "wnam"
	case cloudflared1databasev1alpha1.CloudflareD1Region_enam:
		return "enam"
	default:
		return ""
	}
}
