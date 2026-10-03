# Outputs — identical names and derivations in the Pulumi module
# (KubernetesRabbitMqOutputs).

output "namespace" {
  description = "Namespace the cluster runs in"
  value       = local.namespace
}

output "cluster_name" {
  description = "Name of the RabbitmqCluster resource (equals metadata.name) — every operator-created object is derived from it"
  value       = local.cluster_name
}

output "service_name" {
  description = "Name of the client Service (operator naming contract: <name>)"
  value       = local.service_name
}

output "headless_service_name" {
  description = "Name of the headless inter-node Service (operator naming contract: <name>-nodes)"
  value       = local.headless_service_name
}

output "amqp_endpoint" {
  description = "In-cluster AMQP endpoint for clients on the effective port (5672, or 5671 when the plain listeners are closed)"
  value       = local.amqp_endpoint
}

output "management_endpoint" {
  description = "In-cluster management API / UI endpoint on the effective port (15672, or 15671 when the plain listeners are closed)"
  value       = local.management_endpoint
}

output "default_user_secret_name" {
  description = "Operator-generated Secret holding the administrator credentials (keys: username, password, host, port, connection_string, ...) — empty when the Vault secret backend owns the credentials"
  value       = local.default_user_secret_name
}

output "port_forward_command" {
  description = "Port-forward command for reaching the management UI from a workstation when no exposure is composed"
  value       = local.port_forward_command
}
