# main.tf

# Create the Cloudflare D1 database (a serverless SQLite database a Worker
# reaches via a d1 binding). Placement is fixed at creation by an optional
# region hint.
resource "cloudflare_d1_database" "main" {
  account_id = var.spec.account_id
  name       = var.spec.database_name

  # Region hint (omitted when unspecified so Cloudflare selects a default).
  primary_location_hint = local.primary_location_hint

  # Data-residency jurisdiction (omitted when unset; mutually exclusive with region).
  jurisdiction = local.jurisdiction

  # Read replication is a single nested attribute; set it only when configured.
  read_replication = var.spec.read_replication != null ? {
    mode = var.spec.read_replication.mode
  } : null
}
