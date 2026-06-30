locals {
  # Cluster name from spec
  cluster_name = var.spec.cluster_name
  
  # Region from spec
  region = var.spec.region
  
  # Kubernetes version
  kubernetes_version = var.spec.kubernetes_version
  
  # VPC UUID
  vpc_uuid = var.spec.vpc.value
  
  # HA control plane flag
  highly_available = var.spec.highly_available
  
  # Auto-upgrade settings
  auto_upgrade = var.spec.auto_upgrade
  surge_upgrade = !var.spec.disable_surge_upgrade
  
  # Registry integration
  registry_integration = var.spec.registry_integration
  
  # Maintenance window (only if specified)
  maintenance_window = var.spec.maintenance_window != "" ? var.spec.maintenance_window : null
  
  # Control plane firewall (only if IPs specified)
  has_firewall_ips = length(var.spec.control_plane_firewall_allowed_ips) > 0
  firewall_ips = var.spec.control_plane_firewall_allowed_ips
  
  # Tags - merge spec tags with metadata tags
  all_tags = concat(
    var.spec.tags,
    var.metadata.tags != null ? var.metadata.tags : []
  )
  
  # Node pool configuration
  node_pool_name = "default"
  node_pool_size = var.spec.default_node_pool.size
  node_pool_count = var.spec.default_node_pool.node_count
  node_pool_auto_scale = var.spec.default_node_pool.auto_scale
  node_pool_min_nodes = var.spec.default_node_pool.min_nodes
  node_pool_max_nodes = var.spec.default_node_pool.max_nodes
}

