resource "helm_release" "redis" {
  name       = local.resource_id
  repository = "https://charts.bitnami.com/bitnami"
  chart      = "redis"
  version    = "17.10.1"
  namespace  = local.namespace

  depends_on = [kubernetes_namespace.redis_namespace]

  # Convert your entire map into a single YAML string
  values = [
    yamlencode({
      fullnameOverride = var.metadata.name
      architecture     = "standalone"

      image = {
        registry   = local.redis_image_registry
        repository = local.redis_image_repository
        tag        = local.redis_image_tag
      }

      master = {
        podLabels = local.final_labels
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
        persistence = {
          enabled = var.spec.container.persistence_enabled
          size    = var.spec.container.disk_size
        }
      }

      replica = {
        podLabels    = local.final_labels
        replicaCount = var.spec.container.replicas
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
        persistence = {
          enabled = var.spec.container.persistence_enabled
          size    = var.spec.container.disk_size
        }
      }

      auth = {
        existingSecret            = local.password_secret_name
        existingSecretPasswordKey = "password"
      }
    })
  ]
}
