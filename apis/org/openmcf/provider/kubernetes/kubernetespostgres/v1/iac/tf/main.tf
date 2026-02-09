# Conditionally create namespace for PostgreSQL if create_namespace is true
resource "kubernetes_namespace_v1" "postgres_namespace" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

# External LoadBalancer service (created only if ingress is enabled and a hostname is provided)
resource "kubernetes_service_v1" "external_lb" {
  count = local.ingress_is_enabled && local.ingress_external_hostname != null ? 1 : 0

  metadata {
    name      = local.external_lb_service_name
    namespace = local.namespace_name
    labels    = local.final_labels
    annotations = {
      "external-dns.alpha.kubernetes.io/hostname" = local.ingress_external_hostname
    }
  }

  spec {
    type = "LoadBalancer"

    port {
      name        = "postgres"
      port        = 5432
      target_port = 5432
      protocol    = "TCP"
    }

    selector = local.postgres_pod_selector_labels
  }

  depends_on = [
    kubernetes_namespace_v1.postgres_namespace,
    kubernetes_manifest.database,
  ]
}
