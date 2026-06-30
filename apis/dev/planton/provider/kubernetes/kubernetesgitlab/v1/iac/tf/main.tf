###########################
# main.tf
###########################

# Conditional namespace creation for GitLab
resource "kubernetes_namespace" "gitlab" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

# Data source for existing namespace
data "kubernetes_namespace" "existing" {
  count = var.spec.create_namespace ? 0 : 1

  metadata {
    name = local.namespace
  }
}

# Note: GitLab typically requires a complex Helm chart deployment.
# This is a placeholder for the actual GitLab deployment.
# In production, you would use the official GitLab Helm chart:
# https://docs.gitlab.com/charts/

# For actual deployment, integrate with Helm provider:
# resource "helm_release" "gitlab" {
#   name       = var.metadata.name
#   namespace  = kubernetes_namespace.gitlab.metadata[0].name
#   repository = "https://charts.gitlab.io/"
#   chart      = "gitlab"
#   version    = "7.0.0"
#
#   values = [
#     templatefile("${path.module}/values.yaml.tpl", {
#       resources = var.spec.container.resources
#       ingress_enabled = local.ingress_is_enabled
#       hostname = local.ingress_external_hostname
#     })
#   ]
# }

# Placeholder service (would be created by Helm chart in production)
resource "kubernetes_service" "gitlab" {
  metadata {
    name      = local.gitlab_service_name
    namespace = local.namespace_name
    labels    = local.final_labels
  }

  spec {
    selector = merge(local.final_labels, {
      "app" = "gitlab"
    })

    port {
      name        = "http"
      port        = local.gitlab_port
      target_port = 8080
      protocol    = "TCP"
    }

    type = "ClusterIP"
  }
}

# Ingress for GitLab (if enabled)
resource "kubernetes_ingress_v1" "gitlab" {
  count = local.ingress_is_enabled ? 1 : 0

  metadata {
    name      = local.ingress_name
    namespace = local.namespace_name
    labels    = local.final_labels

    annotations = {
      "cert-manager.io/cluster-issuer" = local.ingress_cert_cluster_issuer_name
    }
  }

  spec {
    ingress_class_name = local.gateway_ingress_class_name

    tls {
      hosts = [local.ingress_external_hostname]
      secret_name = local.ingress_cert_secret_name
    }

    rule {
      host = local.ingress_external_hostname

      http {
        path {
          path      = "/"
          path_type = "Prefix"

          backend {
            service {
              name = kubernetes_service.gitlab.metadata[0].name
              port {
                number = local.gitlab_port
              }
            }
          }
        }
      }
    }
  }

  depends_on = [
    kubernetes_namespace.gitlab,
    data.kubernetes_namespace.existing
  ]
}
