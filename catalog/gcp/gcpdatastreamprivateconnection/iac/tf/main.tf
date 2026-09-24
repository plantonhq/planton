# Enable the Datastream API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one private connection
# must never disable the API for every profile and stream in the project.
resource "google_project_service" "datastream_api" {
  project = local.project_id
  service = "datastream.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The private connection. Everything but labels and deletion_policy is
# immutable. Exactly one connectivity block is declared (the spec's rule).
resource "google_datastream_private_connection" "this" {
  project               = local.project_id
  location              = var.spec.location
  private_connection_id = local.private_connection_id
  display_name          = local.display_name

  create_without_validation = var.spec.create_without_validation

  dynamic "vpc_peering_config" {
    for_each = var.spec.vpc_peering_config != null ? [var.spec.vpc_peering_config] : []
    content {
      vpc    = vpc_peering_config.value.vpc
      subnet = vpc_peering_config.value.subnet
    }
  }

  dynamic "psc_interface_config" {
    for_each = var.spec.psc_interface_config != null ? [var.spec.psc_interface_config] : []
    content {
      network_attachment = psc_interface_config.value.network_attachment
    }
  }

  labels = local.final_labels

  # Client-side destroy behavior: FORCE (the provider's default, which also
  # removes Datastream's routes), DELETE, PREVENT, or ABANDON. Sent only when
  # set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.datastream_api]
}
