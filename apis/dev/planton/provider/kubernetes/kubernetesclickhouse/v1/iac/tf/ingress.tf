resource "kubernetes_service_v1" "ingress_external_lb" {
  count = local.ingress_is_enabled ? 1 : 0

  metadata {
    # Use computed name to avoid conflicts when multiple instances share a namespace
    name      = local.external_lb_service_name
    namespace = local.namespace
    labels    = local.final_labels
    annotations = {
      "external-dns.alpha.kubernetes.io/hostname" = local.ingress_external_hostname
    }
  }

  spec {
    type = "LoadBalancer"

    port {
      name        = "http"
      port        = 8123
      protocol    = "TCP"
      target_port = "http"
    }

    port {
      name        = "tcp"
      port        = 9000
      protocol    = "TCP"
      target_port = "tcp"
    }

    selector = local.clickhouse_pod_selector_labels
  }

  depends_on = [
    kubernetes_manifest.clickhouse_installation
  ]
}
