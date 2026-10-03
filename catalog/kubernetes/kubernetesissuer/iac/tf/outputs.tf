# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "namespace" {
  description = "Namespace the Issuer was created in"
  value       = local.namespace
}

output "issuer_name" {
  description = "Name of the created Issuer (equals metadata.name)"
  value       = local.issuer_name
}

output "acme_account_key_secret_name" {
  description = "Name of the ACME account private key Secret cert-manager creates (empty for non-ACME backends)"
  value       = local.acme_account_key_secret_name
}
