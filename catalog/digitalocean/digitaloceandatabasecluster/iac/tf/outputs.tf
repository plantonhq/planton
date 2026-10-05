# Outputs — exactly the DigitalOceanDatabaseClusterOutputs
# contract, identical across both provisioners.

output "cluster_id" {
  description = "The unique identifier (UUID) of the database cluster"
  value       = digitalocean_database_cluster.cluster.id
}

output "connection_uri" {
  description = "The full public connection URI (includes credentials and database name)"
  value       = digitalocean_database_cluster.cluster.uri
  sensitive   = true
}

output "host" {
  description = "The public hostname for database connections"
  value       = digitalocean_database_cluster.cluster.host
}

output "port" {
  description = "The port for database connections"
  value       = digitalocean_database_cluster.cluster.port
}

output "database_user" {
  description = "The username for the cluster's default database user"
  value       = digitalocean_database_cluster.cluster.user
}

output "database_password" {
  description = "The password for the cluster's default database user"
  value       = digitalocean_database_cluster.cluster.password
  sensitive   = true
}

output "private_host" {
  description = "The private-network hostname, reachable from the same VPC"
  value       = digitalocean_database_cluster.cluster.private_host
}

output "private_uri" {
  description = "The private-network connection URI (includes credentials)"
  value       = digitalocean_database_cluster.cluster.private_uri
  sensitive   = true
}

output "database_name" {
  description = "The name of the cluster's default database"
  value       = digitalocean_database_cluster.cluster.database
}

output "ui_host" {
  description = "OpenSearch only: hostname of the OpenSearch Dashboards endpoint"
  value       = digitalocean_database_cluster.cluster.ui_host
}

output "ui_port" {
  description = "OpenSearch only: port of the OpenSearch Dashboards endpoint"
  value       = digitalocean_database_cluster.cluster.ui_port
}

output "ui_uri" {
  description = "OpenSearch only: OpenSearch Dashboards connection URI (includes credentials)"
  value       = digitalocean_database_cluster.cluster.ui_uri
  sensitive   = true
}

output "ui_database" {
  description = "OpenSearch only: default database of the Dashboards connection"
  value       = digitalocean_database_cluster.cluster.ui_database
}

output "ui_user" {
  description = "OpenSearch only: username for OpenSearch Dashboards"
  value       = digitalocean_database_cluster.cluster.ui_user
}

output "ui_password" {
  description = "OpenSearch only: password for OpenSearch Dashboards"
  value       = digitalocean_database_cluster.cluster.ui_password
  sensitive   = true
}
