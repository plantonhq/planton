locals {
  # Resource identity -- pool name comes from metadata.name
  pool_name = var.metadata.name

  # Cluster reference (resolved from StringValueOrRef before Terraform runs)
  cluster_id = var.spec.cluster_id

  # Spec fields
  region    = var.spec.region
  node_type = var.spec.node_type
  size      = var.spec.size

  # Autoscaling
  auto_scale = var.spec.auto_scale
  min_size   = var.spec.min_size
  max_size   = var.spec.max_size

  # Node configuration
  autohealing       = var.spec.autohealing
  container_runtime = var.spec.container_runtime
  root_volume_type  = var.spec.root_volume_type
  root_volume_size  = var.spec.root_volume_size_in_gb
  public_ip_disabled = var.spec.public_ip_disabled

  # Placement
  zone               = var.spec.zone
  placement_group_id = var.spec.placement_group_id

  # Upgrade policy
  has_upgrade_policy = var.spec.upgrade_policy != null

  # ── Tag generation ───────────────────────────────────────────────────────
  #
  # Three categories of tags are merged into a single list:
  #
  # 1. Standard OpenMCF tags: "planton-ai_resource=true", etc.
  # 2. Kubernetes label tags: "noprefix={key}={value}"
  #    The Scaleway CCM syncs these to K8s node labels as {key}={value}.
  # 3. Kubernetes taint tags: "taint=noprefix={key}={value}:{Effect}"
  #    The CCM syncs these to K8s node taints as {key}={value}:{Effect}.
  #
  # We use the "noprefix=" variant so users get exactly the label/taint
  # keys they specified, without the default k8s.scaleway.com/ prefix.

  # 1. Standard OpenMCF tags
  standard_tags = compact([
    "planton-ai_resource=true",
    "planton-ai_name=${var.metadata.name}",
    "planton-ai_kind=ScalewayKapsulePool",
    var.metadata.org != null ? "planton-ai_org=${var.metadata.org}" : "",
    var.metadata.env != null ? "planton-ai_env=${var.metadata.env}" : "",
    var.metadata.id != null ? "planton-ai_id=${var.metadata.id}" : "",
  ])

  # 2. Kubernetes label tags
  label_tags = [
    for key, value in var.spec.kubernetes_labels :
    "noprefix=${key}=${value}"
  ]

  # 3. Kubernetes taint tags
  taint_tags = [
    for taint in var.spec.taints :
    "taint=noprefix=${taint.key}=${taint.value}:${taint.effect}"
  ]

  # Merged tag list
  all_tags = concat(local.standard_tags, local.label_tags, local.taint_tags)
}
