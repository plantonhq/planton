# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "cluster_issuer_name" {
  description = "Name of the created ClusterIssuer (equals metadata.name)"
  value       = local.cluster_issuer_name
}

output "secrets_namespace" {
  description = "Namespace credential Secrets for this issuer were materialized in (cert-manager's cluster-resource namespace)"
  value       = local.secrets_namespace
}

output "acme_account_key_secret_name" {
  description = "Name of the ACME account private key Secret cert-manager creates (empty for non-ACME backends)"
  value       = local.acme_account_key_secret_name
}
