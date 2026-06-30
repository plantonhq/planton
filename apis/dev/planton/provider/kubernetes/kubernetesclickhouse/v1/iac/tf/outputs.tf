output "namespace" {
  description = "The namespace in which ClickHouse is deployed."
  value       = local.namespace
}

output "service" {
  description = "The base name of the ClickHouse service created by the operator."
  value       = local.kube_service_name
}

output "kube_endpoint" {
  description = "Internal DNS name of the ClickHouse service within the cluster."
  value       = "${local.kube_service_name}.${local.namespace}.svc.cluster.local:8123"
}

output "port_forward_command" {
  description = "Command to port-forward traffic to the ClickHouse service on localhost:8123."
  value       = "kubectl port-forward -n ${local.namespace} service/${local.kube_service_name} 8123:8123"
}

output "external_hostname" {
  description = "The external hostname for ClickHouse if ingress is enabled."
  value       = local.ingress_external_hostname
}

output "username" {
  description = "The default ClickHouse username."
  value       = local.default_username
}

output "password_secret_name" {
  description = "Name of the Secret holding the ClickHouse password."
  value       = local.password_secret_name
}

output "password_secret_key" {
  description = "Key within the Secret that contains the ClickHouse password."
  value       = local.password_secret_key
}
