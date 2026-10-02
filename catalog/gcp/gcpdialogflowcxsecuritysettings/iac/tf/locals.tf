locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Google requires a display name; the spec defaults it to metadata.name --
  # identical to the Pulumi module.
  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  # Empty optional strings become null so the provider omits them from the
  # API payload instead of sending values it would reject or diff on.
  redaction_strategy  = var.spec.redaction_strategy != "" ? var.spec.redaction_strategy : null
  redaction_scope     = var.spec.redaction_scope != "" ? var.spec.redaction_scope : null
  inspect_template    = var.spec.inspect_template != "" ? var.spec.inspect_template : null
  deidentify_template = var.spec.deidentify_template != "" ? var.spec.deidentify_template : null
  retention_strategy  = var.spec.retention_strategy != "" ? var.spec.retention_strategy : null
  deletion_policy     = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # 0 means Google's default TTL, which is also what an unset window means;
  # send the window only when one is declared.
  retention_window_days = var.spec.retention_window_days > 0 ? var.spec.retention_window_days : null
  purge_data_types      = length(var.spec.purge_data_types) > 0 ? var.spec.purge_data_types : null
}
