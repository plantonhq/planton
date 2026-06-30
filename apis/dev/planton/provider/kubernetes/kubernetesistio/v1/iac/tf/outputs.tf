output "namespace" {
  description = "The namespace where Istio control plane is deployed (istio-system)."
  value       = local.system_namespace
}

output "service" {
  description = "The name of the istiod service."
  value       = local.istiod_release_name
}

output "port_forward_command" {
  description = "Command to setup port-forwarding to access Istio control plane from local machine."
  value       = local.port_forward_command
}

output "kube_endpoint" {
  description = "Kubernetes endpoint to connect to istiod from within the cluster."
  value       = local.kube_endpoint
}

output "ingress_endpoint" {
  description = "Ingress endpoint for the Istio gateway."
  value       = local.ingress_endpoint
}

