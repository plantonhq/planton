# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports (KubernetesSolrOutputs).

output "namespace" {
  description = "Namespace the cluster runs in"
  value       = local.namespace
}

output "cluster_name" {
  description = "Name of the SolrCloud resource (equals metadata.name) — every operator-created object is prefixed with it"
  value       = local.cluster_name
}

output "common_service_name" {
  description = "Name of the common Service fronting all Solr nodes (`<name>-solrcloud-common`)"
  value       = local.common_service_name
}

output "internal_endpoint" {
  description = "In-cluster base URL of the cluster through the common service — http (port 80) without TLS, https (port 443) with TLS"
  value       = local.internal_endpoint
}

output "basic_auth_secret_name" {
  description = "Name of the operator-generated basic-auth Secret (`<name>-solrcloud-basic-auth`) — empty when security is disabled or a user-provided basic_auth_secret is in play"
  value       = local.basic_auth_secret_name_output
}

output "zookeeper_connection_string" {
  description = "ZooKeeper connection string the cluster uses (host:port/chroot) — the external ensemble's, or the provided ensemble's operator-named client service"
  value       = local.zookeeper_connection_string
}

output "port_forward_command" {
  description = "Port-forward command for reaching the common service from a workstation when no exposure is composed"
  value       = "kubectl port-forward svc/${local.common_service_name} -n ${local.namespace} 8983:${local.common_service_port}"
}
