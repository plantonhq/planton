# Conditional namespace creation for Argo CD
resource "kubernetes_namespace" "argocd_namespace" {
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

# Deploy Argo CD using the official Helm chart
resource "helm_release" "argocd" {
  name       = local.resource_id
  repository = local.argocd_chart_repo
  chart      = local.argocd_chart_name
  version    = local.argocd_chart_version
  namespace  = local.namespace_name

  # Wait for resources to be ready
  wait          = true
  wait_for_jobs = true
  timeout       = 600 # 10 minutes for initial deployment

  # Enable atomic rollback on failure
  atomic          = true
  cleanup_on_fail = true

  # Core configuration values
  values = [
    yamlencode({
      # Server configuration
      server = {
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
        # Disable admin user in production (enable SSO instead)
        # Admin can still be used for initial setup
        extraArgs = [
          "--insecure" # Allow HTTP access (use ingress for TLS termination)
        ]
      }

      # Application controller resources
      controller = {
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
      }

      # Repo server resources
      repoServer = {
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
      }

      # Redis configuration
      redis = {
        resources = {
          requests = {
            cpu    = "50m"
            memory = "64Mi"
          }
          limits = {
            cpu    = "100m"
            memory = "128Mi"
          }
        }
      }

      # Global configuration
      global = {
        image = {
          # Use official ArgoCD images
          repository = "quay.io/argoproj/argocd"
        }
      }
    })
  ]

  depends_on = [
    kubernetes_namespace.argocd_namespace,
    data.kubernetes_namespace.existing
  ]
}

