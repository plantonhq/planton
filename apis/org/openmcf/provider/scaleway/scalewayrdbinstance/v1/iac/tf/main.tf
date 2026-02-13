# ScalewayRdbInstance Terraform Module
#
# This module provisions a Scaleway Managed Database (RDB) instance with
# bundled databases, users, privileges, and ACL rules.
#
# Supported engines: PostgreSQL (14, 15, 16), MySQL (8)
#
# Bundled Terraform resources:
#   1. scaleway_rdb_instance    -- The managed database engine
#   2. scaleway_rdb_database    -- Logical databases (for_each)
#   3. scaleway_rdb_user        -- Database users (for_each)
#   4. scaleway_rdb_privilege   -- User-database permission grants (for_each)
#   5. scaleway_rdb_acl         -- Network access control (count 0 or 1)

# ── 1. RDB Instance ──────────────────────────────────────────────────────────

resource "scaleway_rdb_instance" "instance" {
  name      = local.instance_name
  engine    = local.engine
  node_type = local.node_type
  region    = local.region
  tags      = local.standard_tags

  # Admin user (created with the instance).
  user_name = local.admin_user
  password  = local.admin_password

  # High availability.
  is_ha_cluster = local.is_ha_cluster

  # Volume configuration.
  volume_type      = local.volume_type
  volume_size_in_gb = local.has_custom_volume ? local.volume_size_in_gb : null

  # Backup configuration.
  disable_backup            = local.disable_backup
  backup_schedule_frequency = local.has_custom_backup_frequency ? var.spec.backup_schedule_frequency_hours : null
  backup_schedule_retention = local.has_custom_backup_retention ? var.spec.backup_schedule_retention_days : null

  # Encryption at rest.
  encryption_at_rest = local.encryption_at_rest

  # Engine settings (applied on create and update).
  settings      = local.has_settings ? var.spec.settings : null
  init_settings = local.has_init_settings ? var.spec.init_settings : null

  # Optional Private Network attachment with IPAM.
  dynamic "private_network" {
    for_each = local.has_private_network ? [1] : []
    content {
      pn_id      = local.private_network_id
      enable_ipam = true
    }
  }

  # Lifecycle: password changes should not trigger replacement.
  lifecycle {
    ignore_changes = [password]
  }
}

# ── 2. Logical Databases ─────────────────────────────────────────────────────

resource "scaleway_rdb_database" "databases" {
  for_each = local.databases_map

  instance_id = scaleway_rdb_instance.instance.id
  name        = each.value.name
  region      = local.region
}

# ── 3. Additional Users ──────────────────────────────────────────────────────

resource "scaleway_rdb_user" "users" {
  for_each = local.users_map

  instance_id = scaleway_rdb_instance.instance.id
  name        = each.value.name
  password    = each.value.password
  is_admin    = each.value.is_admin
  region      = local.region
}

# ── 4. User-Database Privileges ──────────────────────────────────────────────
#
# Each privilege links a user to a database with a specific permission level.
# The for_each key is "user_name/database_name" for uniqueness.
# Privileges depend on both the user and (if managed by this module) the
# database existing first.

resource "scaleway_rdb_privilege" "privileges" {
  for_each = local.privileges_map

  instance_id   = scaleway_rdb_instance.instance.id
  user_name     = scaleway_rdb_user.users[each.value.user_name].name
  database_name = each.value.database_name
  permission    = each.value.permission
  region        = local.region

  depends_on = [
    scaleway_rdb_database.databases,
    scaleway_rdb_user.users,
  ]
}

# ── 5. Network ACL Rules ────────────────────────────────────────────────────
#
# Scaleway's ACL is a single resource per instance that replaces ALL rules
# atomically. Only created when the user specifies at least one ACL rule.
# If no rules are specified, no ACL resource is created and Scaleway's
# default applies (all IPs allowed on the public endpoint).

resource "scaleway_rdb_acl" "acl" {
  count = length(var.spec.acl_rules) > 0 ? 1 : 0

  instance_id = scaleway_rdb_instance.instance.id
  region      = local.region

  dynamic "acl_rules" {
    for_each = var.spec.acl_rules
    content {
      ip          = acl_rules.value.ip
      description = acl_rules.value.description
    }
  }
}
