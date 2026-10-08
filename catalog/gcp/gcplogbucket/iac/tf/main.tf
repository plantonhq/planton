# Enable the Cloud Logging API so a fresh project can host the bucket.
# Rendered only on the project-scope path: folder/org/billing buckets are
# not project resources, so there is no project to enable the API in.
# disable_on_destroy is false: tearing down one bucket must never disable
# logging for everything else in the project.
resource "google_project_service" "logging_api" {
  count = local.is_project_bucket ? 1 : 0

  project = local.project
  service = "logging.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A Cloud Logging bucket — one kind, four scope resources, exactly one
# created (the count guards key off the scope locals).
#
# ADOPTION SEMANTICS (provider truth, all scopes): a bucket_id matching an
# EXISTING bucket adopts and patches it rather than failing — that is how
# the built-in _Default bucket is managed, and the only mode at all on
# folder/org/billing scopes (the Logging API creates NEW custom buckets
# only under projects). _Default/_Required are undeletable: a destroy
# removes them from management and leaves them in GCP.
#
# retention_days and location carry the spec's defaults (30 / "global")
# and are always sent, so the spec default is what the API applies rather
# than a silently different server-side state.
#
# enable_analytics is sent ONLY when the spec explicitly sets it: the
# provider transmits analyticsEnabled solely on explicit configuration (an
# atomic pre-update, separate from other fields), and enabling is ONE-WAY.
# Blanket-sending false would diverge from provider behavior.
resource "google_logging_project_bucket_config" "this" {
  count = local.is_project_bucket ? 1 : 0

  project   = local.project
  bucket_id = var.spec.bucket_id
  location  = local.location

  retention_days = local.retention_days
  description    = var.spec.description != "" ? var.spec.description : null

  # locked matches the provider default (false) when unset; a live unlock
  # is refused server-side (one-way), which is the honest failure for a
  # true -> false transition.
  locked = var.spec.locked

  enable_analytics = var.spec.enable_analytics

  dynamic "cmek_settings" {
    for_each = var.spec.cmek_kms_key != "" ? [1] : []
    content {
      kms_key_name = var.spec.cmek_kms_key
    }
  }

  dynamic "index_configs" {
    for_each = var.spec.index_configs
    content {
      field_path = index_configs.value.field_path
      type       = index_configs.value.type
    }
  }

  # Empty defers to the provider default (DELETE).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  depends_on = [google_project_service.logging_api]
}

# The folder-scoped variant — ADOPT-only (see above); no locked /
# enable_analytics (project-scope-only arguments by provider truth).
resource "google_logging_folder_bucket_config" "this" {
  count = local.is_folder_bucket ? 1 : 0

  folder    = var.spec.scope.folder_id
  bucket_id = var.spec.bucket_id
  location  = local.location

  retention_days = local.retention_days
  description    = var.spec.description != "" ? var.spec.description : null

  dynamic "cmek_settings" {
    for_each = var.spec.cmek_kms_key != "" ? [1] : []
    content {
      kms_key_name = var.spec.cmek_kms_key
    }
  }

  dynamic "index_configs" {
    for_each = var.spec.index_configs
    content {
      field_path = index_configs.value.field_path
      type       = index_configs.value.type
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# The organization-scoped variant.
resource "google_logging_organization_bucket_config" "this" {
  count = local.is_org_bucket ? 1 : 0

  organization = var.spec.scope.organization_id
  bucket_id    = var.spec.bucket_id
  location     = local.location

  retention_days = local.retention_days
  description    = var.spec.description != "" ? var.spec.description : null

  dynamic "cmek_settings" {
    for_each = var.spec.cmek_kms_key != "" ? [1] : []
    content {
      kms_key_name = var.spec.cmek_kms_key
    }
  }

  dynamic "index_configs" {
    for_each = var.spec.index_configs
    content {
      field_path = index_configs.value.field_path
      type       = index_configs.value.type
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# The billing-account-scoped variant.
resource "google_logging_billing_account_bucket_config" "this" {
  count = local.is_billing_bucket ? 1 : 0

  billing_account = var.spec.scope.billing_account
  bucket_id       = var.spec.bucket_id
  location        = local.location

  retention_days = local.retention_days
  description    = var.spec.description != "" ? var.spec.description : null

  dynamic "cmek_settings" {
    for_each = var.spec.cmek_kms_key != "" ? [1] : []
    content {
      kms_key_name = var.spec.cmek_kms_key
    }
  }

  dynamic "index_configs" {
    for_each = var.spec.index_configs
    content {
      field_path = index_configs.value.field_path
      type       = index_configs.value.type
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# Log views: named, independently grantable slices of the bucket
# (roles/logging.viewAccessor on a view shows a reader only the entries
# matching its filter). The view's location and parent derive from the
# bucket's full name; the kind's deletion_policy fans out to every view.
resource "google_logging_log_view" "this" {
  for_each = { for log_view in var.spec.log_views : log_view.view_id => log_view }

  name   = each.value.view_id
  bucket = local.bucket_name

  filter      = each.value.filter != "" ? each.value.filter : null
  description = each.value.description != "" ? each.value.description : null

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# The linked BigQuery dataset (requires analytics — spec-validated
# pairing). Every field of a linked dataset is create-time-only.
resource "google_logging_linked_dataset" "this" {
  count = var.spec.linked_bigquery_dataset != null ? 1 : 0

  link_id = var.spec.linked_bigquery_dataset.link_id
  bucket  = local.bucket_name

  description = var.spec.linked_bigquery_dataset.description != "" ? var.spec.linked_bigquery_dataset.description : null

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# The folder/organization logging-settings singleton (scope-gated by the
# spec's validations). Adopted on create, patched in place; destroy is a
# state-only no-op — the settings resources carry no deletion_policy by
# provider truth (the API has no delete).
resource "google_logging_folder_settings" "this" {
  count = local.is_folder_bucket && var.spec.scope_settings != null ? 1 : 0

  folder = var.spec.scope.folder_id

  # The block's reason to exist — sent explicitly.
  disable_default_sink = var.spec.scope_settings.disable_default_sink

  kms_key_name     = var.spec.scope_settings.kms_key != "" ? var.spec.scope_settings.kms_key : null
  storage_location = var.spec.scope_settings.storage_location != "" ? var.spec.scope_settings.storage_location : null
}

resource "google_logging_organization_settings" "this" {
  count = local.is_org_bucket && var.spec.scope_settings != null ? 1 : 0

  organization = var.spec.scope.organization_id

  disable_default_sink = var.spec.scope_settings.disable_default_sink

  kms_key_name     = var.spec.scope_settings.kms_key != "" ? var.spec.scope_settings.kms_key : null
  storage_location = var.spec.scope_settings.storage_location != "" ? var.spec.scope_settings.storage_location : null
}
