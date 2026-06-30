# DigitalOcean Kubernetes Node Pool Resource
resource "digitalocean_kubernetes_node_pool" "node_pool" {
  cluster_id = local.cluster_id
  name       = local.node_pool_name
  size       = local.size
  node_count = local.node_count
  
  # Autoscaling configuration (optional)
  auto_scale = local.auto_scale
  min_nodes  = local.auto_scale ? local.min_nodes : null
  max_nodes  = local.auto_scale ? local.max_nodes : null
  
  # Labels - for Kubernetes scheduling
  labels = local.all_labels
  
  # Taints - for workload isolation
  dynamic "taint" {
    for_each = local.taints
    content {
      key    = taint.value.key
      value  = taint.value.value
      effect = taint.value.effect
    }
  }
  
  # Tags - for DigitalOcean billing and organization
  tags = local.all_tags
}

