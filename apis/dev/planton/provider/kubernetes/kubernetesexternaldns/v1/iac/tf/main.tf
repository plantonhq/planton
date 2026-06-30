###########################
# main.tf
###########################

# Conditional namespace creation
resource "kubernetes_namespace_v1" "external_dns" {
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
resource "kubernetes_service_account_v1" "external_dns" {
  metadata {
    name        = local.ksa_name
    namespace   = local.namespace_name
    annotations = local.sa_annotations
    labels      = local.final_labels
  }
}

# Create secret for Cloudflare API token (only if using Cloudflare)
resource "kubernetes_secret_v1" "cloudflare_api_token" {
  count = local.is_cloudflare ? 1 : 0

  metadata {
    name      = local.cloudflare_api_token_secret_name
    namespace = local.namespace_name
    labels    = local.final_labels
  }

  data = {
    apiKey = local.cf_api_token
  }

  type = "Opaque"
}

# Deploy ExternalDNS via Helm. Chart values are assembled in local.helm_values
# (locals.tf), which mirrors the Pulumi module's Values map for cross-engine parity.
resource "helm_release" "external_dns" {
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
    yamlencode(local.helm_values)
  ]

  depends_on = [
    kubernetes_namespace_v1.external_dns,
    data.kubernetes_namespace_v1.existing,
    kubernetes_service_account_v1.external_dns
  ]
}
