# ScalewayRedisCluster Terraform Module
#
# This module provisions a Scaleway Managed Redis cluster with optional
# ACL rules or Private Network attachment (mutually exclusive).
#
# Supported deployment modes (determined by cluster_size):
#   - cluster_size = 1: Standalone (single node)
#   - cluster_size = 2: HA (1 main + 1 standby)
#   - cluster_size >= 3: Cluster mode (sharding)
#
# Networking:
#   - ACL rules: Allow specific CIDR ranges on the public endpoint
#   - Private Network: Private endpoint only, no public access
#   - ACL and Private Network CANNOT be used simultaneously

resource "scaleway_redis_cluster" "cluster" {
  name      = local.cluster_name
  version   = local.version
  node_type = local.node_type
  zone      = local.zone
  tags      = local.standard_tags

  # Authentication.
  user_name = local.user_name
  password  = local.password

  # Cluster sizing.
  cluster_size = local.cluster_size

  # TLS encryption (forces recreation if changed).
  tls_enabled = local.tls_enabled

  # ACL rules (mutually exclusive with private_network).
  dynamic "acl" {
    for_each = local.acl_rules
    content {
      ip          = acl.value.ip
      description = acl.value.description
    }
  }

  # Private Network attachment (mutually exclusive with acl).
  dynamic "private_network" {
    for_each = local.has_private_network ? [1] : []
    content {
      id = local.private_network_id
      # IPAM assigns IPs automatically when service_ips is omitted.
    }
  }

  # Redis settings.
  settings = local.has_settings ? var.spec.settings : null

  # Lifecycle: password changes should not trigger replacement.
  lifecycle {
    ignore_changes = [password]
  }
}
