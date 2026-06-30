locals {
  # IP address type defaults to ipv4 when not specified.
  ip_address_type = coalesce(var.spec.ip_address_type, "ipv4")

  # Build a map of listeners keyed by name for use with for_each.
  listener_map = { for l in var.spec.listeners : l.name => l }

  # Build DNS records map when DNS is enabled.
  dns_records = try(var.spec.dns.enabled, false) ? {
    for idx, hostname in try(var.spec.dns.hostnames, []) : "dns-${idx}" => hostname
  } : {}

  # Standard tags applied to all resources.
  tags = {
    "planton.dev/resource"      = "true"
    "planton.dev/organization"  = var.metadata.org
    "planton.dev/environment"   = var.metadata.env
    "planton.dev/resource-kind" = "AwsNetworkLoadBalancer"
    "planton.dev/resource-id"   = var.metadata.id
  }
}
