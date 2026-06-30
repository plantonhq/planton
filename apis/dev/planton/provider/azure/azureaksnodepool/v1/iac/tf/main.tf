# Create Azure AKS Node Pool
resource "azurerm_kubernetes_cluster_node_pool" "node_pool" {
  name                  = local.node_pool_name
  kubernetes_cluster_id = data.azurerm_kubernetes_cluster.aks.id
  vm_size               = var.spec.vm_size
  node_count            = var.spec.initial_node_count

  # Autoscaling configuration
  enable_auto_scaling = var.spec.autoscaling != null
  min_count           = var.spec.autoscaling != null ? var.spec.autoscaling.min_nodes : null
  max_count           = var.spec.autoscaling != null ? var.spec.autoscaling.max_nodes : null

  # Availability zones
  zones = length(var.spec.availability_zones) > 0 ? var.spec.availability_zones : null

  # Mode and OS type
  mode    = local.mode
  os_type = local.os_type

  # Spot instance configuration
  priority        = var.spec.spot_enabled ? "Spot" : "Regular"
  eviction_policy = var.spec.spot_enabled ? "Delete" : null
  spot_max_price  = var.spec.spot_enabled ? -1 : null # -1 means pay up to regular price

  tags = local.tags

  lifecycle {
    ignore_changes = [
      tags["LastModified"],
    ]
  }
}

# Data source to look up the parent AKS cluster
data "azurerm_kubernetes_cluster" "aks" {
  name                = local.cluster_name
  resource_group_name = local.resource_group_name
}

