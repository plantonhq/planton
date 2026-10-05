# Outputs — identical names and derivations in the Pulumi module
# (KubernetesKeycloakOperatorOutputs).

output "namespace" {
  description = "Namespace the operator runs in (where namespaced-watch Keycloak declarations must also live)"
  value       = local.namespace
}

output "deployment" {
  description = "The operator Deployment name (upstream-fixed: keycloak-operator)"
  value       = local.deployment_name
}

output "service" {
  description = "The operator's metrics/health Service name (upstream-fixed: keycloak-operator, port 80 -> 8080)"
  value       = local.service_name
}
