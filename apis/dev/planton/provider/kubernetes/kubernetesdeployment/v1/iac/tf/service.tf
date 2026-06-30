resource "kubernetes_service" "this" {
  metadata {
    name      = local.kube_service_name
    namespace = local.namespace
    labels    = local.final_labels
  }

  spec {
    type     = "ClusterIP"
    selector = local.selector_labels

    dynamic "port" {
      for_each = try(var.spec.container.app.ports, [])
      content {
        name        = port.value.name
        protocol    = port.value.network_protocol
        port        = port.value.service_port
        target_port = port.value.container_port
        # The appProtocol is only recognized in newer
        # Kubernetes versions. If you need it, you can
        # set it via an annotation or the official field:
        app_protocol = port.value.app_protocol
      }
    }
  }

  depends_on = [
    kubernetes_namespace.this,
    kubernetes_deployment.this
  ]
}
