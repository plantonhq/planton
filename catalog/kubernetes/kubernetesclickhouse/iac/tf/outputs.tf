# Outputs — identical names and derivations in the Pulumi module
# (KubernetesClickHouseOutputs).

output "namespace" {
  description = "Namespace the cluster runs in"
  value       = local.namespace
}

output "chi_name" {
  description = "Name of the ClickHouseInstallation resource (equals metadata.name) — every operator-created object is derived from it"
  value       = local.chi_name
}

output "cluster_name" {
  description = "Logical ClickHouse cluster name — the `ON CLUSTER` / remote_servers target"
  value       = local.cluster_name
}

output "service_name" {
  description = "Name of the cluster-wide client Service covering all hosts (operator naming contract: clickhouse-<name>)"
  value       = local.service_name
}

output "tcp_endpoint" {
  description = "In-cluster native-protocol endpoint (clickhouse-client, drivers) on port 9000"
  value       = local.tcp_endpoint
}

output "http_endpoint" {
  description = "In-cluster HTTP interface endpoint (curl, JDBC/ODBC over HTTP) on port 8123"
  value       = local.http_endpoint
}

output "auth_secret_name" {
  description = "Module-managed Secret holding the provisioned users' passwords (one key per user name) — empty when no users are declared"
  value       = local.auth_secret_name
}

output "keeper_name" {
  description = "Name of the managed ClickHouseKeeperInstallation (`<name>-keeper`) — empty when coordination is external or none"
  value       = local.keeper_name
}

output "keeper_service_name" {
  description = "Name of the managed Keeper's client Service (operator naming contract: keeper-<keeper_name>) — empty when coordination is external or none"
  value       = local.keeper_service_name
}

output "port_forward_command" {
  description = "Port-forward command for reaching the HTTP interface from a workstation when no exposure is composed"
  value       = local.port_forward_command
}
