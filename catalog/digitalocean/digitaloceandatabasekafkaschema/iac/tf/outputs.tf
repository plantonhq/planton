# Outputs — exactly the DigitalOceanDatabaseKafkaSchemaOutputs
# contract, identical across both provisioners. The (cluster, subject name)
# pair is the subject's API identity; the registry's internal numeric
# schema id is discarded by the provider and deliberately not exported.

output "cluster_id" {
  description = "UUID of the Kafka database cluster whose registry holds the subject"
  value       = digitalocean_database_kafka_schema_registry.schema.cluster_id
}

output "subject_name" {
  description = "Name of the registered schema subject (its API identity within the registry)"
  value       = digitalocean_database_kafka_schema_registry.schema.subject_name
}
