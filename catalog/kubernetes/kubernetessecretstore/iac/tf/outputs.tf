# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "store_name" {
  description = "Name of the created SecretStore (equals metadata.name) — the name ExternalSecrets in the same namespace reference (kind SecretStore)"
  value       = local.store_name
}

output "namespace" {
  description = "Namespace the SecretStore and its credential Secrets live in"
  value       = local.namespace
}
