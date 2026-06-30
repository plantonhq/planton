locals {
  # Use custom display_name if provided, otherwise fallback to metadata.name
  display_name = var.spec.display_name != null && var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  # Determine cluster type (default to STANDARD if not specified)
  cluster_type = var.spec.cluster_type != null && var.spec.cluster_type != "" ? var.spec.cluster_type : "STANDARD"

  # Determine if network configuration is provided
  has_network_config = var.spec.network_config != null && var.spec.network_config.network_id != null && var.spec.network_config.network_id != ""

  # Validate dedicated cluster configuration
  is_dedicated = local.cluster_type == "DEDICATED"
  has_dedicated_config = var.spec.dedicated_config != null && var.spec.dedicated_config.cku != null

  # Resource tags (combine metadata labels and tags)
  resource_labels = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      "managed-by" = "planton"
      "component"  = "confluent-kafka"
    }
  )
}

