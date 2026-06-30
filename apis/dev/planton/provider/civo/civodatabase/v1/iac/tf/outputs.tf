# Outputs for Civo Database Terraform Module

output "database_id" {
  description = "The ID of the created Civo database"
  value       = civo_database.this.id
}

output "database_name" {
  description = "The name of the database instance"
  value       = civo_database.this.name
}

output "dns_endpoint" {
  description = "The DNS endpoint for connecting to the database (recommended for HA)"
  value       = civo_database.this.dns_endpoint
}

output "host" {
  description = "The hostname/endpoint of the database"
  value       = civo_database.this.endpoint
}

output "port" {
  description = "The port number for database connections"
  value       = civo_database.this.port
}

output "username" {
  description = "The master username for database authentication"
  value       = civo_database.this.username
  sensitive   = true
}

output "password" {
  description = "The master password for database authentication"
  value       = civo_database.this.password
  sensitive   = true
}

output "status" {
  description = "The current status of the database instance"
  value       = civo_database.this.status
}

output "network_id" {
  description = "The ID of the private network the database is attached to"
  value       = civo_database.this.network_id
}

output "firewall_id" {
  description = "The ID of the firewall rule attached to the database"
  value       = civo_database.this.firewall_id
}

output "nodes" {
  description = "The total number of nodes in the database cluster (primary + replicas)"
  value       = civo_database.this.nodes
}

