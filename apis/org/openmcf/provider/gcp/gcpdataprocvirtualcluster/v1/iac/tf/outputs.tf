###############################################################################
# Outputs
###############################################################################

output "cluster_id" {
  description = "Fully qualified Dataproc cluster resource name"
  value       = google_dataproc_cluster.virtual_cluster.id
}

output "cluster_name" {
  description = "Short name of the Dataproc cluster"
  value       = google_dataproc_cluster.virtual_cluster.name
}

output "cluster_uuid" {
  description = "Server-generated UUID for the cluster (not directly exposed by Terraform; use cluster_id for references)"
  value       = ""
}
