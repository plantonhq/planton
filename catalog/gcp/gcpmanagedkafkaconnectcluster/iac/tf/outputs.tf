output "name" {
  description = "Full resource name of the Connect cluster (projects/{project}/locations/{location}/connectClusters/{connect_cluster_id}) -- what connectors reference"
  value       = google_managed_kafka_connect_cluster.this.name
}

output "connect_cluster_id" {
  description = "The Connect cluster's id"
  value       = google_managed_kafka_connect_cluster.this.connect_cluster_id
}

output "location" {
  description = "The region the workers run in"
  value       = google_managed_kafka_connect_cluster.this.location
}
