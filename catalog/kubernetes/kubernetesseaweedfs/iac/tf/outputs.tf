# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.
#
# Service names derive from the chart's componentName helper with
# fullnameOverride pinned to the resource name: `<name>-master`,
# `<name>-filer`, `<name>-s3`, `<name>-admin`. The `-s3` Service exists for
# the embedded and dedicated gateway shapes alike.

output "namespace" {
  description = "Kubernetes namespace the store runs in"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (= metadata.name)"
  value       = local.release_name
}

output "s3_endpoint" {
  description = "In-cluster S3 endpoint clients point their SDKs at; empty when the gateway is disabled"
  value       = local.s3_enabled ? "http://${local.s3_service_name}.${local.namespace}.svc.cluster.local:8333" : ""
}

output "s3_credentials_secret_name" {
  description = "Secret holding the S3 credentials (admin_access_key_id / admin_secret_access_key / read_access_key_id / read_secret_access_key); empty when auth is disabled"
  value       = local.s3_credentials_secret_name
}

output "filer_service_name" {
  description = "Name of the filer Service (file namespace HTTP API, port 8888)"
  value       = local.filer_service_name
}

output "master_service_name" {
  description = "Name of the master Service (cluster coordination, port 9333)"
  value       = local.master_service_name
}

output "admin_endpoint" {
  description = "In-cluster admin console endpoint; empty when the console is disabled"
  value       = local.admin_enabled ? "http://${local.admin_service_name}.${local.namespace}.svc.cluster.local:23646" : ""
}

output "admin_auth_secret_name" {
  description = "Secret holding the admin-console credentials (keys user/password); empty when the console is disabled"
  value       = local.admin_auth_secret_name
}

output "port_forward_command" {
  description = "kubectl one-liner for reaching S3 from a workstation"
  value       = local.s3_enabled ? "kubectl port-forward svc/${local.s3_service_name} -n ${local.namespace} 8333:8333" : ""
}
