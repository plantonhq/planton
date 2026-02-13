# ── 1. Kapsule Cluster (Managed Kubernetes Control Plane) ──────────────────────
#
# Creates the Kapsule cluster with the specified CNI, Kubernetes version,
# and Private Network attachment. The control plane is fully managed by
# Scaleway -- users interact with it via the Kubernetes API.
resource "scaleway_k8s_cluster" "cluster" {
  name                      = local.cluster_name
  version                   = local.kubernetes_version
  cni                       = local.cni
  private_network_id        = local.private_network_id
  type                      = local.cluster_type
  delete_additional_resources = local.delete_additional
  tags                      = local.standard_tags
  region                    = local.region

  # Description (optional).
  description = local.description != "" ? local.description : null

  # Feature gates (optional).
  feature_gates = length(local.feature_gates) > 0 ? local.feature_gates : null

  # Admission plugins (optional).
  admission_plugins = length(local.admission_plugins) > 0 ? local.admission_plugins : null

  # Pod CIDR (optional, ForceNew).
  pod_cidr = local.pod_cidr

  # Service CIDR (optional, ForceNew).
  service_cidr = local.service_cidr

  # Auto-upgrade configuration (optional).
  # Only created when spec.auto_upgrade is set.
  dynamic "auto_upgrade" {
    for_each = local.has_auto_upgrade ? [var.spec.auto_upgrade] : []
    content {
      enable                        = auto_upgrade.value.enable
      maintenance_window_start_hour = auto_upgrade.value.maintenance_window_start_hour
      maintenance_window_day        = auto_upgrade.value.maintenance_window_day
    }
  }

  # Cluster-wide autoscaler configuration (optional).
  # Only created when spec.autoscaler_config is set.
  dynamic "autoscaler_config" {
    for_each = local.has_autoscaler_config ? [var.spec.autoscaler_config] : []
    content {
      disable_scale_down               = autoscaler_config.value.disable_scale_down
      scale_down_delay_after_add       = autoscaler_config.value.scale_down_delay_after_add
      scale_down_unneeded_time         = autoscaler_config.value.scale_down_unneeded_time
      estimator                        = autoscaler_config.value.estimator
      expander                         = autoscaler_config.value.expander
      scale_down_utilization_threshold = autoscaler_config.value.scale_down_utilization_threshold
      max_graceful_termination_sec     = autoscaler_config.value.max_graceful_termination_sec
      ignore_daemonsets_utilization    = autoscaler_config.value.ignore_daemonsets_utilization
      balance_similar_node_groups      = autoscaler_config.value.balance_similar_node_groups
      expendable_pods_priority_cutoff  = autoscaler_config.value.expendable_pods_priority_cutoff
    }
  }

  # Lifecycle management:
  # Ignore version changes to prevent drift when auto-upgrade is enabled.
  # When Scaleway patches the cluster, the IaC state should not revert it.
  lifecycle {
    ignore_changes = [version]
  }
}

# ── 2. Default Node Pool ──────────────────────────────────────────────────────
#
# The embedded default pool that ships with the cluster. Provides immediate
# compute capacity so the cluster is usable from a single resource.
#
# Additional node pools with different instance types, labels, or taints
# should be created as separate ScalewayKapsulePool resources.
resource "scaleway_k8s_pool" "default" {
  cluster_id        = scaleway_k8s_cluster.cluster.id
  name              = local.pool_name
  node_type         = local.pool_node_type
  size              = local.pool_size
  tags              = local.standard_tags
  region            = local.region

  # Autoscaling configuration.
  autoscaling = local.pool_auto_scale
  min_size    = local.pool_auto_scale ? local.pool_min_size : null
  max_size    = local.pool_auto_scale ? local.pool_max_size : null

  # Autohealing.
  autohealing = local.pool_autohealing

  # Container runtime.
  container_runtime = local.pool_container_runtime

  # Root volume configuration.
  root_volume_type    = local.pool_root_volume_type
  root_volume_size_in_gb = local.pool_root_volume_size

  # Disable public IPs on nodes.
  public_ip_disabled = local.pool_public_ip_disabled

  # Wait for pool to be ready before marking complete.
  wait_for_pool_ready = true

  # Upgrade policy (optional).
  dynamic "upgrade_policy" {
    for_each = local.pool_has_upgrade_policy ? [var.spec.default_node_pool.upgrade_policy] : []
    content {
      max_surge       = upgrade_policy.value.max_surge
      max_unavailable = upgrade_policy.value.max_unavailable
    }
  }
}
