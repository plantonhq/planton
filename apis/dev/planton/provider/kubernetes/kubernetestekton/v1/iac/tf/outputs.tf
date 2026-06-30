##############################################
# outputs.tf
#
# Terraform outputs for KubernetesTekton deployment.
##############################################

output "namespace" {
  description = "The namespace where Tekton is installed"
  value       = local.namespace
}

output "pipeline_version" {
  description = "The version of Tekton Pipelines deployed"
  value       = local.pipeline_version
}

output "dashboard_version" {
  description = "The version of Tekton Dashboard deployed (empty if disabled)"
  value       = local.dashboard_enabled ? local.dashboard_version : ""
}

output "dashboard_internal_endpoint" {
  description = "Internal cluster endpoint for the Tekton Dashboard"
  value       = local.dashboard_enabled ? local.dashboard_internal_endpoint : ""
}

output "dashboard_external_hostname" {
  description = "External hostname for the Tekton Dashboard (if ingress enabled)"
  value       = local.ingress_enabled ? local.ingress_hostname : ""
}

output "port_forward_dashboard_command" {
  description = "kubectl port-forward command to access the dashboard locally"
  value       = local.dashboard_enabled ? local.port_forward_dashboard_command : ""
}

output "cloud_events_sink_url" {
  description = "The CloudEvents sink URL configured for pipeline notifications"
  value       = local.cloud_events_sink_url
}
