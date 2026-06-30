###########################
# outputs.tf
###########################

output "namespace" {
  description = "The namespace where External Secrets Operator is deployed"
  value       = local.namespace_name
}

output "release_name" {
  description = "The Helm release name"
  value       = local.release_name
}

output "operator_service_account" {
  description = "The service account name for the External Secrets Operator"
  value       = kubernetes_service_account_v1.external_secrets.metadata[0].name
}

output "identity" {
  description = "The cloud identity (GSA email, IRSA role ARN, or managed identity client ID)"
  value = (
    local.is_gke ? local.gke_gsa_email :
    local.is_eks ? local.eks_irsa_role_arn :
    local.is_aks ? local.aks_managed_identity_client_id : null
  )
}
