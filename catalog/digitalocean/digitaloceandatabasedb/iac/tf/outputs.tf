# Outputs — exactly the DigitalOceanDatabaseDbOutputs contract,
# identical across both provisioners. The (cluster, name) pair is the
# logical database's API identity; DigitalOcean mints no standalone id.

output "cluster_id" {
  description = "UUID of the database cluster the logical database lives in"
  value       = digitalocean_database_db.database.cluster_id
}

output "database_name" {
  description = "Name of the logical database (its API identity within the cluster)"
  value       = digitalocean_database_db.database.name
}
