output "namespace" {
  description = "The Kubernetes namespace where Tekton components are deployed"
  value       = local.components_namespace
}

output "operator_namespace" {
  description = "The Kubernetes namespace where Tekton Operator is deployed"
  value       = local.operator_namespace
}

output "tekton_config_name" {
  description = "The name of the TektonConfig resource"
  value       = local.tekton_config_name
}

output "tekton_profile" {
  description = "The Tekton profile used (all, basic, or lite)"
  value       = local.tekton_profile
}

output "pipelines_controller_service" {
  description = "The Tekton Pipelines controller service name"
  value       = var.spec.components.pipelines ? "tekton-pipelines-controller" : null
}

output "triggers_controller_service" {
  description = "The Tekton Triggers controller service name"
  value       = var.spec.components.triggers ? "tekton-triggers-controller" : null
}

output "dashboard_service" {
  description = "The Tekton Dashboard service name"
  value       = var.spec.components.dashboard ? "tekton-dashboard" : null
}

output "dashboard_port_forward_command" {
  description = "Command to port-forward to the Tekton Dashboard"
  value       = var.spec.components.dashboard ? "kubectl port-forward svc/tekton-dashboard -n ${local.components_namespace} 9097:9097" : null
}
