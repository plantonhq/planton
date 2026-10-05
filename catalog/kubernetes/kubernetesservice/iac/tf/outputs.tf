# Outputs — must flatten onto KubernetesServiceOutputs
# (outputs.proto) identically to the Pulumi module's exports. Every
# output is present (empty when not applicable) so both engines export the
# identical field set.

output "service_name" {
  description = "The name of the Service object as created in the cluster"
  value       = kubernetes_service_v1.service.metadata[0].name
}

output "namespace" {
  description = "The namespace the Service was created in"
  value       = kubernetes_service_v1.service.metadata[0].namespace
}

output "type" {
  description = "The service type as deployed"
  value       = local.service_type
}

output "cluster_ip" {
  description = "The cluster-internal virtual IP; empty for headless and ExternalName services"
  # Headless services carry the literal "None" in the API object and
  # ExternalName services carry nothing — both export empty, per the outputs
  # contract.
  value = (
    var.spec.headless || local.is_external_name
    ? ""
    : try(kubernetes_service_v1.service.spec[0].cluster_ip, "")
  )
}

output "load_balancer_ip" {
  description = "The IP of the provisioned load balancer; empty on hostname-based providers and non-LB types"
  value       = local.is_load_balancer ? try(kubernetes_service_v1.service.status[0].load_balancer[0].ingress[0].ip, "") : ""
}

output "load_balancer_hostname" {
  description = "The hostname of the provisioned load balancer; empty on IP-based providers and non-LB types"
  value       = local.is_load_balancer ? try(kubernetes_service_v1.service.status[0].load_balancer[0].ingress[0].hostname, "") : ""
}

output "kube_endpoint" {
  description = "In-cluster DNS endpoint of the Service"
  value       = local.kube_endpoint
}

output "port_forward_command" {
  description = "Ready-to-run port-forward command; empty for ExternalName services"
  value       = local.port_forward_command
}
