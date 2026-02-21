output "autonomous_database_id" {
  description = "OCID of the autonomous database"
  value       = oci_database_autonomous_database.this.id
}

output "connection_string_high" {
  description = "High-priority connection string for latency-sensitive workloads"
  value       = try(oci_database_autonomous_database.this.connection_strings[0].high, "")
}

output "connection_string_medium" {
  description = "Medium-priority connection string for typical application workloads"
  value       = try(oci_database_autonomous_database.this.connection_strings[0].medium, "")
}

output "connection_string_low" {
  description = "Low-priority connection string for batch and background workloads"
  value       = try(oci_database_autonomous_database.this.connection_strings[0].low, "")
}

output "service_console_url" {
  description = "URL of the OCI Service Console for this database"
  value       = oci_database_autonomous_database.this.service_console_url
}

output "private_endpoint" {
  description = "Private endpoint IP address (empty when not configured)"
  value       = oci_database_autonomous_database.this.private_endpoint
}
