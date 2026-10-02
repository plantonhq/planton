# Enable the Workflows API so a fresh project can host the workflow.
# disable_on_destroy is false: tearing down one workflow must never disable
# Workflows for everything else in the project.
resource "google_project_service" "workflows_api" {
  project = local.project_id
  service = "workflows.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A Cloud Workflows workflow — a serverless orchestrator executing the
# YAML/JSON source on each run. Every source / env-var / service-account
# change deploys a NEW revision; running executions finish on the revision
# they started with.
#
# `deletion_protection` is sent EXPLICITLY on every apply: it is Optional
# in the provider with default true, and a spec transition true -> false
# must reach the engine rather than being omitted (the send-true-or-omit
# class silently no-ops such transitions — a destroy that should have been
# allowed keeps failing).
resource "google_workflows_workflow" "this" {
  name    = local.workflow_name
  project = local.project_id
  region  = local.region

  # REQUIRED by the spec (API truth the provider defers to its 8.0.0 line).
  source_contents = var.spec.source_contents

  description = var.spec.description != "" ? var.spec.description : null
  labels      = local.final_labels

  service_account = var.spec.service_account != "" ? var.spec.service_account : null
  crypto_key_name = var.spec.crypto_key != "" ? var.spec.crypto_key : null

  call_log_level          = var.spec.call_log_level != "" ? var.spec.call_log_level : null
  execution_history_level = var.spec.execution_history_level != "" ? var.spec.execution_history_level : null

  # The literals plus each stored secret's version resource name (see
  # secrets.tf); never a secret value.
  user_env_vars = length(local.user_env_vars) > 0 ? local.user_env_vars : null

  # ForceNew: a tag change REPLACES the workflow (fresh execution history).
  tags = length(var.spec.resource_manager_tags) > 0 ? var.spec.resource_manager_tags : null

  # Explicit send — see the resource comment.
  deletion_protection = var.spec.deletion_protection

  # Empty defers to the provider default (DELETE).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  # Deployed after the grants so the first execution can read its secrets.
  depends_on = [google_project_service.workflows_api, google_secret_manager_secret_iam_member.env]
}
