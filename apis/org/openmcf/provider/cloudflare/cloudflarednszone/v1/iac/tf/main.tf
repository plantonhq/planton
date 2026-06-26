# main.tf

# Create the Cloudflare DNS Zone
resource "cloudflare_zone" "main" {
  account = {
    id = var.spec.account_id
  }
  name   = var.spec.zone_name
  paused = var.spec.paused
  type   = local.zone_type

  vanity_name_servers = length(var.spec.vanity_name_servers) > 0 ? var.spec.vanity_name_servers : null
}
