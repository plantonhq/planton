##############################################
# outputs.tf
#
# Terraform outputs for KubernetesRookCephOperator
##############################################

output "namespace" {
  description = "Kubernetes namespace where the Rook Ceph Operator is deployed"
  value       = local.namespace
}

output "helm_release_name" {
  description = "Name of the Helm release for the Rook Ceph Operator"
  value       = helm_release.rook_ceph_operator.name
}

output "webhook_service" {
  description = "Webhook service name for the Rook Ceph Operator"
  value       = "${local.helm_release_name}-rook-ceph-operator"
}

output "port_forward_command" {
  description = "Command to setup port-forwarding to access the operator metrics"
  value       = "kubectl port-forward svc/rook-ceph-operator -n ${local.namespace} 9443:9443"
}
