output "load_balancer_id" {
  description = "The ID of the Cloudflare Load Balancer"
  value       = cloudflare_load_balancer.main.id
}

output "load_balancer_dns_record_name" {
  description = "The DNS record name (hostname) of the load balancer"
  value       = cloudflare_load_balancer.main.name
}

output "load_balancer_cname_target" {
  description = "The CNAME target clients point their DNS at (the load balancer hostname)"
  value       = cloudflare_load_balancer.main.name
}
