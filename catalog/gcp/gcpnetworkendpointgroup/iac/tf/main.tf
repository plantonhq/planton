# GCP models zonal and global network endpoint groups as two API
# collections with two endpoint mechanisms. spec.zone selects which pair
# below is created; exactly one group exists (count guards) and outputs.tf
# picks whichever was built.
#
# The ZONAL group: VM, hybrid, or internet endpoints inside a VPC, the
# backend of regional and passthrough load balancers. Everything but the
# endpoint list is immutable (ForceNew).
resource "google_compute_network_endpoint_group" "this" {
  count = local.is_zonal ? 1 : 0

  name        = local.neg_name
  project     = local.project_id
  zone        = var.spec.zone
  description = local.description

  network               = var.spec.network
  subnetwork            = local.subnetwork
  network_endpoint_type = local.network_endpoint_type
  default_port          = var.spec.default_port

  deletion_policy = local.deletion_policy
}

# The zonal group's membership as ONE set through Google's bulk endpoint
# operation: the manifest's list is the group's whole membership, so an
# endpoint removed from the list is detached in place and one added is
# attached. Created only when the list is non-empty, so a group declared
# ahead of its members (an autoscaler may own membership) carries no
# endpoint resource at all.
resource "google_compute_network_endpoints" "this" {
  count = local.is_zonal && length(local.zonal_endpoints) > 0 ? 1 : 0

  project                = local.project_id
  zone                   = var.spec.zone
  network_endpoint_group = google_compute_network_endpoint_group.this[0].name

  dynamic "network_endpoints" {
    for_each = local.zonal_endpoints
    content {
      instance   = network_endpoints.value.instance
      ip_address = network_endpoints.value.ip_address
      port       = network_endpoints.value.port
    }
  }

  deletion_policy = local.deletion_policy
}

# The GLOBAL internet group: INTERNET_IP_PORT or INTERNET_FQDN_PORT
# endpoints outside Google Cloud, the backend of a global external
# Application Load Balancer. Immutable in every argument.
resource "google_compute_global_network_endpoint_group" "this" {
  count = local.is_zonal ? 0 : 1

  name        = local.neg_name
  project     = local.project_id
  description = local.description

  network_endpoint_type = var.spec.network_endpoint_type
  default_port          = var.spec.default_port

  deletion_policy = local.deletion_policy
}

# Global endpoints have no bulk twin: each is its own resource, keyed by
# its position in the list. Every argument is immutable, so a changed
# endpoint is replaced.
resource "google_compute_global_network_endpoint" "this" {
  for_each = local.is_zonal ? {} : local.global_endpoints

  project                       = local.project_id
  global_network_endpoint_group = google_compute_global_network_endpoint_group.this[0].name

  ip_address = each.value.ip_address
  fqdn       = each.value.fqdn
  port       = each.value.port

  deletion_policy = local.deletion_policy
}
