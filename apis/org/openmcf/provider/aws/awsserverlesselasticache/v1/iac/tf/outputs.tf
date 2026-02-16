output "arn" {
  description = "The ARN of the ElastiCache Serverless cache."
  value       = aws_elasticache_serverless_cache.this.arn
}

output "endpoint_address" {
  description = "The primary connection endpoint DNS address."
  value       = try(aws_elasticache_serverless_cache.this.endpoint[0].address, "")
}

output "endpoint_port" {
  description = "The port of the primary connection endpoint."
  value       = try(aws_elasticache_serverless_cache.this.endpoint[0].port, 0)
}

output "reader_endpoint_address" {
  description = "The reader endpoint DNS address (Redis/Valkey only, empty for Memcached)."
  value       = try(aws_elasticache_serverless_cache.this.reader_endpoint[0].address, "")
}

output "reader_endpoint_port" {
  description = "The reader endpoint port."
  value       = try(aws_elasticache_serverless_cache.this.reader_endpoint[0].port, 0)
}

output "full_engine_version" {
  description = "The full engine version string deployed."
  value       = aws_elasticache_serverless_cache.this.full_engine_version
}

output "name" {
  description = "The name of the serverless cache."
  value       = aws_elasticache_serverless_cache.this.name
}
