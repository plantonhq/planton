# Cloudflare Tunnel (cloudflared): a secure, outbound-only connection from a private
# network to Cloudflare's edge. The remote ingress configuration is a separate resource
# so editing ingress never recreates the tunnel.
resource "cloudflare_zero_trust_tunnel_cloudflared" "main" {
  account_id = var.spec.account_id
  name       = var.spec.name
  config_src = var.spec.config_src

  tunnel_secret = try(var.spec.tunnel_secret, "") != "" ? var.spec.tunnel_secret : null
}

resource "cloudflare_zero_trust_tunnel_cloudflared_config" "main" {
  count = local.manage_config ? 1 : 0

  account_id = var.spec.account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.main.id
  source     = "cloudflare"

  config = {
    ingress        = local.ingress
    origin_request = local.origin_request
  }
}

# The connector run token is exposed by a data source, not the resource itself.
data "cloudflare_zero_trust_tunnel_cloudflared_token" "main" {
  account_id = var.spec.account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.main.id
}
