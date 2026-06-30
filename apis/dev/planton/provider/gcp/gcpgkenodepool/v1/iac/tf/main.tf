###############################################################################
# Data Sources
###############################################################################

# Lookup the parent GKE cluster to get location information
data "google_container_cluster" "cluster" {
  name     = var.spec.cluster_name.value
  project  = var.spec.cluster_project_id.value
  location = "*" # Wildcard to search all locations
}

###############################################################################
# GKE Node Pool
###############################################################################

resource "google_container_node_pool" "node_pool" {
  name     = var.spec.node_pool_name
  cluster  = data.google_container_cluster.cluster.name
  location = data.google_container_cluster.cluster.location
  project  = var.spec.cluster_project_id.value

  # Node count: either fixed or managed by autoscaler
  # If autoscaling is enabled, node_count should be omitted or set to null
  node_count = var.spec.autoscaling != null ? null : var.spec.node_count

  # Autoscaling configuration (mutually exclusive with fixed node_count)
  dynamic "autoscaling" {
    for_each = var.spec.autoscaling != null ? [var.spec.autoscaling] : []
    content {
      min_node_count  = autoscaling.value.min_nodes
      max_node_count  = autoscaling.value.max_nodes
      location_policy = autoscaling.value.location_policy
    }
  }

  # Node configuration
  node_config {
    machine_type = var.spec.machine_type
    disk_size_gb = var.spec.disk_size_gb
    disk_type    = var.spec.disk_type
    image_type   = var.spec.image_type

    # Spot VMs (preemptible)
    preemptible = var.spec.spot
    spot        = var.spec.spot

    # Service account for nodes
    service_account = var.spec.service_account != "" ? var.spec.service_account : null

    # OAuth scopes
    oauth_scopes = local.oauth_scopes

    # Labels applied to all nodes in the pool
    labels = local.merged_node_labels

    # Network tags (for firewall rules)
    tags = [local.network_tag]

    # Metadata
    metadata = {
      "disable-legacy-endpoints" = "true"
    }
  }

  # Node management: auto-upgrade and auto-repair
  management {
    auto_upgrade = local.auto_upgrade_enabled
    auto_repair  = local.auto_repair_enabled
  }

  # Upgrade settings: controls how nodes are upgraded
  upgrade_settings {
    max_surge       = 2
    max_unavailable = 1
  }

  # Ignore changes to node_count if autoscaling is enabled
  # This prevents Terraform from trying to reset the count when the autoscaler changes it
  lifecycle {
    ignore_changes = [
      node_count
    ]
  }
}

