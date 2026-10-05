# Outputs — identical names and derivations in the Pulumi module
# (KubernetesPlantonRunnerOutputs).

output "namespace" {
  description = "The namespace the runner is installed in."
  value       = local.namespace
}

output "release_name" {
  description = "The Helm release name (metadata.name)."
  value       = local.release_name
}

output "token_secret_name" {
  description = "The Kubernetes Secret holding the runner token -- the token authorizes joining and is never the runner's identity."
  value       = local.token_secret_name
}

output "runner_name" {
  description = "The name the runner registers itself under with the control plane -- shown by `planton runner list` the moment it joins."
  value       = local.runner_name
}
