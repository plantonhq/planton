# Enable the Managed Kafka API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one Connect cluster
# must never disable the API for the Kafka clusters in the project.
resource "google_project_service" "managedkafka_api" {
  project = local.project_id
  service = "managedkafka.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The Connect cluster. project, location, and connect_cluster_id are
# immutable; the attached Kafka cluster, capacity, networks, and labels
# update in place.
resource "google_managed_kafka_connect_cluster" "this" {
  project            = local.project_id
  location           = var.spec.location
  connect_cluster_id = local.connect_cluster_id
  kafka_cluster      = var.spec.kafka_cluster

  # Counts are int64 in the spec; Google's API takes them as decimal
  # strings.
  capacity_config {
    vcpu_count   = tostring(var.spec.capacity_config.vcpu_count)
    memory_bytes = tostring(var.spec.capacity_config.memory_bytes)
  }

  gcp_config {
    access_config {
      dynamic "network_configs" {
        for_each = local.network_configs
        content {
          primary_subnet   = network_configs.value.primary_subnet
          dns_domain_names = network_configs.value.dns_domain_names
        }
      }
    }
  }

  labels = local.final_labels

  # Client-side destroy behavior: DELETE (default), PREVENT, or ABANDON.
  # Sent only when set so the provider default stays in charge otherwise.
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.managedkafka_api]
}
