output "server_id" {
  description = "The Azure Resource Manager ID of the MySQL Flexible Server"
  value       = azurerm_mysql_flexible_server.main.id
}

output "server_name" {
  description = "The name of the MySQL Flexible Server"
  value       = azurerm_mysql_flexible_server.main.name
}

output "fqdn" {
  description = "The fully qualified domain name of the server"
  value       = azurerm_mysql_flexible_server.main.fqdn
}

output "administrator_login" {
  description = "The administrator login name"
  value       = azurerm_mysql_flexible_server.main.administrator_login
}

output "database_ids" {
  description = "Map of database names to their Azure Resource Manager IDs"
  value       = { for name, db in azurerm_mysql_flexible_database.databases : name => db.id }
}
