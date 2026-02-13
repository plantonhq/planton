locals {
  # ── Resource identity ──────────────────────────────────────────────
  instance_name = var.metadata.name
  region        = var.spec.region

  # ── Core configuration ─────────────────────────────────────────────
  version     = var.spec.version
  node_type   = var.spec.node_type
  node_number = var.spec.node_number

  # ── Networking ─────────────────────────────────────────────────────
  private_network_id    = var.spec.private_network_id
  has_private_network   = local.private_network_id != ""
  enable_public_network = var.spec.enable_public_network

  # ── Storage ────────────────────────────────────────────────────────
  volume_type         = var.spec.volume_type
  has_custom_volume   = var.spec.volume_size_in_gb > 0
  volume_size_in_gb   = var.spec.volume_size_in_gb

  # ── Snapshot schedule ──────────────────────────────────────────────
  enable_snapshot_schedule = var.spec.enable_snapshot_schedule
  has_custom_snapshot_frequency = var.spec.snapshot_schedule_frequency_hours > 0
  has_custom_snapshot_retention = var.spec.snapshot_schedule_retention_days > 0

  # ── Admin user ─────────────────────────────────────────────────────
  admin_user     = var.spec.admin_user
  admin_password = var.spec.admin_password

  # ── Settings ───────────────────────────────────────────────────────
  has_settings = length(var.spec.settings) > 0

  # ── Users map (keyed by name for for_each) ─────────────────────────
  users_map = { for user in var.spec.users : user.name => user }

  # ── Standard OpenMCF tags ──────────────────────────────────────────
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayMongodbInstance",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])
}
