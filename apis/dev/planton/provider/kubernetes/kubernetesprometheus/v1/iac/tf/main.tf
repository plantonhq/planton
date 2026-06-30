resource "kubernetes_namespace_v1" "prometheus_namespace" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

# Note: This module creates the namespace foundation for Prometheus deployment.
# The actual Prometheus deployment is expected to be managed via the kube-prometheus-stack Helm chart
# or the Prometheus Operator, which provides:
# - Prometheus server with persistent storage (if enabled)
# - Grafana for visualization
# - Alertmanager for alert routing
# - ServiceMonitors for auto-discovery of targets
# - PrometheusRules for recording and alerting rules
#
# The Helm chart deployment would reference this namespace and use the
# container specifications and ingress settings from the spec variable.

