locals {
  # ── Resource identity ──────────────────────────────────────────────
  cluster_name = var.metadata.name
  zone         = var.spec.zone

  # ── Core configuration ─────────────────────────────────────────────
  version   = var.spec.version
  node_type = var.spec.node_type

  # ── Cluster sizing ─────────────────────────────────────────────────
  cluster_size = var.spec.cluster_size

  # ── Security ───────────────────────────────────────────────────────
  tls_enabled = var.spec.tls_enabled

  # ── Authentication ─────────────────────────────────────────────────
  user_name = var.spec.user_name
  password  = var.spec.password

  # ── Networking ─────────────────────────────────────────────────────
  # ACL and Private Network are mutually exclusive.
  private_network_id  = var.spec.private_network_id
  has_private_network = local.private_network_id != ""

  # ACL rules: only used when NOT on a Private Network.
  acl_rules = var.spec.acl_rules

  # ── Redis settings ─────────────────────────────────────────────────
  has_settings = length(var.spec.settings) > 0

  # ── Standard Planton tags ──────────────────────────────────────────
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayRedisCluster",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])
}
