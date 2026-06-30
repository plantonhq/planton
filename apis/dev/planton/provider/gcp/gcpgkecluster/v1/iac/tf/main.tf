#############################################
# GKE Cluster (Control Plane)
#############################################

resource "google_container_cluster" "cluster" {
  name     = var.spec.cluster_name
  project  = var.spec.project_id.value
  location = var.spec.location

  # VPC-native networking configuration
  network    = var.spec.network_self_link.value
  subnetwork = var.spec.subnetwork_self_link.value

  # Remove default node pool immediately (best practice)
  # Node pools should be managed separately via GcpGkeNodePool resources
  remove_default_node_pool = true
  initial_node_count       = 1

  # Deletion protection disabled to allow IaC-managed lifecycle
  deletion_protection = false

  # Private cluster configuration
  private_cluster_config {
    # Invert enable_public_nodes: if public nodes requested, disable private nodes
    enable_private_nodes    = !var.spec.enable_public_nodes
    enable_private_endpoint = false  # API server remains accessible from VPC
    master_ipv4_cidr_block  = var.spec.master_ipv4_cidr_block
  }

  # VPC-native IP allocation (required for modern GKE features)
  ip_allocation_policy {
    cluster_secondary_range_name  = var.spec.cluster_secondary_range_name.value
    services_secondary_range_name = var.spec.services_secondary_range_name.value
  }

  # Workload Identity configuration (IAM for pods)
  dynamic "workload_identity_config" {
    for_each = var.spec.disable_workload_identity ? [] : [1]
    content {
      workload_pool = "${var.spec.project_id.value}.svc.id.goog"
    }
  }

  # Add-ons configuration
  addons_config {
    # Network policy enforcement (Calico for microsegmentation)
    network_policy_config {
      disabled = var.spec.disable_network_policy
    }
  }

  # Release channel for auto-upgrades
  release_channel {
    channel = local.release_channel
  }

  # Resource labels
  resource_labels = local.gcp_labels
}

