# Cloudflare Tunnel route: advertises a private CIDR as reachable through a tunnel,
# within a virtual network. Independent lifecycle from the tunnel, so adding or removing
# reachable networks never recreates the tunnel.
resource "cloudflare_zero_trust_tunnel_cloudflared_route" "main" {
  account_id = var.spec.account_id
  network    = var.spec.network
  tunnel_id  = var.spec.tunnel_id

  virtual_network_id = try(var.spec.virtual_network_id, "") != "" ? var.spec.virtual_network_id : null
  comment            = try(var.spec.comment, "") != "" ? var.spec.comment : null
}
