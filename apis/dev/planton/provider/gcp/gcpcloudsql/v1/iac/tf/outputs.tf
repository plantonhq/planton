output "instance_name" {
  description = "Name of the Cloud SQL instance"
  value       = google_sql_database_instance.instance.name
}

output "connection_name" {
  description = "Full connection name in the format project:region:instance"
  value       = google_sql_database_instance.instance.connection_name
}

output "private_ip" {
  description = "Private IP address of the instance (if enabled)"
  value       = length(google_sql_database_instance.instance.private_ip_address) > 0 ? google_sql_database_instance.instance.private_ip_address : null
}

output "public_ip" {
  description = "Public IP address of the instance"
  value       = length(google_sql_database_instance.instance.public_ip_address) > 0 ? google_sql_database_instance.instance.public_ip_address : null
}

output "self_link" {
  description = "GCP resource self link for the Cloud SQL instance"
  value       = google_sql_database_instance.instance.self_link
}

