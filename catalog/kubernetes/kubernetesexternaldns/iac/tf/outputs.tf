# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "namespace" {
  description = "Kubernetes namespace ExternalDNS was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.external_dns.name
}

output "service_account_name" {
  description = "Name of the ExternalDNS controller ServiceAccount — the identity to bind on the cloud side for keyless provider authentication"
  value       = local.service_account_name
}
