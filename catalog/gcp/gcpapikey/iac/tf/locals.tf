locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain; an empty string would be sent
  # verbatim and rejected by the API.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The restriction arms the spec declares. Each is a one-element list when
  # present and empty when absent, so the dynamic blocks below emit the
  # provider's block exactly when the spec has one -- an omitted arm means
  # "no restriction of that class" to the API, and a hollow block would be
  # rejected (each arm's list is required).
  restrictions = var.spec.restrictions != null ? [var.spec.restrictions] : []
}
