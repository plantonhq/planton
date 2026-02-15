output "endpoint" {
  description = "The IP address of the cluster's Kubernetes API server"
  value       = google_container_cluster.cluster.endpoint
}

output "cluster_ca_certificate" {
  description = "Base64 encoded public certificate used to verify the cluster's certificate authority"
  value       = google_container_cluster.cluster.master_auth[0].cluster_ca_certificate
  sensitive   = true
}

output "workload_identity_pool" {
  description = "The Workload Identity Pool for this cluster (format: PROJECT_ID.svc.id.goog)"
  value       = var.spec.disable_workload_identity ? "" : "${var.spec.project_id.value}.svc.id.goog"
}

output "cluster_id" {
  description = "Fully qualified GKE cluster resource ID (projects/{project}/locations/{location}/clusters/{name})"
  value       = google_container_cluster.cluster.id
}

