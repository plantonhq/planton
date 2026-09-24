output "name" {
  description = "Full resource name of the connector (projects/{project}/locations/{location}/connectClusters/{connect_cluster}/connectors/{connector_id})"
  value       = google_managed_kafka_connector.this.name
}

output "connector_id" {
  description = "The connector's id"
  value       = google_managed_kafka_connector.this.connector_id
}
