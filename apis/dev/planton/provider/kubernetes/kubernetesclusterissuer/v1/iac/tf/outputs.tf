output "cluster_issuer_name" {
  description = "Name of the created ClusterIssuer (equals dns_domain)"
  value       = local.dns_domain
}

output "acme_account_key_secret_name" {
  description = "Name of the ACME account private key Secret"
  value       = local.acme_account_key_secret_name
}
