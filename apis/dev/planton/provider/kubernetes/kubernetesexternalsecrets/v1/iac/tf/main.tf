###########################
# main.tf
###########################

# Conditional namespace creation
resource "kubernetes_namespace_v1" "external_secrets" {
  count = try(var.spec.create_namespace, false) ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.final_labels
  }
}

# Data source for existing namespace
data "kubernetes_namespace_v1" "existing" {
  count = try(var.spec.create_namespace, false) ? 0 : 1

  metadata {
    name = local.namespace
  }
}

# Create service account with cloud provider annotations
resource "kubernetes_service_account_v1" "external_secrets" {
  metadata {
    name        = local.service_account_name
    namespace   = local.namespace_name
    annotations = local.sa_annotations
    labels      = local.final_labels
  }
}

# Deploy External Secrets Operator via Helm
resource "helm_release" "external_secrets" {
  name       = local.release_name
  namespace  = local.namespace_name
  repository = local.helm_repo_url
  chart      = local.helm_chart_name
  version    = local.helm_chart_version

  atomic          = true
  cleanup_on_fail = true
  wait            = true
  wait_for_jobs   = true
  timeout         = 180

  values = [
    yamlencode({
      installCRDs = true

      serviceAccount = {
        create = false
        name   = local.service_account_name
      }

      rbac = {
        create = true
      }

      env = [
        {
          name  = "POLLER_INTERVAL_MILLISECONDS"
          value = tostring(local.poll_interval_ms)
        }
      ]

      resources = local.container_resources != null ? {
        requests = local.resource_requests != null ? {
          cpu    = try(local.resource_requests.cpu, null)
          memory = try(local.resource_requests.memory, null)
        } : null
        limits = local.resource_limits != null ? {
          cpu    = try(local.resource_limits.cpu, null)
          memory = try(local.resource_limits.memory, null)
        } : null
      } : {}
    })
  ]

  depends_on = [
    kubernetes_namespace_v1.external_secrets,
    data.kubernetes_namespace_v1.existing,
    kubernetes_service_account_v1.external_secrets
  ]
}
