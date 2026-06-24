# Cloudflare Workers KV namespace.
# ttl_seconds and description are part of the OpenMCF spec but have no
# representation on the Cloudflare KV namespace resource, so they are not set.
resource "cloudflare_workers_kv_namespace" "main" {
  account_id = local.account_id
  title      = local.namespace_name
}
