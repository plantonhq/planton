locals {
  # ── Resource identity ──────────────────────────────────────────────
  instance_name = var.metadata.name
  region        = var.spec.region

  # ── Core configuration ─────────────────────────────────────────────
  engine    = var.spec.engine
  node_type = var.spec.node_type

  # ── Networking ─────────────────────────────────────────────────────
  private_network_id    = var.spec.private_network_id
  has_private_network   = local.private_network_id != ""

  # ── High availability ──────────────────────────────────────────────
  is_ha_cluster = var.spec.is_ha_cluster

  # ── Storage ────────────────────────────────────────────────────────
  volume_type         = var.spec.volume_type
  has_custom_volume   = var.spec.volume_size_in_gb > 0
  volume_size_in_gb   = var.spec.volume_size_in_gb

  # ── Backup ─────────────────────────────────────────────────────────
  disable_backup = var.spec.disable_backup
  has_custom_backup_frequency = var.spec.backup_schedule_frequency_hours > 0
  has_custom_backup_retention = var.spec.backup_schedule_retention_days > 0

  # ── Security ───────────────────────────────────────────────────────
  encryption_at_rest = var.spec.encryption_at_rest

  # ── Admin user ─────────────────────────────────────────────────────
  admin_user     = var.spec.admin_user
  admin_password = var.spec.admin_password

  # ── Engine settings ────────────────────────────────────────────────
  has_settings      = length(var.spec.settings) > 0
  has_init_settings = length(var.spec.init_settings) > 0

  # ── Databases map (keyed by name for for_each) ─────────────────────
  databases_map = { for db in var.spec.databases : db.name => db }

  # ── Users map (keyed by name for for_each) ─────────────────────────
  users_map = { for user in var.spec.users : user.name => user }

  # ── Privileges map (flattened from users, keyed by "user/db") ──────
  # Flattens all user privilege entries into a single map suitable for
  # for_each. The key "user_name/database_name" ensures uniqueness
  # (Scaleway enforces one privilege per user per database).
  privileges_map = merge([
    for user in var.spec.users : {
      for priv in user.privileges :
      "${user.name}/${priv.database_name}" => {
        user_name     = user.name
        database_name = priv.database_name
        permission    = priv.permission
      }
    }
  ]...)

  # ── Standard Planton tags ──────────────────────────────────────────
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayRdbInstance",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])
}
