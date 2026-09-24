locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The Connect cluster's GCP id defaults to metadata.name -- the same
  # fallback the Pulumi module applies.
  connect_cluster_id = var.spec.connect_cluster_id != "" ? var.spec.connect_cluster_id : var.metadata.name

  # A GcpSubnetwork reference resolves to the subnet's self link; Google
  # wants the projects/{p}/regions/{r}/subnetworks/{s} form, so the API
  # prefix is trimmed (a literal in the short form passes through).
  network_configs = [
    for n in var.spec.network_configs : {
      primary_subnet   = trimprefix(n.primary_subnet, "https://www.googleapis.com/compute/v1/")
      dns_domain_names = length(n.dns_domain_names) > 0 ? n.dns_domain_names : null
    }
  ]

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The same planton-ai_* label set the Pulumi module applies. User labels
  # merge in first so the platform attribution labels can never be
  # clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpmanagedkafkaconnectcluster"
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
