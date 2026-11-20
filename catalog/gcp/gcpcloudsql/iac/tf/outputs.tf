output "instance_name" {
  description = "Name of the Cloud SQL instance — the composition key databases, users, and replicas reference"
  value       = google_sql_database_instance.this.name
}

output "connection_name" {
  description = "Connection name (project:region:instance) consumed by the Auth Proxy and connectors"
  value       = google_sql_database_instance.this.connection_name
}

output "private_ip" {
  description = "Private IP address (empty unless private_network is configured)"
  value       = google_sql_database_instance.this.private_ip_address
}

output "public_ip" {
  description = "Public IPv4 address (empty unless ipv4_enabled)"
  value       = google_sql_database_instance.this.public_ip_address
}

output "self_link" {
  description = "Self-link URL of the Cloud SQL instance"
  value       = google_sql_database_instance.this.self_link
}

output "service_account_email" {
  description = "Google-managed service account the instance runs as — grant it GCS access for imports/exports and audit uploads"
  value       = google_sql_database_instance.this.service_account_email_address
}

output "dns_name" {
  description = "DNS name of the instance (populated for PSC-enabled instances)"
  value       = google_sql_database_instance.this.dns_name
}

output "socket_path" {
  description = "Unix socket path (/cloudsql/project:region:instance) a workload connects through when Cloud Run or the Auth Proxy mounts the instance"
  value       = "/cloudsql/${google_sql_database_instance.this.connection_name}"
}

output "psc_service_attachment_link" {
  description = "PSC service attachment consumers target with PSC endpoints (populated only when PSC is enabled)"
  value       = google_sql_database_instance.this.psc_service_attachment_link
}
