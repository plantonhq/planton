locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The template id and display name default to metadata.name -- identical
  # to the Pulumi module.
  runtime_template_id = var.spec.runtime_template_id != "" ? var.spec.runtime_template_id : var.metadata.name
  display_name        = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  description     = var.spec.description != "" ? var.spec.description : null
  kms_key_name    = var.spec.kms_key_name != "" ? var.spec.kms_key_name : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
  network_tags    = length(var.spec.network_tags) > 0 ? var.spec.network_tags : null

  # A GcpSubnetwork reference arrives as a compute self-link; Colab takes the
  # relative path. Trimming the prefix leaves a literal relative path as is
  # -- the same rule as the Pulumi module.
  subnetwork = (
    var.spec.network_spec != null && var.spec.network_spec.subnetwork != ""
    ? trimprefix(var.spec.network_spec.subnetwork, "https://www.googleapis.com/compute/v1/")
    : null
  )

  # The same planton-ai_* label set the Pulumi module applies. User labels
  # merge in first so the platform attribution labels can never be
  # clobbered by a spec label with the same key. Google fixes a template's
  # labels at creation: any change here replaces the template.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpcolabruntimetemplate"
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
