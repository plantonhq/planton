locals {
  # Honor the spec contract: an empty project_id falls back to the
  # provider's default project. Passing null (instead of "") lets the
  # google provider resolve its own project from configuration.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Fleet features live in "global" unless the manifest says otherwise --
  # identical to the Pulumi module.
  location        = var.spec.location != "" ? var.spec.location : "global"
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The API each feature needs beside the Fleet API, from Google's
  # per-feature setup guides; features absent here need only the Fleet API.
  # The Pulumi module carries the same table.
  feature_apis = {
    configmanagement             = "anthosconfigmanagement.googleapis.com"
    policycontroller             = "anthospolicycontroller.googleapis.com"
    servicemesh                  = "mesh.googleapis.com"
    multiclusteringress          = "multiclusteringress.googleapis.com"
    multiclusterservicediscovery = "multiclusterservicediscovery.googleapis.com"
  }
  feature_api = lookup(local.feature_apis, var.spec.feature, null)

  # The provider's spec block is sent when any feature settings block is
  # declared; the spec allows only the block of this feature.
  has_feature_spec = anytrue([
    var.spec.multiclusteringress != null,
    var.spec.fleetobservability != null,
    var.spec.clusterupgrade != null,
    var.spec.rbacrolebindingactuation != null,
    var.spec.workloadidentity != null,
  ])

  default_member_config = var.spec.fleet_default_member_config

  # The same planton-ai_* label set the Pulumi module applies. User labels
  # merge in first so the platform attribution labels can never be
  # clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpgkefleetfeature"
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

  # Per-cluster settings keyed by the membership's full name
  # ("projects/{p}/locations/{l}/memberships/{id}"), from which the
  # membership's ID and location are parsed. The entry lives in this
  # feature's project, which Google requires the membership to share.
  membership_configs = {
    for config in var.spec.membership_configs : config.membership => merge(config, {
      membership_location = split("/", config.membership)[3]
      membership_id       = split("/", config.membership)[5]
    })
  }
}
