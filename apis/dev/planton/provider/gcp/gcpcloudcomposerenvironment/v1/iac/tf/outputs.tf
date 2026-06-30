output "environment_id" {
  description = "Fully qualified resource ID of the Composer environment"
  value       = google_composer_environment.environment.id
}

output "environment_name" {
  description = "Short name of the Composer environment"
  value       = google_composer_environment.environment.name
}

output "airflow_uri" {
  description = "URI of the Apache Airflow web UI"
  value       = try(google_composer_environment.environment.config[0].airflow_uri, "")
}

output "dag_gcs_prefix" {
  description = "Cloud Storage prefix for DAG file uploads"
  value       = try(google_composer_environment.environment.config[0].dag_gcs_prefix, "")
}

output "gke_cluster" {
  description = "Name of the underlying GKE cluster managed by Composer"
  value       = try(google_composer_environment.environment.config[0].gke_cluster, "")
}
