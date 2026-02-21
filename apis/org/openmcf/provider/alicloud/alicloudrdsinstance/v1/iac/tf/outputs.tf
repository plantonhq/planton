output "instance_id" {
  description = "The RDS instance ID"
  value       = alicloud_db_instance.main.id
}

output "connection_string" {
  description = "The intranet connection endpoint"
  value       = alicloud_db_instance.main.connection_string
}

output "port" {
  description = "The database service port"
  value       = alicloud_db_instance.main.port
}

output "database_ids" {
  description = "Map of database names to their IDs"
  value = {
    for name, db in alicloud_db_database.databases : name => db.id
  }
}
