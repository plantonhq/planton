# The provider's default project, read only when the spec names none -- the
# nodes' parent path needs it.
data "google_client_config" "current" {
  count = local.project_id == null ? 1 : 0
}

# Enable the Cloud TPU API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one request must never
# disable the API for everything else in the project.
resource "google_project_service" "tpu_api" {
  project = local.project_id
  service = "tpu.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The queued resource, on the beta provider under the recorded admission.
# Everything is immutable. The request waits for capacity, then provisions
# every node it describes; destroying it deletes those nodes.
resource "google_tpu_v2_queued_resource" "this" {
  provider = google-beta

  project = local.project_id
  zone    = var.spec.zone
  name    = local.queued_resource_id

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  deletion_policy = local.deletion_policy

  # The spec lifts Google's one-field tpu wrapper to node_specs.
  tpu {
    dynamic "node_spec" {
      for_each = var.spec.node_specs
      content {
        parent  = local.node_parent
        node_id = node_spec.value.node_id != "" ? node_spec.value.node_id : null

        node {
          runtime_version  = node_spec.value.node.runtime_version
          accelerator_type = node_spec.value.node.accelerator_type != "" ? node_spec.value.node.accelerator_type : null
          description      = node_spec.value.node.description != "" ? node_spec.value.node.description : null

          dynamic "network_config" {
            for_each = node_spec.value.node.network_config != null ? [node_spec.value.node.network_config] : []
            content {
              network             = network_config.value.network != "" ? network_config.value.network : null
              subnetwork          = network_config.value.subnetwork != "" ? trimprefix(network_config.value.subnetwork, local.compute_prefix) : null
              enable_external_ips = network_config.value.enable_external_ips
              can_ip_forward      = network_config.value.can_ip_forward
              queue_count         = network_config.value.queue_count != 0 ? network_config.value.queue_count : null
            }
          }
        }
      }
    }
  }

  depends_on = [google_project_service.tpu_api]
}
