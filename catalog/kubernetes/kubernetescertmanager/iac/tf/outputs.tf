# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.

output "namespace" {
  description = "Kubernetes namespace cert-manager was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.cert_manager.name
}

output "service_account_name" {
  description = "Name of the cert-manager controller ServiceAccount — the identity to bind on the cloud side for keyless DNS-01"
  value       = local.service_account_name
}

output "cluster_resource_namespace" {
  description = "Namespace cert-manager reads Secrets from for cluster-scoped resources (ClusterIssuer credentials live here)"
  value       = local.cluster_resource_namespace
}
