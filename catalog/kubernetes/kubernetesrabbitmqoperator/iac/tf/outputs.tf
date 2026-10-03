# Outputs — identical names and derivations in the Pulumi module
# (KubernetesRabbitMqOperatorOutputs).

output "namespace" {
  description = "Namespace the operator is installed into (always rabbitmq-system — the release manifest's fixed namespace)"
  value       = local.namespace
}

output "deployment_name" {
  description = "Name of the operator Deployment (rabbitmq-cluster-operator — the release manifest's fixed name)"
  value       = local.deployment_name
}

output "metrics_endpoint" {
  description = "In-cluster Prometheus metrics endpoint of the operator"
  value       = local.metrics_endpoint
}

output "crd_name" {
  description = "Name of the RabbitmqCluster CustomResourceDefinition the operator serves — deleted with this resource (see the CRD-lifecycle note on the spec)"
  value       = local.crd_name
}
