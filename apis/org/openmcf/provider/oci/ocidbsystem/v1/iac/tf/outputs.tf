output "db_system_id" {
  description = "OCID of the DB System"
  value       = oci_database_db_system.this.id
}

output "db_home_id" {
  description = "OCID of the first DB Home"
  value       = try(oci_database_db_system.this.db_home[0].id, "")
}

output "database_id" {
  description = "OCID of the initial database"
  value       = try(oci_database_db_system.this.db_home[0].database[0].id, "")
}

output "listener_port" {
  description = "TCP port for database listener connections"
  value       = try(oci_database_db_system.this.listener_port, 0)
}
