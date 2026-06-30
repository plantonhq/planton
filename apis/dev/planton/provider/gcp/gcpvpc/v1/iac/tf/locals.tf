locals {
  # Create GCP labels from metadata and spec
  gcp_labels = {
    resource     = var.spec.network_name
    resource-id  = var.metadata.id
    resource-org = var.metadata.org != null ? var.metadata.org : ""
    env          = var.metadata.env != null ? var.metadata.env : ""
  }

  # Map proto enum to GCP routing mode string
  # 0=REGIONAL, 1=GLOBAL
  routing_mode_map = {
    0 = "REGIONAL"
    1 = "GLOBAL"
  }
  
  # Default to REGIONAL if routing_mode not specified
  routing_mode = var.spec.routing_mode != null ? lookup(local.routing_mode_map, var.spec.routing_mode, "REGIONAL") : "REGIONAL"

  # Private Services Access configuration
  enable_private_services = try(var.spec.private_services_access.enabled, false)
  private_services_prefix_length = try(var.spec.private_services_access.ip_range_prefix_length, 16)
}

