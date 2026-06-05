###########################
# locals.tf
###########################

locals {
  # Derive a stable resource ID (fall back to name if ID is missing/empty).
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base labels
  base_labels = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "kubernetes_external_dns"
  }

  # Organization label only if non-empty
  org_label = (
    var.metadata.org != null && var.metadata.org != ""
    ) ? {
    "organization" = var.metadata.org
  } : {}

  # Environment label only if env is set
  env_label = (
    var.metadata.env != null && var.metadata.env != ""
    ) ? {
    "environment" = var.metadata.env
  } : {}

  final_labels = merge(local.base_labels, local.org_label, local.env_label)

  # Namespace where ExternalDNS is installed. Required field; the proto->tfvars
  # converter flattens the StringValueOrRef to a plain string before apply.
  namespace = var.spec.namespace

  # Namespace reference - either created or existing
  namespace_name = try(var.spec.create_namespace, false) ? (
    length(kubernetes_namespace_v1.external_dns) > 0 ?
    kubernetes_namespace_v1.external_dns[0].metadata[0].name : local.namespace
    ) : (
    length(data.kubernetes_namespace_v1.existing) > 0 ?
    data.kubernetes_namespace_v1.existing[0].metadata[0].name : local.namespace
  )

  # Release name matches the resource name for multi-instance support
  release_name = var.metadata.name
  ksa_name     = local.release_name

  # ExternalDNS and Helm chart versions with defaults
  external_dns_version = try(var.spec.external_dns_version, "v0.19.0")
  helm_chart_version   = try(var.spec.helm_chart_version, "1.19.0")

  # Helm repository and chart
  helm_repo_url   = "https://kubernetes-sigs.github.io/external-dns/"
  helm_chart_name = "external-dns"

  # Determine provider type
  is_gke        = try(var.spec.gke != null, false)
  is_eks        = try(var.spec.eks != null, false)
  is_aks        = try(var.spec.aks != null, false)
  is_cloudflare = try(var.spec.cloudflare != null, false)

  # Provider-specific values
  provider_type = (
    local.is_gke ? "google" :
    local.is_eks ? "aws" :
    local.is_aks ? "azure" :
    local.is_cloudflare ? "cloudflare" : "unknown"
  )

  # GKE configuration
  gke_project_id  = local.is_gke ? try(var.spec.gke.project_id, "") : ""
  gke_dns_zone_id = local.is_gke ? try(var.spec.gke.dns_zone_id, "") : ""
  gke_gsa_email   = local.is_gke ? "${local.ksa_name}@${local.gke_project_id}.iam.gserviceaccount.com" : ""

  # EKS configuration
  eks_route53_zone_id = local.is_eks ? try(var.spec.eks.route53_zone_id, "") : ""
  eks_irsa_role_arn   = local.is_eks ? try(var.spec.eks.irsa_role_arn_override, "") : ""

  # AKS configuration
  aks_dns_zone_id                = local.is_aks ? try(var.spec.aks.dns_zone_id, "") : ""
  aks_managed_identity_client_id = local.is_aks ? try(var.spec.aks.managed_identity_client_id, "") : ""

  # Cloudflare configuration
  cf_api_token   = local.is_cloudflare ? try(var.spec.cloudflare.api_token, "") : ""
  cf_dns_zone_id = local.is_cloudflare ? try(var.spec.cloudflare.dns_zone_id, "") : ""
  cf_is_proxied  = local.is_cloudflare ? try(var.spec.cloudflare.is_proxied, false) : false
  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  cloudflare_api_token_secret_name = local.is_cloudflare ? "${local.release_name}-cloudflare-api-token" : ""

  # Service account annotations
  sa_annotations = (
    local.is_gke ? {
      "iam.gke.io/gcp-service-account" = local.gke_gsa_email
    } :
    local.is_eks && local.eks_irsa_role_arn != "" ? {
      "eks.amazonaws.com/role-arn" = local.eks_irsa_role_arn
    } :
    local.is_aks && local.aks_managed_identity_client_id != "" ? {
      "azure.workload.identity/client-id" = local.aks_managed_identity_client_id
    } : {}
  )

  # Zone-id filter scopes ExternalDNS to the configured DNS zone. GKE/EKS/AKS pass
  # it via the chart's zoneIdFilters value; Cloudflare uses an extraArg instead.
  zone_id_filters = (
    local.is_gke ? [local.gke_dns_zone_id] :
    local.is_eks ? [local.eks_route53_zone_id] :
    local.is_aks && local.aks_dns_zone_id != "" ? [local.aks_dns_zone_id] :
    []
  )

  # ExternalDNS Helm chart values. This nested map mirrors the Pulumi module's
  # `Values` map (iac/pulumi/module/main.go) so both engines render the same chart
  # config; it is passed to helm_release via `values = [yamlencode(local.helm_values)]`.
  # Provider-specific fragments are appended as single-element lists so the conditional
  # branches share a list type (bare-object branches fail tofu's type unification).
  helm_values = merge(concat(
    [
      {
        serviceAccount = {
          create = false
          name   = local.ksa_name
        }
        provider = local.provider_type
        image = {
          tag = local.external_dns_version
        }
      },
    ],
    length(local.zone_id_filters) > 0 ? [{ zoneIdFilters = local.zone_id_filters }] : [],
    local.is_gke ? [{ google = { project = local.gke_project_id } }] : [],
    local.is_cloudflare ? [{
      sources = ["service", "ingress", "gateway-httproute", "istio-gateway"]
      env = [
        {
          name = "CF_API_TOKEN"
          valueFrom = {
            secretKeyRef = {
              name = local.cloudflare_api_token_secret_name
              key  = "apiKey"
            }
          }
        }
      ]
      extraArgs = concat(
        [
          "--cloudflare-dns-records-per-page=5000",
          "--zone-id-filter=${local.cf_dns_zone_id}",
        ],
        local.cf_is_proxied ? ["--cloudflare-proxied"] : [],
      )
    }] : [],
  )...)
}

