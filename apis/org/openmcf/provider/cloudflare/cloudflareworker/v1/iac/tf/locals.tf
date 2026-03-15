locals {
  # Resource naming
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-worker")
  
  # Labels/tags
  labels = merge({
    "name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Worker script configuration
  script_name = var.spec.worker_name
  
  # R2 bundle configuration
  r2_bucket = var.spec.script_bundle.bucket
  r2_path   = var.spec.script_bundle.path

  # KV bindings
  kv_bindings = try(var.spec.kv_bindings, [])

  # DNS configuration
  dns_enabled = try(var.spec.dns.enabled, false)
  dns_zone_id = try(var.spec.dns.zone_id.value, "")
  dns_hostname = try(var.spec.dns.hostname, "")
  
  # Route pattern (defaults to hostname/* if not specified)
  route_pattern = try(var.spec.dns.route_pattern, "") != "" ? var.spec.dns.route_pattern : "${local.dns_hostname}/*"

  # Compatibility date (defaults to today if not specified)
  compatibility_date = try(var.spec.compatibility_date, "") != "" ? var.spec.compatibility_date : formatdate("YYYY-MM-DD", timestamp())

  # Usage model mapping
  usage_model_map = {
    0 = "bundled"
    1 = "unbound"
  }
  usage_model = lookup(local.usage_model_map, try(var.spec.usage_model, 0), "bundled")

  # Environment variables
  env_variables = try(var.spec.env.variables, {})
  
  # Secrets (note: Terraform doesn't support direct secret upload via Workers API)
  env_secrets = try(var.spec.env.secrets, {})
}

