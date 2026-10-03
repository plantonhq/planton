# Outputs — flattened onto KubernetesKafkaOutputs by the
# platform. Keep in lockstep with the Pulumi module's exports.

output "namespace" {
  description = "Kubernetes namespace the Kafka cluster deploys into"
  value       = local.namespace
}

output "cluster_name" {
  description = "Kafka cluster name (metadata.name) — the strimzi.io/cluster binding value for KafkaNodePool/KafkaTopic/KafkaUser resources"
  value       = local.cluster_name
}

output "bootstrap_service_name" {
  description = "Name of the internal bootstrap Service (<cluster>-kafka-bootstrap)"
  value       = local.bootstrap_service_name
}

output "internal_bootstrap_endpoint" {
  description = "In-cluster bootstrap address for the first internal listener (empty when the cluster declares no internal listener)"
  value       = local.internal_bootstrap_endpoint
}

output "cluster_ca_cert_secret_name" {
  description = "Name of the Secret holding the cluster CA certificate (<cluster>-cluster-ca-cert, key ca.crt)"
  value       = local.cluster_ca_cert_secret_name
}
