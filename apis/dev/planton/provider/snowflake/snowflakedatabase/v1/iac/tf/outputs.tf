# Terraform Outputs
# These outputs map to the SnowflakeDatabaseStackOutputs proto definition

output "id" {
  description = "The provider-assigned unique ID for the Snowflake database resource"
  value       = snowflake_database.this.id
}

# Note: The stack_outputs.proto currently has fields for bootstrap_endpoint, crn, and rest_endpoint
# which appear to be copied from a Kafka cluster resource. These are not applicable to Snowflake databases.
# For now, we provide placeholder values to satisfy the proto contract, but these should be updated
# in the proto definition to reflect actual Snowflake database outputs.

output "bootstrap_endpoint" {
  description = "Not applicable for Snowflake databases - placeholder for proto compatibility"
  value       = ""
}

output "crn" {
  description = "Not applicable for Snowflake databases - placeholder for proto compatibility"
  value       = ""
}

output "rest_endpoint" {
  description = "Not applicable for Snowflake databases - placeholder for proto compatibility"
  value       = ""
}

# Additional useful outputs not in the proto but helpful for users
output "name" {
  description = "The name of the Snowflake database"
  value       = snowflake_database.this.name
}

output "is_transient" {
  description = "Whether the database is transient (cost optimization indicator)"
  value       = snowflake_database.this.is_transient
}

output "data_retention_time_in_days" {
  description = "Number of days for Time Travel retention (cost indicator)"
  value       = snowflake_database.this.data_retention_time_in_days
}




