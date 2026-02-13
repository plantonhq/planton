output "account_id" {
  description = "The Azure Resource Manager ID of the Cosmos DB account"
  value       = azurerm_cosmosdb_account.main.id
}

output "account_name" {
  description = "The name of the Cosmos DB account"
  value       = azurerm_cosmosdb_account.main.name
}

output "endpoint" {
  description = "The document endpoint URI"
  value       = azurerm_cosmosdb_account.main.endpoint
}

output "primary_key" {
  description = "The primary access key for the account"
  value       = azurerm_cosmosdb_account.main.primary_key
  sensitive   = true
}

output "primary_connection_string" {
  description = "The primary connection string for the SQL/NoSQL API"
  value       = azurerm_cosmosdb_account.main.primary_sql_connection_string
  sensitive   = true
}

output "primary_mongodb_connection_string" {
  description = "The primary connection string for the MongoDB API"
  value       = azurerm_cosmosdb_account.main.primary_mongodb_connection_string
  sensitive   = true
}

output "database_ids" {
  description = "Map of database names to their Azure Resource Manager IDs"
  value = merge(
    var.spec.kind == "GlobalDocumentDB" ? { for name, db in azurerm_cosmosdb_sql_database.sql_databases : name => db.id } : {},
    var.spec.kind == "MongoDB" ? { for name, db in azurerm_cosmosdb_mongo_database.mongo_databases : name => db.id } : {}
  )
}
