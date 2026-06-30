output "db_system_id" {
  description = "OCID of the PostgreSQL DB System"
  value       = oci_psql_db_system.this.id
}

output "primary_db_endpoint_private_ip" {
  description = "Private IP address of the primary (read-write) endpoint"
  value       = try(oci_psql_db_system.this.network_details[0].primary_db_endpoint_private_ip, "")
}

output "admin_username" {
  description = "Administrator username for the PostgreSQL DB System"
  value       = oci_psql_db_system.this.admin_username
}
