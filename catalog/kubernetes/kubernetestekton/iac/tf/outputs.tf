# Outputs — identical names and derivations in the Pulumi module
# (KubernetesTektonOutputs).

output "namespace" {
  description = "Namespace the Tekton components run in (the TektonConfig targetNamespace)"
  value       = local.target_namespace
}

output "profile" {
  description = "The installed profile (lite, basic or all)"
  value       = local.profile
}

output "dashboard_service" {
  description = "Name of the dashboard Service in the target namespace — the backend handle exposure kinds reference; empty unless profile is all"
  value       = local.dashboard_service
}

output "dashboard_kube_endpoint" {
  description = "In-cluster endpoint of the dashboard; empty unless profile is all"
  value       = local.dashboard_kube_endpoint
}

output "port_forward_command" {
  description = "Command to port-forward the dashboard to a workstation; empty unless profile is all"
  value       = local.port_forward_command
}
