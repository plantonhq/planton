# Outputs — identical names and derivations in the Pulumi module
# (KubernetesGhaRunnerScaleSetOutputs).

output "namespace" {
  description = "Namespace the scale set (listener + runner pods) runs in"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (equals metadata.name)"
  value       = local.release_name
}

output "runner_scale_set_name" {
  description = "Name the fleet registered under in GitHub — the exact value workflows put in runs-on to target this fleet"
  value       = local.runner_scale_set_name
}

output "github_config_url" {
  description = "The GitHub URL the fleet serves (repository, organization or enterprise)"
  value       = var.spec.github_config_url
}
