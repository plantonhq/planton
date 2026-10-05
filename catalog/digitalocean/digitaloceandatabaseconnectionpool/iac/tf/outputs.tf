# Outputs — exactly the DigitalOceanDatabaseConnectionPoolOutputs
# contract, identical across both provisioners. The (cluster, name) pair is
# the pool's API identity. The connection URIs are assembled by the
# provider from live connection details and embed credentials -- verifiers
# assert host/port, never URI equality.

output "cluster_id" {
  description = "UUID of the PostgreSQL cluster the pool runs on"
  value       = digitalocean_database_connection_pool.pool.cluster_id
}

output "pool_name" {
  description = "Name of the connection pool (clients connect to it as a database name)"
  value       = digitalocean_database_connection_pool.pool.name
}

output "host" {
  description = "Public hostname of the pool endpoint"
  value       = digitalocean_database_connection_pool.pool.host
}

output "private_host" {
  description = "Private-network hostname of the pool endpoint (same-VPC access)"
  value       = digitalocean_database_connection_pool.pool.private_host
}

output "port" {
  description = "Port the pool listens on (distinct from the cluster's own port)"
  value       = digitalocean_database_connection_pool.pool.port
}

output "uri" {
  description = "Full public connection URI for the pool (includes credentials)"
  value       = digitalocean_database_connection_pool.pool.uri
  sensitive   = true
}

output "private_uri" {
  description = "Full private-network connection URI for the pool (includes credentials)"
  value       = digitalocean_database_connection_pool.pool.private_uri
  sensitive   = true
}

output "password" {
  description = "Password of the pool's user (empty for inbound-user pools)"
  value       = digitalocean_database_connection_pool.pool.password
  sensitive   = true
}
