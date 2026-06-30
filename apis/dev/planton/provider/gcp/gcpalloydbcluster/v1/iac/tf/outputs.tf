output "cluster_id" {
  description = "Fully qualified cluster resource name"
  value       = google_alloydb_cluster.cluster.name
}

output "cluster_name" {
  description = "Short name of the cluster"
  value       = var.spec.cluster_name
}

output "primary_instance_ip" {
  description = "Private IP address of the primary instance"
  value       = google_alloydb_instance.primary.ip_address
}

output "primary_instance_name" {
  description = "Fully qualified primary instance resource name"
  value       = google_alloydb_instance.primary.name
}

output "database_version" {
  description = "Computed database engine version"
  value       = google_alloydb_cluster.cluster.database_version
}

output "state" {
  description = "Current state of the cluster"
  value       = google_alloydb_cluster.cluster.state
}
