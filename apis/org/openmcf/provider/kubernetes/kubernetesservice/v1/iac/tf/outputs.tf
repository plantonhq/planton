# Terraform outputs for KubernetesService
# These map to KubernetesServiceStackOutputs fields.

output "service_name" {
  description = "The created service name"
  value       = kubernetes_service_v1.service.metadata[0].name
}

output "namespace" {
  description = "The namespace where the service was created"
  value       = kubernetes_service_v1.service.metadata[0].namespace
}

output "type" {
  description = "The service type (ClusterIP, NodePort, LoadBalancer, ExternalName)"
  value       = local.service_type
}

output "cluster_ip" {
  description = "The cluster-internal IP assigned to the service"
  value       = kubernetes_service_v1.service.spec[0].cluster_ip
}

output "load_balancer_ingress" {
  description = "The load balancer ingress hostname or IP (LoadBalancer type only)"
  value = try(
    kubernetes_service_v1.service.status[0].load_balancer[0].ingress[0].hostname,
    try(
      kubernetes_service_v1.service.status[0].load_balancer[0].ingress[0].ip,
      ""
    )
  )
}

output "internal_dns_name" {
  description = "Fully qualified internal DNS name for the service"
  value       = local.internal_dns_name
}
