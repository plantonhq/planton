# Enable the Cloud Logging API so a fresh project can host the sink.
# Rendered only on the project-scope path: folder/org/billing sinks are not
# project resources, so there is no project to enable the API in.
# disable_on_destroy is false: tearing down one sink must never disable
# logging for everything else in the project.
resource "google_project_service" "logging_api" {
  count = local.is_project_sink ? 1 : 0

  project = local.project_id
  service = "logging.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A Cloud Logging sink — the routing rule that exports matching log entries
# to the rendered destination. One kind, four provider resources, exactly
# one created (the count guards key off the scope locals).
#
# THE post-create step every sink needs: grant the writer_identity output
# write access on the destination (roles/storage.objectCreator on a bucket,
# roles/bigquery.dataEditor on a dataset, roles/pubsub.publisher on a
# topic) — through a standalone grant (GcpGcsBucketIamMember,
# GcpPubSubTopicIamMember) whose member references writer_identity, never
# the destination's own IAM fields, which would form a cycle — or the sink
# silently exports nothing.
#
# unique_writer_identity is sent EXPLICITLY on the project sink: it is
# Optional in the provider with default true, and a spec transition
# true -> false must reach the API rather than being omitted (the
# send-true-or-omit class). The other scopes ALWAYS mint a unique writer —
# their resources carry no writer-identity arguments at all, which is why
# the spec's validations pin those fields to the project scope.
resource "google_logging_project_sink" "this" {
  count = local.is_project_sink ? 1 : 0

  name        = local.sink_name
  project     = local.project_id
  destination = local.destination

  filter      = var.spec.filter != "" ? var.spec.filter : null
  description = var.spec.description != "" ? var.spec.description : null
  disabled    = var.spec.disabled ? true : null

  unique_writer_identity = var.spec.unique_writer_identity == null ? true : var.spec.unique_writer_identity
  custom_writer_identity = var.spec.custom_writer_identity != "" ? var.spec.custom_writer_identity : null

  dynamic "bigquery_options" {
    for_each = local.render_bigquery_options ? [1] : []
    content {
      use_partitioned_tables = true
    }
  }

  dynamic "exclusions" {
    for_each = var.spec.exclusions
    content {
      name        = exclusions.value.name
      filter      = exclusions.value.filter
      description = exclusions.value.description != "" ? exclusions.value.description : null
      disabled    = exclusions.value.disabled
    }
  }

  # Empty defers to the provider default (DELETE).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  depends_on = [google_project_service.logging_api]
}

# The folder-scoped variant — adds the children-routing capability
# (include_children / intercept_children) that only folder and organization
# sinks carry.
resource "google_logging_folder_sink" "this" {
  count = local.is_folder_sink ? 1 : 0

  name        = local.sink_name
  folder      = var.spec.scope.folder_id
  destination = local.destination

  filter      = var.spec.filter != "" ? var.spec.filter : null
  description = var.spec.description != "" ? var.spec.description : null
  disabled    = var.spec.disabled ? true : null

  include_children   = var.spec.include_children
  intercept_children = var.spec.intercept_children

  dynamic "bigquery_options" {
    for_each = local.render_bigquery_options ? [1] : []
    content {
      use_partitioned_tables = true
    }
  }

  dynamic "exclusions" {
    for_each = var.spec.exclusions
    content {
      name        = exclusions.value.name
      filter      = exclusions.value.filter
      description = exclusions.value.description != "" ? exclusions.value.description : null
      disabled    = exclusions.value.disabled
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# The organization-scoped variant.
resource "google_logging_organization_sink" "this" {
  count = local.is_org_sink ? 1 : 0

  name        = local.sink_name
  org_id      = var.spec.scope.organization_id
  destination = local.destination

  filter      = var.spec.filter != "" ? var.spec.filter : null
  description = var.spec.description != "" ? var.spec.description : null
  disabled    = var.spec.disabled ? true : null

  include_children   = var.spec.include_children
  intercept_children = var.spec.intercept_children

  dynamic "bigquery_options" {
    for_each = local.render_bigquery_options ? [1] : []
    content {
      use_partitioned_tables = true
    }
  }

  dynamic "exclusions" {
    for_each = var.spec.exclusions
    content {
      name        = exclusions.value.name
      filter      = exclusions.value.filter
      description = exclusions.value.description != "" ? exclusions.value.description : null
      disabled    = exclusions.value.disabled
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}

# The billing-account-scoped variant — the leanest surface: no writer
# identity controls, no children routing.
resource "google_logging_billing_account_sink" "this" {
  count = local.is_billing_sink ? 1 : 0

  name            = local.sink_name
  billing_account = var.spec.scope.billing_account
  destination     = local.destination

  filter      = var.spec.filter != "" ? var.spec.filter : null
  description = var.spec.description != "" ? var.spec.description : null
  disabled    = var.spec.disabled ? true : null

  dynamic "bigquery_options" {
    for_each = local.render_bigquery_options ? [1] : []
    content {
      use_partitioned_tables = true
    }
  }

  dynamic "exclusions" {
    for_each = var.spec.exclusions
    content {
      name        = exclusions.value.name
      filter      = exclusions.value.filter
      description = exclusions.value.description != "" ? exclusions.value.description : null
      disabled    = exclusions.value.disabled
    }
  }

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
