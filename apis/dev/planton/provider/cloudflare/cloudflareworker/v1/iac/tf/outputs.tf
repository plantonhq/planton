output "script_id" {
  description = "The Cloudflare-assigned ID of the deployed Worker script"
  value       = cloudflare_workers_script.main.id
}

output "script_name" {
  description = "The Worker script name (the target a service binding references)"
  value       = cloudflare_workers_script.main.script_name
}

output "custom_domain_hostnames" {
  description = "The custom-domain hostnames attached to this Worker"
  value       = [for cd in cloudflare_workers_custom_domain.main : cd.hostname]
}

output "route_patterns" {
  description = "The route patterns mapped to this Worker"
  value       = [for r in cloudflare_workers_route.main : r.pattern]
}
