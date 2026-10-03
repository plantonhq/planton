# Outputs — identical names and derivations in the Pulumi module
# (KubernetesKeycloakOutputs).

output "namespace" {
  description = "Namespace the Keycloak server runs in"
  value       = local.namespace
}

output "stateful_set" {
  description = "The operator-created StatefulSet (named exactly after this resource)"
  value       = local.stateful_set
}

output "service" {
  description = "The main Service (`<name>-service`) — the backend handle exposure kinds reference"
  value       = local.service_name
}

output "discovery_service" {
  description = "The headless discovery Service (`<name>-discovery`) — JGroups cluster formation between instances"
  value       = local.discovery_service
}

output "api_endpoint" {
  description = "In-cluster API endpoint, scheme included — https on the https port when TLS is configured, plain http otherwise"
  value       = local.api_endpoint
}

output "management_endpoint" {
  description = "The management endpoint (health probes and metrics) on the management port"
  value       = local.management_endpoint
}

output "initial_admin_secret_name" {
  description = "The bootstrap-admin credential Secret: user-provided when declared, else the operator-generated create-once `<name>-initial-admin` (username temp-admin) — break-glass material"
  value       = local.initial_admin_secret_name
}

output "port_forward_command" {
  description = "Command to reach the server from a workstation"
  value       = local.port_forward_command
}
