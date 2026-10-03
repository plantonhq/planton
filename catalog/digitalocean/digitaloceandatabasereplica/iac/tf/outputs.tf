# Outputs — exactly the DigitalOceanDatabaseReplicaOutputs
# contract, identical across both provisioners. DigitalOcean reads and
# deletes replicas by (cluster, name); the replica's own UUID (the uuid
# attribute -- the Terraform state id is a legacy composite string) is
# exported for resize operations and cross-references. The connection URIs
# embed credentials -- verifiers assert host/port, never URI equality.

output "replica_id" {
  description = "UUID of the replica itself"
  value       = digitalocean_database_replica.replica.uuid
}

output "cluster_id" {
  description = "UUID of the primary database cluster this replica follows"
  value       = digitalocean_database_replica.replica.cluster_id
}

output "replica_name" {
  description = "Name of the replica (its API identity within the cluster)"
  value       = digitalocean_database_replica.replica.name
}

output "host" {
  description = "Public hostname of the replica endpoint"
  value       = digitalocean_database_replica.replica.host
}

output "private_host" {
  description = "Private-network hostname of the replica endpoint (same-VPC access)"
  value       = digitalocean_database_replica.replica.private_host
}

output "port" {
  description = "Port the replica listens on"
  value       = digitalocean_database_replica.replica.port
}

output "database" {
  description = "Name of the default database served by the replica"
  value       = digitalocean_database_replica.replica.database
}

output "user" {
  description = "Username of the replica's default user"
  value       = digitalocean_database_replica.replica.user
}

output "password" {
  description = "Password of the replica's default user"
  value       = digitalocean_database_replica.replica.password
  sensitive   = true
}

output "uri" {
  description = "Full public connection URI for the replica (includes credentials)"
  value       = digitalocean_database_replica.replica.uri
  sensitive   = true
}

output "private_uri" {
  description = "Full private-network connection URI for the replica (includes credentials)"
  value       = digitalocean_database_replica.replica.private_uri
  sensitive   = true
}
