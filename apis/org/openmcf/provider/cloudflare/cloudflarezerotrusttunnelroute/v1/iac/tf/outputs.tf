output "route_id" {
  description = "The Cloudflare-assigned UUID of the route"
  value       = cloudflare_zero_trust_tunnel_cloudflared_route.main.id
}

output "network" {
  description = "The private CIDR advertised by this route"
  value       = cloudflare_zero_trust_tunnel_cloudflared_route.main.network
}
