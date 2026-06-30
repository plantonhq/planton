output "monitor_id" {
  description = "The Cloudflare-assigned ID of the monitor (referenced by a pool's monitor field)"
  value       = cloudflare_load_balancer_monitor.main.id
}

output "monitor_type" {
  description = "The health-check protocol of the monitor"
  value       = cloudflare_load_balancer_monitor.main.type
}
