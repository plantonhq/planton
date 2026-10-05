# Outputs — flattened onto KubernetesIngressNginxOutputs by the
# platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the controller was installed into"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (= metadata.name; controller resources are named \"<release>-controller\")"
  value       = local.release_name
}

output "ingress_class_name" {
  description = "Name of the IngressClass this controller owns — what KubernetesIngress resources reference to route through it"
  value       = local.ingress_class_name
}

output "controller_service_name" {
  description = "Name of the controller's external Service (the traffic entry point)"
  value       = local.controller_service_name
}

output "internal_service_name" {
  description = "Name of the controller's internal Service — empty unless spec.service.internal.enabled"
  value       = local.internal_service_name
}

output "load_balancer_ip" {
  description = "External IP of the controller's LoadBalancer (providers that populate an IP; empty otherwise)"
  value       = local.is_load_balancer ? try(data.kubernetes_service_v1.controller[0].status[0].load_balancer[0].ingress[0].ip, "") : ""
}

output "load_balancer_hostname" {
  description = "External hostname of the controller's LoadBalancer (providers that populate a DNS name; empty otherwise)"
  value       = local.is_load_balancer ? try(data.kubernetes_service_v1.controller[0].status[0].load_balancer[0].ingress[0].hostname, "") : ""
}
