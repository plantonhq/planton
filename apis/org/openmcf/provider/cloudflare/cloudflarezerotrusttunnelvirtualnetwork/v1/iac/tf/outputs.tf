output "virtual_network_id" {
  description = "The Cloudflare-assigned UUID of the virtual network (referenced by a tunnel route)"
  value       = cloudflare_zero_trust_tunnel_cloudflared_virtual_network.main.id
}

output "virtual_network_name" {
  description = "The virtual network name"
  value       = cloudflare_zero_trust_tunnel_cloudflared_virtual_network.main.name
}
