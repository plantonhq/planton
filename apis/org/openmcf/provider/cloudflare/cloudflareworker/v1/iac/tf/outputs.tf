output "script_id" {
  description = "The ID of the Cloudflare Worker script"
  value       = cloudflare_workers_script.main.id
}

output "route_urls" {
  description = "List of route URLs where the Worker is accessible"
  value = local.dns_enabled ? [
    "https://${local.dns_hostname}"
  ] : []
}

