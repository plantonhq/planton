locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The TPU's id defaults to metadata.name -- identical to the Pulumi module.
  node_id = var.spec.node_id != "" ? var.spec.node_id : var.metadata.name

  accelerator_type = var.spec.accelerator_type != "" ? var.spec.accelerator_type : null
  description      = var.spec.description != "" ? var.spec.description : null
  cidr_block       = var.spec.cidr_block != "" ? var.spec.cidr_block : null
  metadata         = length(var.spec.metadata) > 0 ? var.spec.metadata : null
  tags             = length(var.spec.tags) > 0 ? var.spec.tags : null
  deletion_policy  = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Compute references (a GcpSubnetwork, a GcpComputeDisk) arrive as
  # self-links; Cloud TPU takes the relative path. Trimming the prefix leaves
  # a literal relative path as is -- the same rule as the Pulumi module.
  compute_prefix = "https://www.googleapis.com/compute/v1/"

  network_configs = [
    for nc in(var.spec.network_config != null ? [var.spec.network_config] : var.spec.network_configs) : {
      network             = nc.network != "" ? nc.network : null
      subnetwork          = nc.subnetwork != "" ? trimprefix(nc.subnetwork, local.compute_prefix) : null
      enable_external_ips = nc.enable_external_ips
      can_ip_forward      = nc.can_ip_forward
      queue_count         = nc.queue_count != 0 ? nc.queue_count : null
    }
  ]

  # The same planton-ai_* label set the Pulumi module applies. User labels
  # merge in first so the platform attribution labels can never be
  # clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcptpuvm"
  }

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "planton-ai_organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "planton-ai_environment" = var.metadata.env } : {}

  id_label = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "planton-ai_id" = var.metadata.id } : {}

  final_labels = merge(var.spec.labels, local.base_labels, local.org_label, local.env_label, local.id_label)
}
