output "host" {
  description = "Hostname or IP address of the primary Redis endpoint"
  value       = google_redis_instance.this.host
}

output "port" {
  description = "Port number of the primary Redis endpoint"
  value       = google_redis_instance.this.port
}

output "read_endpoint" {
  description = "Hostname or IP address of the read replica endpoint (STANDARD_HA with read replicas only)"
  value       = google_redis_instance.this.read_endpoint
}

output "read_endpoint_port" {
  description = "Port number of the read replica endpoint"
  value       = google_redis_instance.this.read_endpoint_port
}

output "current_location_id" {
  description = "Zone where the Redis primary is currently running"
  value       = google_redis_instance.this.current_location_id
}

output "auth_string" {
  description = "Redis AUTH string (populated only when auth_enabled is true)"
  value       = google_redis_instance.this.auth_string
  sensitive   = true
}
