locals {
  # Node pool name
  node_pool_name = var.metadata.name

  # Cluster name
  cluster_name = var.spec.cluster_name

  # Resource group name (explicit spec field)
  resource_group_name = var.spec.resource_group

  # OS type mapping
  os_type = var.spec.os_type == "WINDOWS" ? "Windows" : "Linux"

  # Mode mapping
  mode = var.spec.mode == "SYSTEM" ? "System" : "User"

  # Tags from metadata
  tags = merge(
    var.metadata.labels != null ? var.metadata.labels : {},
    {
      Name        = var.metadata.name
      Environment = var.metadata.env != null ? var.metadata.env : "default"
      ManagedBy   = "terraform"
    }
  )
}

