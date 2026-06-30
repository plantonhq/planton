###########################
# outputs.tf
###########################
# Output shape matches the KubernetesExternalDnsStackOutputs proto and the Pulumi
# module's exports (namespace, release_name, solver_sa) so both engines populate the
# same status.outputs fields. Do not emit flat/extra names that never reach the proto.

output "namespace" {
  description = "Namespace where ExternalDNS is deployed"
  value       = local.namespace_name
}

output "release_name" {
  description = "Helm release name for ExternalDNS"
  value       = helm_release.external_dns.name
}

output "solver_sa" {
  description = "Kubernetes service account name for ExternalDNS"
  value       = kubernetes_service_account_v1.external_dns.metadata[0].name
}
