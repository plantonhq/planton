# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.
#
# The client Service is `<name>` and the headless sibling
# `<name>-headless` (fullnameOverride pins the chart's fullname to the
# resource name); the websocket endpoint and auth Secret name are empty
# when their features are off.

output "namespace" {
  description = "Kubernetes namespace the NATS servers run in"
  value       = local.namespace
}

output "service_name" {
  description = "Name of the client Service (port 4222; equals metadata.name)"
  value       = local.release_name
}

output "headless_service_name" {
  description = "Name of the headless Service — per-server DNS for clients that need direct server addressing"
  value       = "${local.release_name}-headless"
}

output "client_endpoint" {
  description = "In-cluster client endpoint — what NATS clients set as their server URL"
  value       = "nats://${local.release_name}.${local.namespace}.svc.cluster.local:4222"
}

output "websocket_endpoint" {
  description = "In-cluster WebSocket endpoint; empty when the websocket listener is off"
  value = try(var.spec.websocket.enabled, false) ? format(
    "ws://%s.%s.svc.cluster.local:%d",
    local.release_name, local.namespace,
    try(coalesce(var.spec.websocket.port), null) != null ? var.spec.websocket.port : 8080
  ) : ""
}

output "auth_secret_name" {
  description = "Name of the module-generated auth Secret (one key per declared username); empty when auth is not declared"
  value       = local.auth_secret_name
}

output "port_forward_command" {
  description = "kubectl one-liner for reaching the client port from a workstation when no exposure is composed"
  value       = "kubectl port-forward svc/${local.release_name} -n ${local.namespace} 4222:4222"
}
