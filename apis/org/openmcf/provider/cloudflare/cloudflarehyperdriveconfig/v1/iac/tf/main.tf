# Cloudflare Hyperdrive config: a pooled, cached connection to a regional SQL
# database that a Worker reaches via a `hyperdrive` binding.
resource "cloudflare_hyperdrive_config" "main" {
  account_id = var.spec.account_id
  name       = var.spec.name

  origin = {
    database = local.origin.database
    scheme   = local.origin.scheme
    user     = local.origin.user
    host     = try(local.origin.host, "") != "" ? local.origin.host : null
    # Omit port when unset so the engine default (5432 PG / 3306 MySQL) applies.
    port     = try(local.origin.port, 0) > 0 ? local.origin.port : null
    password = local.origin.password

    access_client_id     = try(local.origin.access_client_id, "") != "" ? local.origin.access_client_id : null
    access_client_secret = try(local.origin.access_client_secret, "") != "" ? local.origin.access_client_secret : null

    # Egress through a Workers VPC Service when set; omit so the public host applies.
    service_id = try(local.origin.service_id, "") != "" ? local.origin.service_id : null
  }

  caching = local.caching

  mtls = local.mtls

  origin_connection_limit = var.spec.origin_connection_limit > 0 ? var.spec.origin_connection_limit : null
}
