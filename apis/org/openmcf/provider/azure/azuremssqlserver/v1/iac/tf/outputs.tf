output "server_id" {
  description = "The Azure Resource Manager ID of the SQL Server"
  value       = azurerm_mssql_server.main.id
}

output "server_name" {
  description = "The name of the SQL Server"
  value       = azurerm_mssql_server.main.name
}

output "fqdn" {
  description = "The fully qualified domain name of the server"
  value       = azurerm_mssql_server.main.fully_qualified_domain_name
}

output "administrator_login" {
  description = "The administrator login name"
  value       = azurerm_mssql_server.main.administrator_login
}

output "database_ids" {
  description = "Map of database names to their Azure Resource Manager IDs"
  value       = { for name, db in azurerm_mssql_database.databases : name => db.id }
}
