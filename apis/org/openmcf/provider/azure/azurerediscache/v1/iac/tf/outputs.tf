output "redis_id" {
  description = "The Azure Resource Manager ID of the Redis cache"
  value       = azurerm_redis_cache.main.id
}

output "hostname" {
  description = "The hostname of the Redis cache"
  value       = azurerm_redis_cache.main.hostname
}

output "ssl_port" {
  description = "The SSL port of the Redis cache (always 6380)"
  value       = azurerm_redis_cache.main.ssl_port
}

output "primary_access_key" {
  description = "The primary access key for authentication"
  value       = azurerm_redis_cache.main.primary_access_key
  sensitive   = true
}

output "primary_connection_string" {
  description = "The primary connection string for the Redis cache"
  value       = azurerm_redis_cache.main.primary_connection_string
  sensitive   = true
}
