output "name" {
  description = "Full resource name of the cluster (projects/{project}/locations/{location}/clusters/{cluster_id}) -- what topics, ACLs, and Connect clusters reference"
  value       = google_managed_kafka_cluster.this.name
}

output "cluster_id" {
  description = "The cluster's id"
  value       = google_managed_kafka_cluster.this.cluster_id
}

output "location" {
  description = "The region the cluster runs in"
  value       = google_managed_kafka_cluster.this.location
}
