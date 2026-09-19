# Attaches the service project to its Shared VPC host.
#
# Both projects are immutable (moving to another host is a detach and an
# attach). deletion_policy on this resource is a single opt-in: ABANDON
# leaves the attachment in place on destroy; anything else (the default)
# detaches, which Google refuses while resources in the service project
# still use a host subnetwork.
resource "google_compute_shared_vpc_service_project" "this" {
  host_project    = var.spec.host_project_id
  service_project = var.spec.service_project_id

  deletion_policy = local.deletion_policy
}
