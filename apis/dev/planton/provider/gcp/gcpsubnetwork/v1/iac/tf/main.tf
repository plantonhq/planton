# Enable required GCP APIs for subnetwork operations
resource "google_project_service" "compute_api" {
  project = var.spec.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# Create the GCP Subnetwork in custom mode VPC
resource "google_compute_subnetwork" "main" {
  name          = var.spec.subnetwork_name
  project       = var.spec.project_id
  region        = var.spec.region
  network       = var.spec.vpc_self_link
  ip_cidr_range = var.spec.ip_cidr_range

  # Enable Private Google Access for internal-only instances
  private_ip_google_access = local.private_ip_google_access

  # Define secondary IP ranges for alias IPs (e.g., GKE pods and services)
  dynamic "secondary_ip_range" {
    for_each = local.secondary_ip_ranges
    content {
      range_name    = secondary_ip_range.value.range_name
      ip_cidr_range = secondary_ip_range.value.ip_cidr_range
    }
  }

  # Depend on API enablement to ensure compute API is active
  depends_on = [google_project_service.compute_api]
}

