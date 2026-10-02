locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The id and display name default to metadata.name -- the same fallback
  # the Pulumi module applies.
  stream_id    = var.spec.stream_id != "" ? var.spec.stream_id : var.metadata.name
  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  # desired_state is always sent, NOT_STARTED when unset, so both engines
  # declare the same state whatever each does with the provider's schema
  # default.
  desired_state = var.spec.desired_state != "" ? var.spec.desired_state : "NOT_STARTED"

  customer_managed_encryption_key = var.spec.customer_managed_encryption_key != "" ? var.spec.customer_managed_encryption_key : null
  deletion_policy                 = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  source      = var.spec.source_config
  destination = var.spec.destination_config
  bigquery    = local.destination.bigquery_destination_config
  gcs         = local.destination.gcs_destination_config

  # A GcpBigQueryDataset reference resolves to the dataset's self link;
  # Datastream wants projects/{project}/datasets/{dataset}, so the API
  # prefix is trimmed (a literal in either of Google's forms passes
  # through).
  single_target_dataset_id = (
    local.bigquery != null && try(local.bigquery.single_target_dataset, null) != null
  ) ? trimprefix(local.bigquery.single_target_dataset.dataset_id, "https://bigquery.googleapis.com/bigquery/v2/") : null

  # A GcpBigQueryConnection reference resolves to the connection's full
  # name; BigLake managed tables want {project}.{location}.{connection_id},
  # so a full name is converted (a literal in the dotted form passes
  # through).
  blmt_connection_name = (
    local.bigquery != null && try(local.bigquery.blmt_config, null) != null
    ) ? (
    can(regex("^projects/[^/]+/locations/[^/]+/connections/[^/]+$", local.bigquery.blmt_config.connection_name))
    ? join(".", regex("^projects/([^/]+)/locations/([^/]+)/connections/([^/]+)$", local.bigquery.blmt_config.connection_name))
    : local.bigquery.blmt_config.connection_name
  ) : null

  # The same planton-ai_* label set the Pulumi module applies. User labels
  # merge in first so the platform attribution labels can never be
  # clobbered by a spec label with the same key.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpdatastreamstream"
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
