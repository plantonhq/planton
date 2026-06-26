# Cloudflare Tunnel virtual network: an isolated routing segment that lets overlapping
# private CIDRs be reached through separate tunnels. Account-scoped and independent of
# any tunnel's lifecycle.
resource "cloudflare_zero_trust_tunnel_cloudflared_virtual_network" "main" {
  account_id         = var.spec.account_id
  name               = var.spec.name
  is_default_network = var.spec.is_default_network

  comment = try(var.spec.comment, "") != "" ? var.spec.comment : null
}
