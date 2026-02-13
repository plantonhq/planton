# ── Scaleway Kapsule Node Pool ─────────────────────────────────────────────────
#
# Creates an additional node pool in an existing Kapsule Kubernetes cluster.
#
# Kubernetes labels and taints are applied via Scaleway's Cloud Controller
# Manager (CCM) tag convention. The tag generation logic in locals.tf produces:
#   - Label tags: "noprefix={key}={value}" → K8s label {key}={value}
#   - Taint tags: "taint=noprefix={key}={value}:{Effect}" → K8s taint
#
# These tags are merged with standard OpenMCF tags and passed to the pool's
# `tags` field. The CCM automatically syncs them to Kubernetes nodes.
resource "scaleway_k8s_pool" "pool" {
  cluster_id = local.cluster_id
  name       = local.pool_name
  node_type  = local.node_type
  size       = local.size
  tags       = local.all_tags
  region     = local.region

  # Autoscaling configuration
  autoscaling = local.auto_scale
  min_size    = local.auto_scale ? local.min_size : null
  max_size    = local.auto_scale ? local.max_size : null

  # Node health
  autohealing = local.autohealing

  # Container runtime
  container_runtime = local.container_runtime

  # Root volume configuration
  root_volume_type       = local.root_volume_type
  root_volume_size_in_gb = local.root_volume_size

  # Network isolation
  public_ip_disabled = local.public_ip_disabled

  # Zone placement (optional)
  zone = local.zone != null && local.zone != "" ? local.zone : null

  # Anti-affinity placement group (optional)
  placement_group_id = local.placement_group_id != null && local.placement_group_id != "" ? local.placement_group_id : null

  # Custom kubelet arguments (optional)
  kubelet_args = length(var.spec.kubelet_args) > 0 ? var.spec.kubelet_args : null

  # Wait for pool to be ready before marking complete
  wait_for_pool_ready = true

  # Upgrade policy (optional)
  dynamic "upgrade_policy" {
    for_each = local.has_upgrade_policy ? [var.spec.upgrade_policy] : []
    content {
      max_surge       = upgrade_policy.value.max_surge
      max_unavailable = upgrade_policy.value.max_unavailable
    }
  }
}
