locals {
  # Resource identity
  cluster_name = var.metadata.name

  # Spec fields
  region             = var.spec.region
  kubernetes_version = var.spec.kubernetes_version
  cni                = var.spec.cni
  private_network_id = var.spec.private_network_id
  cluster_type       = var.spec.type
  description        = var.spec.description
  delete_additional  = var.spec.delete_additional_resources

  # Optional networking overrides (ForceNew)
  pod_cidr     = var.spec.pod_cidr
  service_cidr = var.spec.service_cidr

  # Feature gates and admission plugins
  feature_gates     = var.spec.feature_gates
  admission_plugins = var.spec.admission_plugins

  # Auto-upgrade: only create the block when the spec provides it
  has_auto_upgrade = var.spec.auto_upgrade != null

  # Autoscaler config: only create the block when the spec provides it
  has_autoscaler_config = var.spec.autoscaler_config != null

  # Default node pool fields
  pool_name = (
    var.spec.default_node_pool.name != ""
    ? var.spec.default_node_pool.name
    : "${var.metadata.name}-default"
  )
  pool_node_type          = var.spec.default_node_pool.node_type
  pool_size               = var.spec.default_node_pool.size
  pool_auto_scale         = var.spec.default_node_pool.auto_scale
  pool_min_size           = var.spec.default_node_pool.min_size
  pool_max_size           = var.spec.default_node_pool.max_size
  pool_autohealing        = var.spec.default_node_pool.autohealing
  pool_container_runtime  = var.spec.default_node_pool.container_runtime
  pool_root_volume_type   = var.spec.default_node_pool.root_volume_type
  pool_root_volume_size   = var.spec.default_node_pool.root_volume_size_in_gb
  pool_public_ip_disabled = var.spec.default_node_pool.public_ip_disabled
  pool_has_upgrade_policy = var.spec.default_node_pool.upgrade_policy != null

  # Standard OpenMCF tags formatted as flat strings for Scaleway
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayKapsuleCluster",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])
}
