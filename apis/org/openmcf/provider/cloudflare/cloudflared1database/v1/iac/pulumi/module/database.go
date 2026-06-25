package module

import (
	"github.com/pkg/errors"
	cloudflared1databasev1 "github.com/plantonhq/openmcf/apis/org/openmcf/provider/cloudflare/cloudflared1database/v1"
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
	if locals.CloudflareD1Database.Spec.Region != cloudflared1databasev1.CloudflareD1Region_cloudflare_d1_region_unspecified {
		regionStr := mapRegionToString(locals.CloudflareD1Database.Spec.Region)
		if regionStr != "" {
			d1Args.PrimaryLocationHint = pulumi.String(regionStr)
		}
	}

	// 3. Add optional read replication configuration if specified.
	if locals.CloudflareD1Database.Spec.ReadReplication != nil {
		d1Args.ReadReplication = &cloudflare.D1DatabaseReadReplicationArgs{
			Mode: pulumi.String(locals.CloudflareD1Database.Spec.ReadReplication.Mode.String()),
		}
	}

	// 4.  Create the resource.
	createdD1Database, err := cloudflare.NewD1Database(
		ctx,
		"database",
		d1Args,
		pulumi.Provider(cloudflareProvider),
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cloudflare d1 database")
	}

	// 5.  Export stack outputs. A Worker reaches D1 through its binding, so there
	// is no connection string to export (none exists on the v5 resource).
	ctx.Export(OpDatabaseId, createdD1Database.ID())
	ctx.Export(OpDatabaseName, createdD1Database.Name)

	return createdD1Database, nil
}

// mapRegionToString converts the proto enum to the Cloudflare API region string.
func mapRegionToString(region cloudflared1databasev1.CloudflareD1Region) string {
	switch region {
	case cloudflared1databasev1.CloudflareD1Region_weur:
		return "weur"
	case cloudflared1databasev1.CloudflareD1Region_eeur:
		return "eeur"
	case cloudflared1databasev1.CloudflareD1Region_apac:
		return "apac"
	case cloudflared1databasev1.CloudflareD1Region_oc:
		return "oc"
	case cloudflared1databasev1.CloudflareD1Region_wnam:
		return "wnam"
	case cloudflared1databasev1.CloudflareD1Region_enam:
		return "enam"
	default:
		return ""
	}
}
