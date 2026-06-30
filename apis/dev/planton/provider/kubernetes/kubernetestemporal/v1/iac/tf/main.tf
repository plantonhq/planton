# Create namespace for Temporal deployment (only if create_namespace is true)
resource "kubernetes_namespace_v1" "temporal_namespace" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

# Create secret for external database password (only when external database is configured with string_value)
# When using secret_ref, the user's existing secret is used directly
resource "kubernetes_secret_v1" "db_password" {
  count = local.has_external_database && !local.use_existing_db_secret && local.external_db_password_string != "" ? 1 : 0

  metadata {
    name      = local.database_secret_name
    namespace = local.namespace
    labels    = local.final_labels
  }

  data = {
    (local.database_secret_key) = local.external_db_password_string
  }

  type = "Opaque"
}

# Deploy Temporal using Helm chart.
# helm provider v3 replaced the `set {}`/`dynamic "set"` blocks with list attributes;
# the chart configuration is assembled in local.helm_values (locals.tf) using the house
# values=[yamlencode(...)] idiom, which mirrors the Pulumi module's nested values map.
resource "helm_release" "temporal" {
  name       = var.metadata.name
  repository = local.helm_chart_repository
  chart      = local.helm_chart_name
  version    = local.helm_chart_version
  namespace  = local.namespace

  # Wait for deployment to complete
  wait          = true
  wait_for_jobs = true
  timeout       = 600

  values = [
    yamlencode(local.helm_values)
  ]

  depends_on = [
    kubernetes_secret_v1.db_password
  ]
}

# Create LoadBalancer service for frontend gRPC ingress (when enabled)
resource "kubernetes_service_v1" "frontend_grpc_lb" {
  count = local.frontend_ingress_enabled && local.frontend_grpc_hostname != "" ? 1 : 0

  metadata {
    name      = "${var.metadata.name}-frontend-grpc-lb"
    namespace = local.namespace
    labels    = local.final_labels
    annotations = {
      "external-dns.alpha.kubernetes.io/hostname" = local.frontend_grpc_hostname
    }
  }

  spec {
    type = "LoadBalancer"

    port {
      name        = "grpc"
      port        = local.frontend_grpc_port
      target_port = local.frontend_grpc_port
      protocol    = "TCP"
    }

    selector = {
      "app.kubernetes.io/name"      = "temporal"
      "app.kubernetes.io/instance"  = var.metadata.name
      "app.kubernetes.io/component" = "frontend"
    }
  }

  depends_on = [
    helm_release.temporal
  ]
}
