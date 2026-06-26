output "pool_id" {
  description = "The Cloudflare-assigned ID of the pool (referenced by a load balancer's pool lists)"
  value       = cloudflare_load_balancer_pool.main.id
}

output "pool_name" {
  description = "The pool name"
  value       = cloudflare_load_balancer_pool.main.name
}
