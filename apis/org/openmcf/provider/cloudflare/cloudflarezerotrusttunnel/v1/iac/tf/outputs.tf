output "tunnel_id" {
  description = "The Cloudflare-assigned UUID of the tunnel (referenced by routes)"
  value       = cloudflare_zero_trust_tunnel_cloudflared.main.id
}

output "tunnel_cname" {
  description = "The CNAME target for public hostnames served by this tunnel"
  value       = "${cloudflare_zero_trust_tunnel_cloudflared.main.id}.cfargotunnel.com"
}

output "tunnel_token" {
  description = "The connector run token used by cloudflared to authenticate (sensitive)"
  value       = data.cloudflare_zero_trust_tunnel_cloudflared_token.main.token
  sensitive   = true
}

output "tunnel_status" {
  description = "The tunnel status: inactive, degraded, healthy, or down"
  value       = cloudflare_zero_trust_tunnel_cloudflared.main.status
}

output "account_tag" {
  description = "The Cloudflare account tag the tunnel belongs to"
  value       = cloudflare_zero_trust_tunnel_cloudflared.main.account_tag
}

output "created_on" {
  description = "RFC3339 timestamp of when the tunnel was created"
  value       = cloudflare_zero_trust_tunnel_cloudflared.main.created_at
}
