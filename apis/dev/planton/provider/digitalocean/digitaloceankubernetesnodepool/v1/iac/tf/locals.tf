locals {
  # Node pool name from spec
  node_pool_name = var.spec.node_pool_name
  
  # Cluster ID from spec
  cluster_id = var.spec.cluster.value
  
  # Droplet size
  size = var.spec.size
  
  # Node count
  node_count = var.spec.node_count
  
  # Autoscaling configuration
  auto_scale = var.spec.auto_scale
  min_nodes  = var.spec.min_nodes
  max_nodes  = var.spec.max_nodes
  
  # Merge metadata labels with spec labels
  all_labels = merge(
    {
      "planton-resource"      = "true"
      "planton-resource-name" = var.metadata.name
      "planton-resource-kind" = "DigitalOceanKubernetesNodePool"
    },
    var.metadata.org != null ? { "planton-organization" = var.metadata.org } : {},
    var.metadata.env != null ? { "planton-environment" = var.metadata.env } : {},
    var.metadata.id != null ? { "planton-resource-id" = var.metadata.id } : {},
    var.spec.labels
  )
  
  # Tags - merge metadata tags with spec tags
  all_tags = concat(
    var.metadata.tags != null ? var.metadata.tags : [],
    var.spec.tags
  )
  
  # Taints from spec
  taints = var.spec.taints
}

