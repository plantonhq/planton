# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "namespace" {
  description = "Kubernetes namespace the External Secrets Operator was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.external_secrets.name
}

output "controller_service_account" {
  description = "Name of the controller ServiceAccount — the identity to bind on the cloud side for ambient (keyless) store access"
  value       = local.controller_service_account
}
