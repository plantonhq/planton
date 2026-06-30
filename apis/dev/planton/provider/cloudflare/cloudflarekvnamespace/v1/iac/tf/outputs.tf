output "namespace_id" {
  description = "The unique identifier of the created KV namespace"
  value       = cloudflare_workers_kv_namespace.main.id
}

output "supports_url_encoding" {
  description = "Whether keys in this namespace support URL encoding"
  value       = cloudflare_workers_kv_namespace.main.supports_url_encoding
}
