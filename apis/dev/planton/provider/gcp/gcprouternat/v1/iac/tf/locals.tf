locals {
  # Router and NAT names from spec
  router_name = var.spec.router_name
  nat_name    = var.spec.nat_name

  # Determine NAT IP allocation strategy
  nat_ip_allocate_option = length(var.spec.nat_ip_names) > 0 ? "MANUAL_ONLY" : "AUTO_ONLY"

  # Determine subnet coverage mode
  source_subnetwork_ip_ranges_to_nat = length(var.spec.subnetwork_self_links) > 0 ? "LIST_OF_SUBNETWORKS" : "ALL_SUBNETWORKS_ALL_IP_RANGES"

  # Build subnetworks configuration (only if specific subnets are provided)
  subnetworks = length(var.spec.subnetwork_self_links) > 0 ? [
    for subnet in var.spec.subnetwork_self_links : {
      name                    = subnet
      source_ip_ranges_to_nat = ["ALL_IP_RANGES"]
    }
  ] : []

  # Logging configuration
  enable_logging = var.spec.log_filter != "DISABLED"
  log_filter     = var.spec.log_filter != "DISABLED" ? var.spec.log_filter : "ERRORS_ONLY"

  # GCP labels for resource tagging
  gcp_labels = merge(
    {
      "resource"      = "true"
      "resource-name" = var.spec.router_name
      "resource-kind" = "gcprouternat"
    },
    var.metadata.org != null ? { "organization" = var.metadata.org } : {},
    var.metadata.env != null ? { "environment" = var.metadata.env } : {},
    var.metadata.id != null ? { "resource-id" = var.metadata.id } : {},
    var.metadata.labels != null ? var.metadata.labels : {}
  )
}

