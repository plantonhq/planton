output "cluster_id" {
  description = "OCID of the OKE cluster"
  value       = oci_containerengine_cluster.this.id
}

output "kubernetes_version" {
  description = "Kubernetes version running on the cluster control plane"
  value       = oci_containerengine_cluster.this.kubernetes_version
}

output "kubernetes_endpoint" {
  description = "Kubernetes API server endpoint URL"
  value       = try(oci_containerengine_cluster.this.endpoints[0].kubernetes, "")
}

output "private_endpoint" {
  description = "Private native networking Kubernetes API server endpoint URL"
  value       = try(oci_containerengine_cluster.this.endpoints[0].private_endpoint, "")
}

output "public_endpoint" {
  description = "Public native networking Kubernetes API server endpoint URL"
  value       = try(oci_containerengine_cluster.this.endpoints[0].public_endpoint, "")
}
