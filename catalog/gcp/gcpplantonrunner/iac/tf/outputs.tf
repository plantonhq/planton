# Outputs — identical names and derivations in the Pulumi module
# (GcpPlantonRunnerOutputs).

output "service_name" {
  description = "The fully qualified Cloud Run service name (projects/{project}/locations/{region}/services/{name})."
  value       = google_cloud_run_v2_service.runner.id
}

output "service_short_name" {
  description = "The service's short name (metadata.name)."
  value       = google_cloud_run_v2_service.runner.name
}

output "service_account_email" {
  description = "The runner's runtime identity -- grant it roles for keyless cloud access."
  value       = local.service_account_email
}

output "token_secret_id" {
  description = "The Secret Manager secret holding the runner token -- the token authorizes joining and is never the runner's identity."
  value       = "projects/${google_cloud_run_v2_service.runner.project}/secrets/${local.token_secret_id}"
}

output "runner_name" {
  description = "The name the runner registers itself under with the control plane -- shown by `planton runner list` the moment it joins."
  value       = local.registration_name
}

output "project_id" {
  description = "The GCP project the runner was deployed in."
  value       = google_cloud_run_v2_service.runner.project
}

output "region" {
  description = "The GCP region the runner was deployed in."
  value       = var.spec.region
}
