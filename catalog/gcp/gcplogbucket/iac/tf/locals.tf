locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Which scope arm is set (at most one — spec-validated; all empty means a
  # project bucket in the provider's default project).
  is_folder_bucket  = var.spec.scope != null && var.spec.scope.folder_id != ""
  is_org_bucket     = var.spec.scope != null && var.spec.scope.organization_id != ""
  is_billing_bucket = var.spec.scope != null && var.spec.scope.billing_account != ""
  is_project_bucket = !local.is_folder_bucket && !local.is_org_bucket && !local.is_billing_bucket

  # The project the spec names, or null -- handed to the provider so an
  # import records it. The google_client_config fallback stays out of it:
  # a provider cannot depend on its own data source.
  project_id = local.is_project_bucket && var.spec.scope != null && var.spec.scope.project_id != "" ? var.spec.scope.project_id : null

  # The project bucket resource REQUIRES an explicit project. Honor the
  # spec contract for the empty case by reading the provider's resolved
  # default project from google_client_config (count-gated on that one
  # case so every plan that names its project runs credential-free).
  project = (
    local.is_project_bucket
    ? (
      var.spec.scope != null && var.spec.scope.project_id != ""
      ? var.spec.scope.project_id
      : data.google_client_config.current[0].project
    )
    : null
  )

  # The spec's defaults, sent explicitly (both engines apply the same ones
  # so behavior is identical regardless of engine): GCP's own retention
  # default is 30 days; "global" is the location unless data residency
  # demands a region.
  location       = var.spec.location != "" ? var.spec.location : "global"
  retention_days = var.spec.retention_days != 0 ? var.spec.retention_days : 30

  # The created bucket's full resource name — computed by whichever scope
  # variant exists; the views and the linked dataset attach to it, and it
  # is THE composition output.
  bucket_name = (
    local.is_project_bucket
    ? google_logging_project_bucket_config.this[0].name
    : local.is_folder_bucket
    ? google_logging_folder_bucket_config.this[0].name
    : local.is_org_bucket
    ? google_logging_organization_bucket_config.this[0].name
    : google_logging_billing_account_bucket_config.this[0].name
  )
}

# The provider's own resolved configuration — the source of the default
# project when the scope is empty. Count-gated on that one case.
data "google_client_config" "current" {
  count = local.is_project_bucket && (var.spec.scope == null || var.spec.scope.project_id == "") ? 1 : 0
}
