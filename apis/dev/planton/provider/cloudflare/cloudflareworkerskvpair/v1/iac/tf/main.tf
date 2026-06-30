# A single Workers KV entry (key/value pair) written into an existing namespace.
resource "cloudflare_workers_kv" "main" {
  account_id   = var.spec.account_id
  namespace_id = var.spec.namespace_id
  key_name     = var.spec.key_name
  value        = var.spec.value
  metadata     = try(var.spec.metadata, "") != "" ? var.spec.metadata : null
}
