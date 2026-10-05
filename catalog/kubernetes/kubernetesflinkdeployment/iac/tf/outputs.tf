# Outputs — identical names and derivations in the Pulumi module
# (KubernetesFlinkDeploymentOutputs).

output "namespace" {
  description = "Namespace the Flink cluster runs in"
  value       = local.namespace
}

output "rest_service" {
  description = "The JobManager REST Service name (`<name>-rest`) — the Flink REST API and web UI (port 8081); where session-mode jobs submit and where job status reads from"
  value       = local.rest_service
}

output "rest_endpoint" {
  description = "In-cluster REST endpoint (`<rest_service>.<namespace>.svc.cluster.local:8081`)"
  value       = local.rest_endpoint
}

output "port_forward_command" {
  description = "kubectl port-forward one-liner for reaching the Flink UI from a workstation"
  value       = local.port_forward_command
}
