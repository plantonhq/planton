# Outputs — identical names and derivations in the Pulumi module's
# outputs.go / main.go exports.
#
# Every child service name derives from the pinned fullname (= the
# resource name); the Prometheus endpoint is the URL Grafana datasources
# and remote readers compose against; the Grafana admin Secret follows the
# credential arm (subchart-generated `<name>-grafana` or the referenced
# existing Secret).

output "namespace" {
  description = "Kubernetes namespace the stack runs in"
  value       = local.namespace
}

output "release_name" {
  description = "Helm release name (= metadata.name)"
  value       = local.release_name
}

output "prometheus_service" {
  description = "Name of the Prometheus Service (`<name>-prometheus`, port 9090)"
  value       = local.prometheus_service
}

output "prometheus_endpoint" {
  description = "In-cluster Prometheus endpoint — the URL datasources and remote readers use"
  value       = "http://${local.prometheus_service}.${local.namespace}.svc.cluster.local:9090"
}

output "alertmanager_service" {
  description = "Name of the Alertmanager Service (`<name>-alertmanager`, port 9093); empty when disabled"
  value       = local.alertmanager_service
}

output "alertmanager_endpoint" {
  description = "In-cluster Alertmanager endpoint; empty when disabled"
  value       = local.alertmanager_enabled ? "http://${local.alertmanager_service}.${local.namespace}.svc.cluster.local:9093" : ""
}

output "grafana_service" {
  description = "Name of the bundled Grafana Service (`<name>-grafana`, port 80); empty when disabled"
  value       = local.grafana_service
}

output "grafana_endpoint" {
  description = "In-cluster Grafana endpoint; empty when disabled"
  value       = local.grafana_enabled ? "http://${local.grafana_service}.${local.namespace}.svc.cluster.local" : ""
}

output "grafana_admin_secret_name" {
  description = "Secret holding the bundled Grafana's admin credentials (keys admin-user / admin-password for the generated arm); empty when disabled"
  value       = local.grafana_admin_secret_name
}

output "prometheus_port_forward_command" {
  description = "kubectl one-liner for reaching the Prometheus UI from a workstation"
  value       = "kubectl port-forward svc/${local.prometheus_service} -n ${local.namespace} 9090:9090"
}

output "grafana_port_forward_command" {
  description = "kubectl one-liner for reaching the bundled Grafana from a workstation; empty when disabled"
  value       = local.grafana_enabled ? "kubectl port-forward svc/${local.grafana_service} -n ${local.namespace} 3000:80" : ""
}
