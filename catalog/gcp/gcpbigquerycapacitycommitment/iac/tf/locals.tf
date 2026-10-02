locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The commitment's id defaults to metadata.name -- the same fallback the
  # Pulumi module applies -- instead of an id Google generates.
  capacity_commitment_id = var.spec.capacity_commitment_id != "" ? var.spec.capacity_commitment_id : var.metadata.name

  # Empty optionals become null so the provider omits them and Google's
  # defaults apply.
  location        = var.spec.location != "" ? var.spec.location : null
  renewal_plan    = var.spec.renewal_plan != "" ? var.spec.renewal_plan : null
  edition         = var.spec.edition != "" ? var.spec.edition : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The provider types this boolean as a string; the spec's bool is sent as
  # "true" when set and omitted otherwise -- identical to the Pulumi module.
  enforce_single_admin_project_per_org = var.spec.enforce_single_admin_project_per_org ? "true" : null
}
