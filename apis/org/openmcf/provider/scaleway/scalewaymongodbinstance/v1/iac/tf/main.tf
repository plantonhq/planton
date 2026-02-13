# ScalewayMongodbInstance Terraform Module
#
# This module provisions a Scaleway Managed MongoDB instance with
# bundled database users and role-based access control.
#
# Bundled Terraform resources:
#   1. scaleway_mongodb_instance  -- The managed MongoDB engine
#   2. scaleway_mongodb_user      -- Additional database users (for_each)
#
# Key differences from ScalewayRdbInstance:
#   - No database creation (MongoDB creates databases implicitly)
#   - No ACL resource (no IP-based access control, only PN/public toggle)
#   - No privilege resource (roles are inline on users)
#   - HA via node_number=3 (replica set), not a boolean flag

# ── 1. MongoDB Instance ──────────────────────────────────────────────────────

resource "scaleway_mongodb_instance" "instance" {
  name        = local.instance_name
  version     = local.version
  node_type   = local.node_type
  node_number = local.node_number
  region      = local.region
  tags        = local.standard_tags

  # Admin user (created with the instance).
  user_name = local.admin_user
  password  = local.admin_password

  # Volume configuration.
  volume_type      = local.volume_type
  volume_size_in_gb = local.has_custom_volume ? local.volume_size_in_gb : null

  # Snapshot schedule configuration.
  is_snapshot_schedule_enabled      = local.enable_snapshot_schedule
  snapshot_schedule_frequency_hours = local.has_custom_snapshot_frequency ? var.spec.snapshot_schedule_frequency_hours : null
  snapshot_schedule_retention_days  = local.has_custom_snapshot_retention ? var.spec.snapshot_schedule_retention_days : null

  # MongoDB settings.
  settings = local.has_settings ? var.spec.settings : null

  # Optional Private Network attachment with IPAM.
  dynamic "private_network" {
    for_each = local.has_private_network ? [1] : []
    content {
      pn_id = local.private_network_id
    }
  }

  # Optional Public Network endpoint.
  # Added when:
  #   - PN is attached AND user explicitly wants public endpoint too.
  #   - PN is NOT attached (public by default -- Scaleway behavior).
  #     When no PN, we don't add the block; Scaleway auto-creates public.
  dynamic "public_network" {
    for_each = local.has_private_network && local.enable_public_network ? [1] : []
    content {}
  }

  # Lifecycle: password changes should not trigger replacement.
  lifecycle {
    ignore_changes = [password]
  }
}

# ── 2. Additional Users ──────────────────────────────────────────────────────
#
# Each user is a separate `scaleway_mongodb_user` resource with inline role
# assignments. Roles scope permissions to specific databases or all databases.

resource "scaleway_mongodb_user" "users" {
  for_each = local.users_map

  instance_id = scaleway_mongodb_instance.instance.id
  name        = each.value.name
  password    = each.value.password
  region      = local.region

  # Role assignments (inline on the user resource).
  dynamic "roles" {
    for_each = each.value.roles
    content {
      role          = roles.value.role
      database_name = roles.value.database_name != "" ? roles.value.database_name : null
      any_database  = roles.value.any_database ? true : null
    }
  }
}
