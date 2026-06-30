##############################################
# outputs.tf
#
# Output values for KubernetesGhaRunnerScaleSet
##############################################

output "namespace" {
  description = "Namespace where the runner scale set is deployed"
  value       = local.namespace
}

output "release_name" {
  description = "Name of the Helm release"
  value       = local.release_name
}

output "chart_version" {
  description = "Version of the deployed Helm chart"
  value       = local.chart_version
}

output "runner_scale_set_name" {
  description = "Name of the runner scale set as registered with GitHub"
  value       = local.runner_scale_set_name
}

output "github_config_url" {
  description = "GitHub configuration URL"
  value       = var.spec.github.config_url
}

output "github_secret_name" {
  description = "Name of the Kubernetes secret containing GitHub credentials"
  value       = local.github_secret_name
}

output "pvc_names" {
  description = "Names of PVCs created for persistent volumes"
  value       = local.pvc_names
}

output "min_runners" {
  description = "Minimum runners configured"
  value       = local.min_runners
}

output "max_runners" {
  description = "Maximum runners configured"
  value       = local.max_runners
}

output "container_mode" {
  description = "Container mode type used"
  value       = local.container_mode_type
}

