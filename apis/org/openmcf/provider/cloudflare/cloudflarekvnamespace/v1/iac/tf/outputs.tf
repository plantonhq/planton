output "namespace_id" {
  description = "The unique identifier of the created KV namespace"
  value       = cloudflare_workers_kv_namespace.main.id
}
