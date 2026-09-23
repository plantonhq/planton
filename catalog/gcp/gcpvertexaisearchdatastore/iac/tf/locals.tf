locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The store's GCP id and display name default to metadata.name (the
  # spec-level contract) -- identical to the Pulumi module.
  data_store_id = var.spec.data_store_id != "" ? var.spec.data_store_id : var.metadata.name
  display_name  = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending empty values it would reject or diff on.
  content_config  = var.spec.content_config != "" ? var.spec.content_config : null
  solution_types  = length(var.spec.solution_types) > 0 ? var.spec.solution_types : null
  kms_key_name    = var.spec.kms_key_name != "" ? var.spec.kms_key_name : null
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # The folded companions. Target sites are keyed by their URI pattern so
  # adding or removing one never renumbers the others; sitemaps by URI.
  target_sites = { for site in var.spec.target_sites : site.provided_uri_pattern => site }
  sitemaps     = toset(var.spec.sitemap_uris)
}
