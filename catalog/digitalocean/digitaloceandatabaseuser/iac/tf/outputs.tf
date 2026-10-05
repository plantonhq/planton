# Outputs — exactly the DigitalOceanDatabaseUserOutputs
# contract, identical across both provisioners. The (cluster, name) pair is
# the user's API identity; DigitalOcean mints no standalone user id.

output "cluster_id" {
  description = "UUID of the database cluster the user belongs to"
  value       = digitalocean_database_user.user.cluster_id
}

output "user_name" {
  description = "Name of the database user (its API identity within the cluster)"
  value       = digitalocean_database_user.user.name
}

output "role" {
  description = "Role DigitalOcean assigned to the user (normally \"normal\")"
  value       = digitalocean_database_user.user.role
}

output "password" {
  description = "Server-generated password for the user"
  value       = digitalocean_database_user.user.password
  sensitive   = true
}

# The client certificate is public; its private half is access_key. The
# provider marks the attribute sensitive; nonsensitive() unwraps it.
output "access_cert" {
  description = "Kafka only: PEM access certificate for mutual TLS"
  value       = nonsensitive(digitalocean_database_user.user.access_cert)
}

output "access_key" {
  description = "Kafka only: PEM access key paired with access_cert"
  value       = digitalocean_database_user.user.access_key
  sensitive   = true
}
