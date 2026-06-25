locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-load-balancer-monitor")

  # The enum flattens to its string name; the unspecified zero value maps to the
  # Cloudflare default protocol (http).
  monitor_type = (
    try(var.spec.type, "") == "" || var.spec.type == "monitor_type_unspecified"
  ) ? "http" : var.spec.type

  # HTTP headers as the provider's map(name -> list(values)) shape.
  headers = { for h in try(var.spec.headers, []) : h.name => h.values }
}
