# DigitalOcean Kubernetes Cluster Resource
resource "digitalocean_kubernetes_cluster" "cluster" {
  name    = local.cluster_name
  region  = local.region
  version = local.kubernetes_version
  
  # VPC configuration
  vpc_uuid = local.vpc_uuid
  
  # High availability control plane
  ha = local.highly_available
  
  # Auto-upgrade settings
  auto_upgrade  = local.auto_upgrade
  surge_upgrade = local.surge_upgrade
  
  # Registry integration
  registry_integration = local.registry_integration
  
  # Maintenance window (optional)
  dynamic "maintenance_policy" {
    for_each = local.maintenance_window != null ? [1] : []
    content {
      start_time = local.maintenance_window
    }
  }
  
  # Control plane firewall (optional)
  dynamic "firewall" {
    for_each = local.has_firewall_ips ? [1] : []
    content {
      allowed_addresses = local.firewall_ips
    }
  }
  
  # Default node pool
  node_pool {
    name       = local.node_pool_name
    size       = local.node_pool_size
    node_count = local.node_pool_count
    auto_scale = local.node_pool_auto_scale
    min_nodes  = local.node_pool_auto_scale ? local.node_pool_min_nodes : null
    max_nodes  = local.node_pool_auto_scale ? local.node_pool_max_nodes : null
  }
  
  # Tags
  tags = local.all_tags
  
  # Lifecycle settings
  lifecycle {
    # Prevent accidental deletion of production clusters
    prevent_destroy = false
    
    # Ignore changes to version to prevent drift on auto-upgrades
    ignore_changes = [
      version
    ]
  }
}

