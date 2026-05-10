output "namespace" {
  description = "Namespace where the controller is deployed"
  value       = var.spec.namespace
}

output "release_name" {
  description = "Name of the Helm release"
  value       = helm_release.controller.name
}

output "chart_version" {
  description = "Version of the deployed Helm chart"
  value       = helm_release.controller.version
}

output "deployment_name" {
  description = "Name of the controller deployment"
  value       = local.release_name
}

output "service_account_name" {
  description = "Name of the controller service account"
  value       = local.release_name
}

output "metrics_endpoint" {
  description = "Controller metrics endpoint (if metrics are enabled)"
  value = local.metrics_enabled ? format(
    "%s.%s.svc.cluster.local%s",
    local.release_name,
    var.spec.namespace,
    var.spec.metrics.controller_manager_addr
  ) : ""
}
