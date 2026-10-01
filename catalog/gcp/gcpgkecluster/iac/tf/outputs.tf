output "endpoint" {
  description = "IP address of the cluster's Kubernetes API server (private endpoint on private-only control planes)"
  value       = google_container_cluster.this.endpoint
}

output "cluster_ca_certificate" {
  description = "Base64-encoded CA certificate clients use to validate the API server's TLS certificate (public trust material)"
  value       = google_container_cluster.this.master_auth[0].cluster_ca_certificate
}

output "workload_identity_pool" {
  description = "Workload Identity pool for this cluster (PROJECT_ID.svc.id.goog); empty when Workload Identity is disabled on a Standard cluster"
  value       = (var.spec.workload_identity_enabled || var.spec.enable_autopilot) ? local.workload_pool : ""
}

output "cluster_id" {
  description = "Fully qualified GKE cluster resource ID (projects/{project}/locations/{location}/clusters/{name})"
  value       = google_container_cluster.this.id
}

output "name" {
  description = "Name of the cluster as created in GCP — the handle node pools and gcloud commands use"
  value       = google_container_cluster.this.name
}

output "location" {
  description = "Cluster location exactly as provided in the spec (region for regional clusters, zone for zonal)"
  value       = var.spec.location
}

output "self_link" {
  description = "Server-defined URL of the cluster resource"
  value       = google_container_cluster.this.self_link
}

output "master_version" {
  description = "Kubernetes version currently running on the control plane"
  value       = google_container_cluster.this.master_version
}

# Google creates the membership when the cluster joins a fleet through
# fleet_project; empty when the cluster joins no fleet.
output "fleet_membership" {
  description = "Full name of the fleet membership Google created for the cluster (empty when it joins no fleet)"
  value       = try(google_container_cluster.this.fleet[0].membership, "")
}
