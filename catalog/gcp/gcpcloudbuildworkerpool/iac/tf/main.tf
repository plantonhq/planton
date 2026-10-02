# The Cloud Build API. disable_on_destroy is false: tearing down one pool
# must never disable the API for every other build in the project.
resource "google_project_service" "cloudbuild_api" {
  project = local.project_id
  service = "cloudbuild.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The project NUMBER the peered network's path needs, read only when the
# path carries a project ID; a numeric path (or no peering) performs no
# read.
data "google_project" "peered_network" {
  count      = local.peered_network_needs_number ? 1 : 0
  project_id = local.peered_network_project
}

# The private worker pool. Location, ID, and both network arms are
# immutable; the display name, annotations, and worker shape update in
# place.
resource "google_cloudbuild_worker_pool" "this" {
  project         = local.project_id
  location        = var.spec.location
  name            = local.worker_pool_id
  display_name    = var.spec.display_name != "" ? var.spec.display_name : null
  annotations     = length(var.spec.annotations) > 0 ? var.spec.annotations : null
  deletion_policy = local.deletion_policy

  dynamic "network_config" {
    for_each = var.spec.network_config != null ? [var.spec.network_config] : []
    content {
      peered_network          = local.peered_network
      peered_network_ip_range = network_config.value.peered_network_ip_range != "" ? network_config.value.peered_network_ip_range : null
    }
  }

  dynamic "private_service_connect" {
    for_each = var.spec.private_service_connect != null ? [var.spec.private_service_connect] : []
    content {
      network_attachment = private_service_connect.value.network_attachment
      route_all_traffic  = private_service_connect.value.route_all_traffic
    }
  }

  dynamic "worker_config" {
    for_each = var.spec.worker_config != null ? [var.spec.worker_config] : []
    content {
      machine_type                 = worker_config.value.machine_type != "" ? worker_config.value.machine_type : null
      disk_size_gb                 = worker_config.value.disk_size_gb != 0 ? worker_config.value.disk_size_gb : null
      no_external_ip               = worker_config.value.no_external_ip
      enable_nested_virtualization = worker_config.value.enable_nested_virtualization
    }
  }

  depends_on = [google_project_service.cloudbuild_api]
}
