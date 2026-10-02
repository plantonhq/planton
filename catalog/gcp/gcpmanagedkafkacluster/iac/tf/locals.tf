locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The cluster's GCP id defaults to metadata.name -- the same fallback the
  # Pulumi module applies.
  cluster_id = var.spec.cluster_id != "" ? var.spec.cluster_id : var.metadata.name

  # A GcpSubnetwork reference resolves to the subnet's self link; Google
  # wants the projects/{p}/regions/{r}/subnetworks/{s} form, so the API
  # prefix is trimmed (a literal in the short form passes through).
  subnets = [
    for n in var.spec.network_configs : trimprefix(n.subnet, "https://www.googleapis.com/compute/v1/")
  ]

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending values it would reject or diff on.
  kms_key         = var.spec.kms_key != "" ? var.spec.kms_key : null
  rebalance_mode  = var.spec.rebalance_mode != "" ? var.spec.rebalance_mode : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The same planton-ai_* label set the Pulumi module applies, so the
  # cluster is attributable to its Planton object regardless of the engine
  # that created it. User labels merge in first so the platform attribution
  # labels can never be clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpmanagedkafkacluster"
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
