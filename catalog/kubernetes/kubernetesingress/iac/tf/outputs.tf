# Outputs — must flatten onto KubernetesIngressOutputs
# (outputs.proto) identically to the Pulumi module's exports.
#
# The load-balancer handles read the object's status WITHOUT waiting for a
# controller (wait_for_load_balancer = false — see main.tf): on a cluster
# where a controller reconciles quickly the values land on the same deploy;
# on a cluster with no controller they export empty, matching the object's
# real state.

output "ingress_name" {
  description = "The name of the Ingress object as created in the cluster"
  value       = kubernetes_ingress_v1.ingress.metadata[0].name
}

output "namespace" {
  description = "The namespace the Ingress was created in"
  value       = kubernetes_ingress_v1.ingress.metadata[0].namespace
}

output "load_balancer_ip" {
  description = "The IP the controller's load balancer exposes; empty until a controller reconciles the Ingress"
  value       = try(kubernetes_ingress_v1.ingress.status[0].load_balancer[0].ingress[0].ip, "")
}

output "load_balancer_hostname" {
  description = "The hostname the controller's load balancer exposes; empty until a controller reconciles the Ingress"
  value       = try(kubernetes_ingress_v1.ingress.status[0].load_balancer[0].ingress[0].hostname, "")
}

output "first_host" {
  description = "The first host declared in the rules — the primary public FQDN this Ingress serves"
  value       = local.first_host
}
