resource "kubernetes_namespace_v1" "grafana_namespace" {
  count = try(var.spec.create_namespace, false) ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

resource "helm_release" "grafana" {
  name       = var.metadata.name
  namespace  = local.namespace_name
  repository = "https://grafana.github.io/helm-charts"
  chart      = "grafana"
  version    = "8.7.0"

  # helm provider v3 replaced `set {}` blocks with list attributes; use the house
  # values=[yamlencode(...)] idiom. Mirrors the Pulumi module's helm values.
  values = [
    yamlencode({
      fullnameOverride = var.metadata.name
      resources = {
        requests = {
          cpu    = var.spec.container.resources.requests.cpu
          memory = var.spec.container.resources.requests.memory
        }
        limits = {
          cpu    = var.spec.container.resources.limits.cpu
          memory = var.spec.container.resources.limits.memory
        }
      }
      service = {
        type = "ClusterIP"
      }
      adminUser     = "admin"
      adminPassword = "admin"
      persistence = {
        enabled = false
      }
    })
  ]
}

resource "kubernetes_ingress_v1" "grafana_external" {
  count = local.ingress_is_enabled && local.ingress_external_hostname != "" ? 1 : 0

  metadata {
    name      = local.external_ingress_name
    namespace = local.namespace_name
    annotations = {
      "kubernetes.io/ingress.class" = "nginx"
    }
  }

  spec {
    rule {
      host = "grafana-${var.metadata.name}.${var.spec.ingress.dns_domain}"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = local.kube_service_name
              port {
                number = 80
              }
            }
          }
        }
      }
    }
  }

  depends_on = [helm_release.grafana]
}

resource "kubernetes_ingress_v1" "grafana_internal" {
  count = local.ingress_is_enabled && local.ingress_internal_hostname != "" ? 1 : 0

  metadata {
    name      = local.internal_ingress_name
    namespace = local.namespace_name
    annotations = {
      "kubernetes.io/ingress.class" = "nginx-internal"
    }
  }

  spec {
    rule {
      host = "grafana-${var.metadata.name}-internal.${var.spec.ingress.dns_domain}"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = local.kube_service_name
              port {
                number = 80
              }
            }
          }
        }
      }
    }
  }

  depends_on = [helm_release.grafana]
}

