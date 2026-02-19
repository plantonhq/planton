output "node_pool_id" {
  description = "OCID of the OKE node pool"
  value       = oci_containerengine_node_pool.this.id
}

output "kubernetes_version" {
  description = "Kubernetes version running on the nodes in this pool"
  value       = oci_containerengine_node_pool.this.kubernetes_version
}
