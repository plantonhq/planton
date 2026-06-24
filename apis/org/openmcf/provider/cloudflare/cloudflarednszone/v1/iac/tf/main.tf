# main.tf

# Create the Cloudflare DNS Zone
resource "cloudflare_zone" "main" {
  account = {
    id = var.spec.account_id
  }
  name   = var.spec.zone_name
  paused = var.spec.paused

  # Type is always "full" (standard nameserver delegation)
  # "partial" (CNAME setup) is Business/Enterprise only and not in the 80/20
  type = "full"
}
