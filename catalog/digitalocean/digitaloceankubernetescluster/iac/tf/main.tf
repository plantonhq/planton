# DigitalOcean Managed Kubernetes Cluster (DOKS)
#
# Provisions a managed Kubernetes cluster on DigitalOcean, modeling the
# complete digitalocean_kubernetes_cluster resource surface: version/region/
# VPC placement, the inline default node pool (labels, taints, tags,
# autoscaling, GPU partitioning), HA control plane, surge and auto upgrades,
# maintenance policy, control-plane firewall, pod/service/worker subnets,
# isolated workers, SSO, cluster-autoscaler tuning, registry integration,
# kubeconfig expiry, destroy-time cleanup, and the managed addon toggles.
#
# Additional node pools are separate digitalocean_kubernetes_node_pool
# resources with independent lifecycles and are not part of this module.

resource "digitalocean_kubernetes_cluster" "cluster" {
  name    = var.spec.cluster_name
  region  = var.spec.region
  version = var.spec.kubernetes_version

  # References are resolved to the literal VPC UUID before the module runs,
  # so the field arrives as a plain string (create-only).
  vpc_uuid = var.spec.vpc

  # HA is one-way: once enabled it cannot be turned off. An explicit false
  # (proto3 unset) keeps the cheaper single-replica control plane even on
  # DOKS versions whose server-side default is HA on.
  ha = var.spec.highly_available

  auto_upgrade = var.spec.auto_upgrade

  # null when unset in the spec: the provider's default (true) then applies,
  # matching DigitalOcean's own surge-upgrade default. Never coalesce to
  # false -- that would silently flip the default off.
  surge_upgrade = var.spec.surge_upgrade

  registry_integration = var.spec.registry_integration

  # Pod / service / worker network placement (all create-only). Empty means
  # DigitalOcean assigns them; the provider rejects "" for the CIDRs, so
  # unset must arrive as null.
  cluster_subnet     = local.cluster_subnet
  service_subnet     = local.service_subnet
  worker_subnet_uuid = local.worker_subnet_uuid
  isolated_workers   = var.spec.isolated_workers

  # Destroy-time behavior flag: also deletes the load balancers, volumes,
  # and volume snapshots the cluster created. Never sent to the API.
  destroy_all_associated_resources = var.spec.destroy_all_associated_resources

  # 0 (unset) means DigitalOcean's 7-day default credential validity.
  kubeconfig_expire_seconds = local.kubeconfig_expire_seconds

  # Weekly maintenance window. Day is lowercased because the provider
  # accepts any case but reads back lowercase -- mixed case would drift.
  dynamic "maintenance_policy" {
    for_each = var.spec.maintenance_policy != null ? [var.spec.maintenance_policy] : []
    content {
      day        = lower(maintenance_policy.value.day)
      start_time = maintenance_policy.value.start_time
    }
  }

  # Firewall in front of the public API server endpoint. Both leaves are
  # required by the provider's block schema.
  dynamic "control_plane_firewall" {
    for_each = var.spec.control_plane_firewall != null ? [var.spec.control_plane_firewall] : []
    content {
      enabled           = control_plane_firewall.value.enabled
      allowed_addresses = control_plane_firewall.value.allowed_addresses
    }
  }

  # Tuning for the DigitalOcean-managed cluster-autoscaler.
  dynamic "cluster_autoscaler_configuration" {
    for_each = var.spec.cluster_autoscaler_configuration != null ? [var.spec.cluster_autoscaler_configuration] : []
    content {
      scale_down_utilization_threshold = cluster_autoscaler_configuration.value.scale_down_utilization_threshold
      scale_down_unneeded_time         = cluster_autoscaler_configuration.value.scale_down_unneeded_time != "" ? cluster_autoscaler_configuration.value.scale_down_unneeded_time : null
      expanders                        = length(cluster_autoscaler_configuration.value.expanders) > 0 ? cluster_autoscaler_configuration.value.expanders : null
    }
  }

  # OpenID Connect single sign-on for the Kubernetes API.
  dynamic "sso" {
    for_each = var.spec.sso != null ? [var.spec.sso] : []
    content {
      enabled    = sso.value.enabled
      required   = sso.value.required
      issuer_url = sso.value.issuer_url != "" ? sso.value.issuer_url : null
      client_id  = sso.value.client_id != "" ? sso.value.client_id : null
    }
  }

  # Managed addon toggles. An unset spec message emits no block, deferring
  # to DigitalOcean's own default for that addon; a set message asserts the
  # desired state, on or off. The GPU-family blocks are accepted by the API
  # only on clusters with GPU node pools, and the P2P OCI registry plugin
  # only on Kubernetes 1.36.0-do.2 or later -- an older version fails the
  # whole create with a validation 422 that creates nothing. The module
  # sends what the manifest states and lets DigitalOcean's own validation
  # speak, because the version floor moves with DigitalOcean's release train
  # and a module-side check would go stale.
  dynamic "routing_agent" {
    for_each = var.spec.routing_agent != null ? [var.spec.routing_agent] : []
    content {
      enabled = routing_agent.value.enabled
    }
  }

  dynamic "p2p_oci_registry_plugin" {
    for_each = var.spec.p2p_oci_registry_plugin != null ? [var.spec.p2p_oci_registry_plugin] : []
    content {
      enabled = p2p_oci_registry_plugin.value.enabled
    }
  }

  dynamic "amd_gpu_device_plugin" {
    for_each = var.spec.amd_gpu_device_plugin != null ? [var.spec.amd_gpu_device_plugin] : []
    content {
      enabled = amd_gpu_device_plugin.value.enabled
    }
  }

  dynamic "amd_gpu_dra_driver" {
    for_each = var.spec.amd_gpu_dra_driver != null ? [var.spec.amd_gpu_dra_driver] : []
    content {
      enabled = amd_gpu_dra_driver.value.enabled
    }
  }

  dynamic "amd_gpu_device_metrics_exporter_plugin" {
    for_each = var.spec.amd_gpu_device_metrics_exporter_plugin != null ? [var.spec.amd_gpu_device_metrics_exporter_plugin] : []
    content {
      enabled = amd_gpu_device_metrics_exporter_plugin.value.enabled
    }
  }

  dynamic "nvidia_gpu_device_plugin" {
    for_each = var.spec.nvidia_gpu_device_plugin != null ? [var.spec.nvidia_gpu_device_plugin] : []
    content {
      enabled = nvidia_gpu_device_plugin.value.enabled
    }
  }

  dynamic "nvidia_gpu_dra_driver" {
    for_each = var.spec.nvidia_gpu_dra_driver != null ? [var.spec.nvidia_gpu_dra_driver] : []
    content {
      enabled = nvidia_gpu_dra_driver.value.enabled
    }
  }

  dynamic "rdma_shared_device_plugin" {
    for_each = var.spec.rdma_shared_device_plugin != null ? [var.spec.rdma_shared_device_plugin] : []
    content {
      enabled = rdma_shared_device_plugin.value.enabled
    }
  }

  dynamic "coredns_autoscaler" {
    for_each = var.spec.coredns_autoscaler != null ? [var.spec.coredns_autoscaler] : []
    content {
      enabled = coredns_autoscaler.value.enabled
    }
  }

  # The inline default node pool. Its name is synthesized -- the pool has no
  # independent identity. Changing size or gpu_partition_mode replaces the
  # ENTIRE cluster (provider ForceNew inside this block).
  node_pool {
    name = "default"
    size = var.spec.default_node_pool.size
    # Exactly one sizing mode owns the count -- matching the Pulumi module.
    # A fixed pool sends node_count; an autoscaled pool sends only the
    # bounds and NO count, because the provider writes the live count back
    # into node_count on every read and re-applies a stated one on every
    # update, so a stated count and the autoscaler would fight forever
    # (measured: a pool that autoscaled to two nodes planned `2 -> 1`).
    # Without a count the API starts the pool at min_nodes.
    node_count = var.spec.default_node_pool.auto_scale ? null : var.spec.default_node_pool.node_count
    auto_scale = var.spec.default_node_pool.auto_scale
    min_nodes  = var.spec.default_node_pool.auto_scale ? var.spec.default_node_pool.min_nodes : null
    max_nodes  = var.spec.default_node_pool.auto_scale ? var.spec.default_node_pool.max_nodes : null

    # Kubernetes node labels: user labels over the standard Planton labels
    # (identical set in both provisioners).
    labels = local.default_node_pool_labels

    # Pool-level Droplet tags. The provider itself appends the
    # terraform:default-node-pool marker tag; never add it here.
    tags = length(var.spec.default_node_pool.tags) > 0 ? var.spec.default_node_pool.tags : null

    gpu_partition_mode = var.spec.default_node_pool.gpu_partition_mode != "" ? var.spec.default_node_pool.gpu_partition_mode : null

    dynamic "taint" {
      for_each = var.spec.default_node_pool.taints
      content {
        key    = taint.value.key
        value  = taint.value.value
        effect = taint.value.effect
      }
    }
  }

  # User tags plus the standard Planton labels (identical set in both
  # provisioners).
  tags = local.tags

  lifecycle {
    # Auto-upgrade moves the live version ahead of the configured pin, and
    # the provider DESTROYS AND RECREATES the cluster when the configured
    # version is lower than the live one. Ignoring version drift makes the
    # pin creation-only; upgrades ride auto_upgrade inside the maintenance
    # window.
    ignore_changes = [version]
  }
}
