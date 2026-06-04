# Terraform module for Percona Operator for MongoDB

resource "kubernetes_namespace" "percona_operator" {
  count = var.spec.create_namespace ? 1 : 0

  metadata {
    name   = local.namespace
    labels = local.labels
  }
}

resource "helm_release" "percona_operator" {
  # Use computed release name from metadata.name to avoid conflicts when multiple instances share a namespace
  name       = local.helm_release_name
  repository = local.helm_chart_repo
  chart      = local.helm_chart_name
  version    = local.helm_chart_version
  namespace  = var.spec.create_namespace ? kubernetes_namespace.percona_operator[0].metadata[0].name : local.namespace

  # helm provider v3 replaced `set {}` blocks with list attributes; use the house
  # values=[yamlencode(...)] idiom. Mirrors the Pulumi module's helm values.
  values = [
    yamlencode({
      watchAllNamespaces = true
      resources = {
        limits = {
          cpu    = var.spec.container.resources.limits.cpu
          memory = var.spec.container.resources.limits.memory
        }
        requests = {
          cpu    = var.spec.container.resources.requests.cpu
          memory = var.spec.container.resources.requests.memory
        }
      }
    })
  ]

  timeout         = 300
  atomic          = true
  cleanup_on_fail = true
  wait            = true
}

