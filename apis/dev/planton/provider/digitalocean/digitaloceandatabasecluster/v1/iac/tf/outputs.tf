output "cluster_id" {
  description = "The unique identifier (UUID) of the database cluster"
  value       = digitalocean_database_cluster.cluster.id
}

output "cluster_name" {
  description = "The name of the database cluster"
  value       = digitalocean_database_cluster.cluster.name
}

output "engine" {
  description = "The database engine (pg, mysql, redis, mongodb)"
  value       = digitalocean_database_cluster.cluster.engine
}

output "version" {
  description = "The engine version"
  value       = digitalocean_database_cluster.cluster.version
}

output "host" {
  description = "The hostname for database connections"
  value       = digitalocean_database_cluster.cluster.host
}

output "private_host" {
  description = "The private hostname (VPC-internal) for database connections"
  value       = digitalocean_database_cluster.cluster.private_host
}

output "port" {
  description = "The port for database connections"
  value       = digitalocean_database_cluster.cluster.port
}

output "database_name" {
  description = "The default database name"
  value       = digitalocean_database_cluster.cluster.database
}

output "username" {
  description = "The admin username"
  value       = digitalocean_database_cluster.cluster.user
  sensitive   = true
}

output "password" {
  description = "The admin password"
  value       = digitalocean_database_cluster.cluster.password
  sensitive   = true
}

output "connection_uri" {
  description = "The full connection URI (includes credentials)"
  value       = digitalocean_database_cluster.cluster.uri
  sensitive   = true
}

output "private_uri" {
  description = "The private connection URI (VPC-internal, includes credentials)"
  value       = digitalocean_database_cluster.cluster.private_uri
  sensitive   = true
}

output "region" {
  description = "The region where the cluster is deployed"
  value       = digitalocean_database_cluster.cluster.region
}

output "node_count" {
  description = "The number of nodes in the cluster"
  value       = digitalocean_database_cluster.cluster.node_count
}

output "size" {
  description = "The size slug of the cluster nodes"
  value       = digitalocean_database_cluster.cluster.size
}

output "status" {
  description = "The current status of the cluster"
  value       = digitalocean_database_cluster.cluster.status
}

output "created_at" {
  description = "The timestamp when the cluster was created"
  value       = digitalocean_database_cluster.cluster.created_at
}

output "vpc_uuid" {
  description = "The VPC UUID (if cluster is in a VPC)"
  value       = digitalocean_database_cluster.cluster.private_network_uuid
}

