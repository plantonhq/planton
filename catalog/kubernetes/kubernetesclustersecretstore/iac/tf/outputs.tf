# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "store_name" {
  description = "Name of the created ClusterSecretStore (equals metadata.name) — the name ExternalSecrets reference with kind ClusterSecretStore"
  value       = local.store_name
}

output "secrets_namespace" {
  description = "Namespace credential Secrets for this store were materialized in (conventionally the operator's install namespace)"
  value       = local.secrets_namespace
}
