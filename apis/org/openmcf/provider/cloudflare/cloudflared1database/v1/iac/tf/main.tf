# main.tf

# Create the Cloudflare D1 database
resource "cloudflare_d1_database" "main" {
  account_id = var.spec.account_id
  name       = var.spec.database_name

  # Add optional primary location hint (region) if specified
  primary_location_hint = var.spec.region

  # Read replication is a single nested attribute; set it only when configured.
  read_replication = var.spec.read_replication != null ? {
    mode = var.spec.read_replication.mode
  } : null
}

