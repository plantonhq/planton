locals {
  namespace          = var.spec.namespace
  helm_chart_name    = "cert-manager"
  helm_chart_repo    = "https://charts.jetstack.io"
  helm_chart_version = var.spec.helm_chart_version
  ksa_name           = var.metadata.name

  namespace_name = var.spec.create_namespace ? kubernetes_namespace.cert_manager[0].metadata[0].name : data.kubernetes_namespace.cert_manager[0].metadata[0].name

  sa_annotations = merge(
    var.spec.workload_identity != null && var.spec.workload_identity.gke != null ? {
      "iam.gke.io/gcp-service-account" = var.spec.workload_identity.gke.service_account_email
    } : {},
    var.spec.workload_identity != null && var.spec.workload_identity.eks != null ? {
      "eks.amazonaws.com/role-arn" = var.spec.workload_identity.eks.role_arn
    } : {},
    var.spec.workload_identity != null && var.spec.workload_identity.aks != null ? {
      "azure.workload.identity/client-id" = var.spec.workload_identity.aks.client_id
    } : {}
  )
}
