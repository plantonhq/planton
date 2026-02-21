output "db_system_id" {
  description = "OCID of the MySQL DB System"
  value       = oci_mysql_mysql_db_system.this.id
}

output "endpoint_hostname" {
  description = "Hostname of the primary (read/write) endpoint"
  value       = try(oci_mysql_mysql_db_system.this.endpoints[0].hostname, "")
}

output "endpoint_ip_address" {
  description = "Private IP address of the primary (read/write) endpoint"
  value       = try(oci_mysql_mysql_db_system.this.endpoints[0].ip_address, "")
}

output "endpoint_port" {
  description = "TCP port of the primary (read/write) endpoint"
  value       = try(oci_mysql_mysql_db_system.this.endpoints[0].port, 0)
}
