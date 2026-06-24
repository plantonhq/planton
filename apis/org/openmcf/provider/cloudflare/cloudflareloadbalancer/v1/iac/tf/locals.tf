locals {
  # Resource naming
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-load-balancer")
  
  # Tags/labels
  labels = merge({
    "name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Zone ID (StringValueOrRef flattened to a plain string by the converter)
  zone_id = try(var.spec.zone_id, "")

  # Health probe path with default
  health_probe_path = coalesce(try(var.spec.health_probe_path, null), "/")

  # Proxied setting with default
  proxied = coalesce(try(var.spec.proxied, null), true)

  # Session affinity and steering policy: the enum flattens to its string name,
  # which matches the value Cloudflare expects directly.
  session_affinity = coalesce(try(var.spec.session_affinity, null), "none")
  steering_policy   = coalesce(try(var.spec.steering_policy, null), "off")

  # Origins list
  origins = try(var.spec.origins, [])
}

