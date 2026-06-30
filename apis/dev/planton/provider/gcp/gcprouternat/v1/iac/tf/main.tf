# =============================================================================
# Cloud Router
# =============================================================================

resource "google_compute_router" "router" {
  name    = local.router_name
  region  = var.spec.region
  network = var.spec.vpc_self_link
  project = var.spec.project_id
}

# =============================================================================
# Static External IP Addresses (only if manual allocation is required)
# =============================================================================

resource "google_compute_address" "nat_ips" {
  count = length(var.spec.nat_ip_names)

  name         = var.spec.nat_ip_names[count.index]
  region       = var.spec.region
  address_type = "EXTERNAL"
  labels       = local.gcp_labels
  project      = var.spec.project_id
}

# =============================================================================
# Cloud NAT
# =============================================================================

resource "google_compute_router_nat" "nat" {
  name    = local.nat_name
  router  = google_compute_router.router.name
  region  = var.spec.region
  project = var.spec.project_id

  # NAT IP allocation strategy (AUTO_ONLY or MANUAL_ONLY)
  nat_ip_allocate_option = local.nat_ip_allocate_option

  # Static IP addresses (only used if MANUAL_ONLY)
  nat_ips = local.nat_ip_allocate_option == "MANUAL_ONLY" ? google_compute_address.nat_ips[*].self_link : []

  # Subnet coverage (ALL_SUBNETWORKS_ALL_IP_RANGES or LIST_OF_SUBNETWORKS)
  source_subnetwork_ip_ranges_to_nat = local.source_subnetwork_ip_ranges_to_nat

  # Specific subnetworks (only used if LIST_OF_SUBNETWORKS)
  dynamic "subnetwork" {
    for_each = local.subnetworks
    content {
      name                    = subnetwork.value.name
      source_ip_ranges_to_nat = subnetwork.value.source_ip_ranges_to_nat
    }
  }

  # Logging configuration
  log_config {
    enable = local.enable_logging
    filter = local.log_filter
  }

  # Production defaults for NAT behavior
  # These follow GCP best practices and match the Pulumi implementation
  min_ports_per_vm                    = 64     # Default ports per VM (sufficient for most workloads)
  enable_endpoint_independent_mapping = true   # Allows port reuse across different destinations
  enable_dynamic_port_allocation      = false  # Keep static allocation for predictable behavior

  # TCP timeouts (GCP defaults)
  tcp_established_idle_timeout_sec = 1200 # 20 minutes
  tcp_transitory_idle_timeout_sec  = 30   # 30 seconds

  # UDP timeout (GCP default)
  udp_idle_timeout_sec = 30 # 30 seconds

  # ICMP timeout (GCP default)
  icmp_idle_timeout_sec = 30 # 30 seconds
}

