output "name" {
  description = "Full resource name of the connection (projects/{project}/locations/{location}/connections/{connection_id})"
  value       = google_bigquery_connection.this.name
}

output "connection_id" {
  description = "The connection's id -- what a table, routine, or remote model names as {project}.{location}.{connection_id}"
  value       = google_bigquery_connection.this.connection_id
}

output "location" {
  description = "The connection's location"
  value       = google_bigquery_connection.this.location
}

# The Google-owned identities below exist only for the declared arm; the
# others are empty.
output "cloud_resource_service_account_id" {
  description = "cloud_resource arm: the service account BigQuery acts as -- grant it access to what it reads"
  value       = try(google_bigquery_connection.this.cloud_resource[0].service_account_id, "")
}

output "spark_service_account_id" {
  description = "spark arm: the service account the Spark procedures run as"
  value       = try(google_bigquery_connection.this.spark[0].service_account_id, "")
}

output "cloud_sql_service_account_id" {
  description = "cloud_sql arm: the service account BigQuery connects to the instance as"
  value       = try(google_bigquery_connection.this.cloud_sql[0].service_account_id, "")
}

output "connector_service_account" {
  description = "configuration arm: the service account the connector authenticates with"
  value       = try(google_bigquery_connection.this.configuration[0].authentication[0].service_account, "")
}

output "aws_identity" {
  description = "aws arm: the Google-owned identity your IAM role must trust"
  value       = try(google_bigquery_connection.this.aws[0].access_role[0].identity, "")
}

output "azure_identity" {
  description = "azure arm: the Google-owned identity your federated credential trusts"
  value       = try(google_bigquery_connection.this.azure[0].identity, "")
}

output "azure_application" {
  description = "azure arm: the Azure AD application Google created"
  value       = try(google_bigquery_connection.this.azure[0].application, "")
}

output "azure_client_id" {
  description = "azure arm: that application's client id"
  value       = try(google_bigquery_connection.this.azure[0].client_id, "")
}

output "azure_object_id" {
  description = "azure arm: that application's object id"
  value       = try(google_bigquery_connection.this.azure[0].object_id, "")
}

output "azure_redirect_uri" {
  description = "azure arm: the consent redirect URL for that application"
  value       = try(google_bigquery_connection.this.azure[0].redirect_uri, "")
}
