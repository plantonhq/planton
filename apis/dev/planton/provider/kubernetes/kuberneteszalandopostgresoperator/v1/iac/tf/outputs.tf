output "namespace" {
  description = "Kubernetes namespace in which the Zalando Postgres Operator is created"
  value       = local.namespace
}

output "service" {
  description = "Kubernetes service name for the Zalando Postgres Operator"
  value       = local.service_name
}

output "port_forward_command" {
  description = "Command to setup port-forwarding to access the operator from developers laptop"
  value       = local.port_forward_command
}

output "kube_endpoint" {
  description = "Kubernetes endpoint to connect to the operator from within the cluster"
  value       = local.kube_endpoint
}

output "ingress_endpoint" {
  description = "Public endpoint to open the operator from clients outside kubernetes (not applicable for this operator)"
  value       = ""
}

