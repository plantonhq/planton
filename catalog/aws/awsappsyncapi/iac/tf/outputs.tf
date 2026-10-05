# Outputs - identical key set to the Pulumi module's exports.

output "api_id" {
  description = "The API's id - the pivot's import ID, every satellite composite's prefix, and the MERGED source join key."
  value       = local.api_id
}

output "api_arn" {
  description = "The API's ARN."
  value       = local.is_graphql ? aws_appsync_graphql_api.this[0].arn : aws_appsync_api.this[0].api_arn
}

output "graphql_url" {
  description = "GraphQL arm: the endpoint URL clients query."
  value       = local.is_graphql ? try(aws_appsync_graphql_api.this[0].uris["GRAPHQL"], "") : ""
}

output "realtime_url" {
  description = "GraphQL arm: the real-time (subscriptions) endpoint URL."
  value       = local.is_graphql ? try(aws_appsync_graphql_api.this[0].uris["REALTIME"], "") : ""
}

output "events_http_endpoint" {
  description = "Events arm: the HTTP endpoint domain clients publish through."
  value       = local.is_events ? try(aws_appsync_api.this[0].dns["HTTP"], "") : ""
}

output "events_realtime_endpoint" {
  description = "Events arm: the real-time (WebSocket) endpoint domain."
  value       = local.is_events ? try(aws_appsync_api.this[0].dns["REALTIME"], "") : ""
}

output "appsync_domain_name" {
  description = "The AppSync-managed domain to point DNS at when custom_domain is set."
  value       = var.spec.custom_domain != null ? aws_appsync_domain_name.this[0].appsync_domain_name : ""
}

output "domain_hosted_zone_id" {
  description = "The Route53 hosted zone id for alias records at the custom domain."
  value       = var.spec.custom_domain != null ? aws_appsync_domain_name.this[0].hosted_zone_id : ""
}

output "datasource_arns" {
  description = "Data source ARNs keyed by spec datasource name."
  value       = { for name, d in aws_appsync_datasource.this : name => d.arn }
}

output "function_ids" {
  description = "Function ids keyed by spec function name - part of each function's composite import ID."
  value       = { for name, f in aws_appsync_function.this : name => f.function_id }
}

output "api_key_ids" {
  description = "API key ids keyed by spec key name (the IDs, never the secrets - AWS returns a key's secret only at creation)."
  value       = { for name, k in aws_appsync_api_key.this : name => k.api_key_id }
}

output "channel_namespace_arns" {
  description = "Channel namespace ARNs keyed by namespace name (Events arm)."
  value       = { for name, n in aws_appsync_channel_namespace.this : name => n.channel_namespace_arn }
}

output "source_api_association_ids" {
  description = "Source API association ids keyed by spec entry name (MERGED APIs) - part of each association's composite import ID."
  value       = { for name, s in aws_appsync_source_api_association.this : name => s.association_id }
}

output "type_formats" {
  description = "Import-derivation echo map: each managed type's format keyed by type name (part of the type's composite import ID)."
  value       = local.is_graphql ? { for t in var.spec.graphql.types : t.name => t.format } : {}
}
