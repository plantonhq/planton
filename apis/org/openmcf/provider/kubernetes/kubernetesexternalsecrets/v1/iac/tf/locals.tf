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
    "resource_kind" = "kubernetes_external_secrets"
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

  # Namespace with default
  namespace = coalesce(var.spec.namespace, "kubernetes-external-secrets")

  # Namespace reference - either created or existing
  namespace_name = try(var.spec.create_namespace, false) ? (
    length(kubernetes_namespace_v1.external_secrets) > 0 ? 
      kubernetes_namespace_v1.external_secrets[0].metadata[0].name : local.namespace
  ) : (
    length(data.kubernetes_namespace_v1.existing) > 0 ? 
      data.kubernetes_namespace_v1.existing[0].metadata[0].name : local.namespace
  )

  # Computed resource names to avoid conflicts when multiple instances share a namespace
  # Format: {metadata.name}-{purpose}
  # Users can prefix metadata.name with component type if needed (e.g., "eso-my-instance")
  service_account_name = var.metadata.name
  release_name         = var.metadata.name

  # Helm chart configuration
  helm_repo_url        = "https://charts.external-secrets.io"
  helm_chart_name      = "external-secrets"
  helm_chart_version   = try(var.spec.helm_chart_version, "0.9.20")

  # Default polling interval for secrets (in milliseconds)
  poll_interval_ms = try(var.spec.poll_interval_seconds, 10) * 1000

  # Determine provider type
  is_gke = try(var.spec.gke != null, false)
  is_eks = try(var.spec.eks != null, false)
  is_aks = try(var.spec.aks != null, false)

  # GKE configuration
  gke_gsa_email = local.is_gke ? try(var.spec.gke.gsa_email, "") : ""

  # EKS configuration
  eks_irsa_role_arn = local.is_eks ? try(var.spec.eks.irsa_role_arn_override, "") : ""

  # AKS configuration
  aks_managed_identity_client_id = local.is_aks ? try(var.spec.aks.managed_identity_client_id, "") : ""

  # Service account annotations based on cloud provider
  sa_annotations = (
    local.is_gke && local.gke_gsa_email != "" ? {
      "iam.gke.io/gcp-service-account" = local.gke_gsa_email
    } :
    local.is_eks && local.eks_irsa_role_arn != "" ? {
      "eks.amazonaws.com/role-arn" = local.eks_irsa_role_arn
    } :
    local.is_aks && local.aks_managed_identity_client_id != "" ? {
      "azure.workload.identity/client-id" = local.aks_managed_identity_client_id
    } : {}
  )

  # Container resources
  container_resources = try(var.spec.container.resources, null)
  resource_requests = try(local.container_resources.requests, null)
  resource_limits   = try(local.container_resources.limits, null)
}
