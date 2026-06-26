output "key_name" {
  description = "The entry's key name"
  value       = cloudflare_workers_kv.main.key_name
}

output "namespace_id" {
  description = "The namespace ID the entry was written to"
  value       = cloudflare_workers_kv.main.namespace_id
}
