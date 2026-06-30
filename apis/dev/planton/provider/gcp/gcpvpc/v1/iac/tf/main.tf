#############################################
# Enable Compute API
#############################################

resource "google_project_service" "compute_api" {
  project                    = var.spec.project_id.value
  service                    = "compute.googleapis.com"
  disable_dependent_services = true
}

#############################################
# GCP VPC Network
#############################################

resource "google_compute_network" "vpc" {
  name                    = var.spec.network_name
  project                 = var.spec.project_id.value
  auto_create_subnetworks = var.spec.auto_create_subnetworks
  routing_mode            = local.routing_mode

  # Ensure Compute API is enabled before creating VPC
  depends_on = [google_project_service.compute_api]
}

#############################################
# Private Services Access (Optional)
# Enables private IP connectivity to Google
# managed services (Cloud SQL, Memorystore, etc.)
# PREREQUISITE: servicenetworking.googleapis.com
# must be enabled on the project via GcpProject.
#############################################

# Allocate IP range for private services
resource "google_compute_global_address" "private_services_range" {
  count         = local.enable_private_services ? 1 : 0
  name          = "${var.spec.network_name}-private-svc"
  project       = var.spec.project_id.value
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = local.private_services_prefix_length
  network       = google_compute_network.vpc.id
}

# Create private service connection (VPC peering with Google's service network)
resource "google_service_networking_connection" "private_services" {
  count                   = local.enable_private_services ? 1 : 0
  network                 = google_compute_network.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_services_range[0].name]
}

