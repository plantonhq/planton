# The Service exposing flagd's four ports (Pulumi twin: deployment.go
# service).
resource "kubernetes_service_v1" "flagd" {
  metadata {
    name      = local.name
    namespace = local.namespace
    # The ServiceMonitor selects the Service by its selector labels.
    labels = merge(local.labels, local.selector_labels)
  }
  spec {
    type     = local.service_type
    selector = local.selector_labels
    port {
      name        = "evaluation"
      port        = local.port
      target_port = "evaluation"
    }
    port {
      name        = "management"
      port        = local.management_port
      target_port = "management"
    }
    port {
      name        = "sync"
      port        = local.sync_port
      target_port = "sync"
    }
    port {
      name        = "ofrep"
      port        = local.ofrep_port
      target_port = "ofrep"
    }
  }
  depends_on = [kubernetes_namespace_v1.flagd]
}
