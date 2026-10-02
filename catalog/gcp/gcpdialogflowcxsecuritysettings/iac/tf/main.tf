# Enable the Dialogflow API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one set of security
# settings must never disable the API for every agent in the project.
resource "google_project_service" "dialogflow_api" {
  project = local.project_id
  service = "dialogflow.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The security settings. project and location are immutable; everything
# else updates in place.
resource "google_dialogflow_cx_security_settings" "this" {
  project      = local.project_id
  location     = var.spec.location
  display_name = local.display_name

  redaction_strategy    = local.redaction_strategy
  redaction_scope       = local.redaction_scope
  inspect_template      = local.inspect_template
  deidentify_template   = local.deidentify_template
  purge_data_types      = local.purge_data_types
  retention_strategy    = local.retention_strategy
  retention_window_days = local.retention_window_days

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  # Setting gcs_bucket makes Google grant the Dialogflow service agent
  # Storage Object Creator on the bucket.
  dynamic "audio_export_settings" {
    for_each = var.spec.audio_export_settings != null ? [var.spec.audio_export_settings] : []
    content {
      gcs_bucket             = audio_export_settings.value.gcs_bucket != "" ? audio_export_settings.value.gcs_bucket : null
      audio_export_pattern   = audio_export_settings.value.audio_export_pattern != "" ? audio_export_settings.value.audio_export_pattern : null
      audio_format           = audio_export_settings.value.audio_format != "" ? audio_export_settings.value.audio_format : null
      enable_audio_redaction = audio_export_settings.value.enable_audio_redaction ? true : null
    }
  }

  # The spec lifts the block's one required flag; the block is sent only
  # when export is on, so turning it off removes the block and Google's
  # export with it.
  dynamic "insights_export_settings" {
    for_each = var.spec.enable_insights_export ? [true] : []
    content {
      enable_insights_export = true
    }
  }

  depends_on = [google_project_service.dialogflow_api]
}
