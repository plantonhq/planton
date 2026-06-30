# Outputs for Argo CD deployment
# These outputs match the KubernetesArgocdStackOutputs protobuf message

output "namespace" {
  description = "Kubernetes namespace in which Argo CD is created"
  value       = local.namespace_name
}

output "service" {
  description = "Kubernetes service name for Argo CD server"
  value       = local.service_name
}

output "port_forward_command" {
  description = "Command to setup port-forwarding to open Argo CD from developer's laptop"
  value       = local.port_forward_command
}

output "kube_endpoint" {
  description = "Kubernetes endpoint to connect to Argo CD from within the cluster"
  value       = local.kube_service_fqdn
}

output "external_hostname" {
  description = "Public endpoint to open Argo CD from clients outside Kubernetes"
  value       = local.ingress_external_hostname
}

output "internal_hostname" {
  description = "Internal endpoint to open Argo CD from clients within the network"
  value       = local.ingress_internal_hostname
}

