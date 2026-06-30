# Confluent Kafka Cluster Outputs
# https://registry.terraform.io/providers/confluentinc/confluent/latest/docs/resources/confluent_kafka_cluster#attributes-reference

output "id" {
  description = "The provider-assigned unique ID for this managed resource"
  value       = confluent_kafka_cluster.main.id
}

output "bootstrap_endpoint" {
  description = "The bootstrap endpoint used by Kafka clients to connect to the Kafka cluster (e.g., SASL_SSL://pkc-00000.us-central1.gcp.confluent.cloud:9092)"
  value       = confluent_kafka_cluster.main.bootstrap_endpoint
}

output "crn" {
  description = "The Confluent Resource Name (CRN) of the Kafka cluster"
  value       = confluent_kafka_cluster.main.rbac_crn
}

output "rest_endpoint" {
  description = "The REST endpoint of the Kafka cluster (e.g., https://pkc-00000.us-central1.gcp.confluent.cloud:443)"
  value       = confluent_kafka_cluster.main.rest_endpoint
}

output "api_version" {
  description = "The API version of the Kafka cluster"
  value       = confluent_kafka_cluster.main.api_version
}

output "kind" {
  description = "The kind of the Kafka cluster"
  value       = confluent_kafka_cluster.main.kind
}

