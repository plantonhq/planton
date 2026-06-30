##############################################
# outputs.tf
#
# Terraform outputs for KubernetesRookCephCluster
##############################################

output "namespace" {
  description = "Kubernetes namespace where the Ceph cluster is deployed"
  value       = local.namespace
}

output "helm_release_name" {
  description = "Name of the Helm release for the Rook Ceph Cluster"
  value       = helm_release.rook_ceph_cluster.name
}

output "ceph_cluster_name" {
  description = "Name of the CephCluster custom resource"
  value       = local.ceph_cluster_name
}

output "block_pool_names" {
  description = "Names of the created CephBlockPool resources"
  value       = local.block_pool_names
}

output "block_storage_class_names" {
  description = "Names of the created StorageClasses for block storage"
  value       = local.block_storage_class_names
}

output "filesystem_names" {
  description = "Names of the created CephFilesystem resources"
  value       = local.filesystem_names
}

output "filesystem_storage_class_names" {
  description = "Names of the created StorageClasses for CephFS"
  value       = local.filesystem_storage_class_names
}

output "object_store_names" {
  description = "Names of the created CephObjectStore resources"
  value       = local.object_store_names
}

output "object_storage_class_names" {
  description = "Names of the created StorageClasses for object bucket claims"
  value       = local.object_storage_class_names
}

output "dashboard_port_forward_command" {
  description = "Command to access the Ceph dashboard via port-forwarding"
  value       = "kubectl port-forward svc/rook-ceph-mgr-dashboard -n ${local.namespace} 7000:7000"
}

output "dashboard_url" {
  description = "URL to access the Ceph dashboard after port-forwarding"
  value       = "https://localhost:7000"
}

output "dashboard_password_command" {
  description = "Command to retrieve the Ceph dashboard admin password"
  value       = "kubectl -n ${local.namespace} get secret rook-ceph-dashboard-password -o jsonpath=\"{['data']['password']}\" | base64 -d"
}

output "toolbox_exec_command" {
  description = "Command to access the Ceph toolbox for debugging"
  value       = "kubectl -n ${local.namespace} exec -it deploy/rook-ceph-tools -- bash"
}
