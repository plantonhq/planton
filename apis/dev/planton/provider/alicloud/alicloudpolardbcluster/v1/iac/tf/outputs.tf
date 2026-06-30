output "cluster_id" {
  description = "The PolarDB cluster ID"
  value       = alicloud_polardb_cluster.main.id
}

output "connection_string" {
  description = "The primary endpoint connection string"
  value       = alicloud_polardb_cluster.main.connection_string
}

output "port" {
  description = "The database service port"
  value       = alicloud_polardb_cluster.main.port
}

output "database_ids" {
  description = "Map of database names to their IDs"
  value = {
    for name, db in alicloud_polardb_database.databases : name => db.id
  }
}
