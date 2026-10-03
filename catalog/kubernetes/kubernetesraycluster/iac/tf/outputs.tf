# Outputs — identical names and derivations in the Pulumi module
# (KubernetesRayClusterOutputs).

output "namespace" {
  description = "Namespace the Ray cluster runs in"
  value       = local.namespace
}

output "head_service" {
  description = "The head Service (`<name>-head-svc`, the operator's naming contract) — every endpoint below rides it"
  value       = local.head_service
}

output "client_endpoint" {
  description = "In-cluster CLIENT endpoint (port 10001) — what ray.init(\"ray://…\") dials from notebooks and applications"
  value       = local.client_endpoint
}

output "dashboard_endpoint" {
  description = "In-cluster DASHBOARD/Job-API endpoint (port 8265) — the Job Submission API and the web dashboard; authenticated in token mode"
  value       = local.dashboard_endpoint
}

output "gcs_endpoint" {
  description = "In-cluster GCS endpoint (port 6379) — Ray's own coordination port; what `ray start --address` joins"
  value       = local.gcs_endpoint
}

output "auth_token_secret" {
  description = "The bearer-token Secret (key `auth_token`) in token auth mode — the credential handle for the dashboard/job/client APIs; unset when auth is disabled"
  value = local.token_auth_enabled ? {
    name = local.auth_token_secret_name
    key  = "auth_token"
  } : null
}

output "port_forward_command" {
  description = "Command to reach the dashboard from a workstation"
  value       = local.port_forward_command
}
