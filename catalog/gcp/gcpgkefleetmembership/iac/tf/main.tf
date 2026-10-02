# The Fleet API (GKE Hub). disable_on_destroy is false: unregistering one
# cluster must never disable the API for the rest of the fleet.
resource "google_project_service" "gkehub_api" {
  project = local.project_id
  service = "gkehub.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The membership: a cluster registered with the fleet. The ID, location,
# cluster, and issuer are create-time decisions; labels update in place.
resource "google_gke_hub_membership" "this" {
  project         = local.project_id
  membership_id   = local.membership_id
  location        = local.location
  labels          = local.final_labels
  deletion_policy = local.deletion_policy

  dynamic "endpoint" {
    for_each = local.gke_cluster_resource_link != null ? [local.gke_cluster_resource_link] : []
    content {
      gke_cluster {
        resource_link = endpoint.value
      }
    }
  }

  # Fleet Workload Identity: Google trusts this issuer's OIDC tokens within
  # the fleet's workload identity pool.
  dynamic "authority" {
    for_each = var.spec.issuer != "" ? [var.spec.issuer] : []
    content {
      issuer = authority.value
    }
  }

  depends_on = [google_project_service.gkehub_api]
}
