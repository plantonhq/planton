# The optional ServiceMonitor scraping /metrics on the relay Service's
# `monitoring` port (Pulumi twin: service_monitor.go). It selects the
# chart's own Service labels. Requires the Prometheus Operator CRDs.
resource "kubectl_manifest" "service_monitor" {
  count = try(var.spec.metrics.service_monitor_enabled, false) ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "monitoring.coreos.com/v1"
    kind       = "ServiceMonitor"
    metadata = {
      name      = local.service_monitor_name
      namespace = local.namespace
      labels    = merge(try(var.spec.metrics.service_monitor_labels, {}), local.labels)
    }
    spec = {
      selector = {
        matchLabels = {
          "app.kubernetes.io/name"     = local.helm_chart_name
          "app.kubernetes.io/instance" = local.release_name
        }
      }
      endpoints = [
        {
          port     = "monitoring"
          path     = "/metrics"
          interval = "30s"
        }
      ]
    }
  })

  depends_on = [helm_release.relay_proxy]
}
