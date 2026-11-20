# Full resource name (projects/{project}/locations/{region}/workerPools/{name}).
output "name" {
  description = "Full resource name of the worker pool"
  value       = google_cloud_run_v2_worker_pool.main.id
}

output "worker_pool_name" {
  description = "Bare worker pool name in GCP"
  value       = google_cloud_run_v2_worker_pool.main.name
}

output "uid" {
  description = "Server-generated unique identifier for the worker pool"
  value       = google_cloud_run_v2_worker_pool.main.uid
}

output "location" {
  description = "Region the worker pool runs in"
  value       = google_cloud_run_v2_worker_pool.main.location
}

output "project_id" {
  description = "GCP project the worker pool lives in"
  value       = google_cloud_run_v2_worker_pool.main.project
}

output "latest_created_revision" {
  description = "Name of the most recently created revision"
  value       = google_cloud_run_v2_worker_pool.main.latest_created_revision
}

output "latest_ready_revision" {
  description = "Name of the newest revision serving instances"
  value       = google_cloud_run_v2_worker_pool.main.latest_ready_revision
}

output "observed_generation" {
  description = "Generation the controller has reconciled"
  value       = google_cloud_run_v2_worker_pool.main.observed_generation
}

output "etag" {
  description = "Opaque version token for optimistic concurrency"
  value       = google_cloud_run_v2_worker_pool.main.etag
}

# The pool's overall readiness condition, flattened from the provider's
# one-element list; empty strings until Google reports one.
output "terminal_condition" {
  description = "The pool's overall readiness condition (type Ready), with the failure detail when a rollout fails"
  value = {
    type                 = try(google_cloud_run_v2_worker_pool.main.terminal_condition[0].type, "")
    state                = try(google_cloud_run_v2_worker_pool.main.terminal_condition[0].state, "")
    message              = try(google_cloud_run_v2_worker_pool.main.terminal_condition[0].message, "")
    last_transition_time = try(google_cloud_run_v2_worker_pool.main.terminal_condition[0].last_transition_time, "")
    severity             = try(google_cloud_run_v2_worker_pool.main.terminal_condition[0].severity, "")
    reason               = try(google_cloud_run_v2_worker_pool.main.terminal_condition[0].reason, "")
    revision_reason      = try(google_cloud_run_v2_worker_pool.main.terminal_condition[0].revision_reason, "")
    execution_reason     = try(google_cloud_run_v2_worker_pool.main.terminal_condition[0].execution_reason, "")
  }
}

# The sub-resources' conditions, in Google's order: the diagnostics to read
# when the terminal condition has not succeeded.
output "conditions" {
  description = "Conditions of the pool's sub-resources, the extra diagnostics behind a failed rollout"
  value = [
    for condition in try(google_cloud_run_v2_worker_pool.main.conditions, []) : {
      type                 = condition.type
      state                = condition.state
      message              = condition.message
      last_transition_time = condition.last_transition_time
      severity             = condition.severity
      reason               = condition.reason
      revision_reason      = condition.revision_reason
      execution_reason     = condition.execution_reason
    }
  ]
}
