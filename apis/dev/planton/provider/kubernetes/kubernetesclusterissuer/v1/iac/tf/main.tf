# Create Kubernetes Secret for Cloudflare API token (only when using Cloudflare provider)
resource "kubernetes_secret" "cloudflare" {
  count = local.is_cloudflare ? 1 : 0

  metadata {
    name      = local.cloudflare_secret_name
    namespace = local.cert_manager_namespace
  }

  data = {
    "api-token" = var.spec.cloudflare.api_token
  }
}

# Create the ClusterIssuer custom resource
resource "kubernetes_manifest" "cluster_issuer" {
  manifest = {
    apiVersion = "cert-manager.io/v1"
    kind       = "ClusterIssuer"

    metadata = {
      name = local.dns_domain
    }

    spec = {
      acme = {
        email  = var.spec.acme.email
        server = var.spec.acme.server

        privateKeySecretRef = {
          name = local.acme_account_key_secret_name
        }

        solvers = [local.solver]
      }
    }
  }

  depends_on = [kubernetes_secret.cloudflare]
}
