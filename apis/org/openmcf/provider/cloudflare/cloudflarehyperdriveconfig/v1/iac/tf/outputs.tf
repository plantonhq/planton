output "hyperdrive_id" {
  description = "The Cloudflare-assigned ID of the Hyperdrive config (referenced by a Worker hyperdrive binding)"
  value       = cloudflare_hyperdrive_config.main.id
}

output "name" {
  description = "The Hyperdrive config name"
  value       = cloudflare_hyperdrive_config.main.name
}
