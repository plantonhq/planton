# The optional ServiceMonitor scraping /metrics on the management port
# (Pulumi twin: service_monitor.go). Requires the Prometheus Operator CRDs.
resource "kubectl_manifest" "service_monitor" {
  count = try(var.spec.metrics.service_monitor_enabled, false) ? 1 : 0

  yaml_body = yamlencode({
    apiVersion = "monitoring.coreos.com/v1"
    kind       = "ServiceMonitor"
    metadata = {
      name      = "${local.name}-metrics"
      namespace = local.namespace
      labels    = merge(try(var.spec.metrics.service_monitor_labels, {}), local.labels)
    }
    spec = {
      selector  = { matchLabels = local.selector_labels }
      endpoints = [{ port = "management", path = "/metrics", interval = "30s" }]
    }
  })

  depends_on = [kubernetes_deployment_v1.flagd]
}
