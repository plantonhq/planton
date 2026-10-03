# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "external_secret_name" {
  description = "Name of the created ExternalSecret (equals metadata.name)"
  value       = local.external_secret_name
}

output "namespace" {
  description = "Namespace the ExternalSecret (and its materialized Secret) lives in"
  value       = local.namespace
}

output "secret_name" {
  description = "Name of the Kubernetes Secret the operator materializes (target.name when set, else metadata.name) — the handle workloads reference"
  value       = local.secret_name
}
