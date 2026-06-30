output "cluster_id" {
  description = "Fully qualified cluster resource name"
  value       = google_dataproc_cluster.cluster.id
}

output "cluster_name" {
  description = "Short name of the cluster"
  value       = var.spec.cluster_name
}

output "cluster_uuid" {
  description = "Server-generated unique identifier for the cluster"
  value       = ""
}

output "staging_bucket" {
  description = "Cloud Storage bucket used for staging job dependencies"
  value       = try(google_dataproc_cluster.cluster.cluster_config[0].bucket, "")
}
