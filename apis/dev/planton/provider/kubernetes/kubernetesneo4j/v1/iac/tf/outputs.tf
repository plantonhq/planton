output "namespace" {
  description = "The Kubernetes namespace where the Neo4j instance is deployed."
  value       = local.namespace
}

output "service" {
  description = "The in-cluster Service name for connecting via Bolt/HTTP."
  value       = local.service_name
}

output "bolt_uri_kube_endpoint" {
  description = "Bolt URI (internal) for database connections. Example: bolt://my-graph-db.default.svc.cluster.local:7687"
  value       = local.bolt_uri
}

output "http_uri_kube_endpoint" {
  description = "HTTP URL for the Neo4j browser (internal). Example: http://my-graph-db.default.svc.cluster.local:7474"
  value       = local.http_uri
}

output "port_forward_command" {
  description = "Command to port-forward the Neo4j browser to localhost:7474 for local development."
  value       = local.port_forward_command
}

output "username" {
  description = "The default Neo4j username."
  value       = "neo4j"
}

output "password_secret_name" {
  description = "Name of the Kubernetes secret containing the Neo4j password."
  value       = local.password_secret_name
}

output "password_secret_key" {
  description = "Key within the secret that contains the Neo4j password."
  value       = "neo4j-password"
}

output "external_hostname" {
  description = "The external hostname for Neo4j if ingress is enabled."
  value       = local.ingress_enabled ? local.ingress_external_hostname : null
}

