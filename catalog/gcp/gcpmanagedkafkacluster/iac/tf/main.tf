# Enable the Managed Kafka API first so a fresh project works on the first
# deploy. disable_on_destroy is false: tearing down one cluster must never
# disable the API for every other cluster and Connect cluster in the
# project.
resource "google_project_service" "managedkafka_api" {
  project = local.project_id
  service = "managedkafka.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# The cluster. project, location, cluster_id, and the KMS key are
# immutable; capacity, networks, rebalancing, TLS, and labels update in
# place. Internet (public) access is not offered: the pinned Pulumi SDK
# lacks public_cluster_config, and a one-engine field would break parity.
resource "google_managed_kafka_cluster" "this" {
  project    = local.project_id
  location   = var.spec.location
  cluster_id = local.cluster_id

  # Counts are int64 in the spec; Google's API takes them as decimal
  # strings.
  capacity_config {
    vcpu_count   = tostring(var.spec.capacity_config.vcpu_count)
    memory_bytes = tostring(var.spec.capacity_config.memory_bytes)
  }

  gcp_config {
    kms_key = local.kms_key

    access_config {
      dynamic "network_configs" {
        for_each = local.subnets
        content {
          subnet = network_configs.value
        }
      }
    }
  }

  # The spec lifts the block's one leaf; the block is sent only when a disk
  # size is declared, leaving Google's default otherwise.
  dynamic "broker_capacity_config" {
    for_each = var.spec.broker_disk_size_gib > 0 ? [var.spec.broker_disk_size_gib] : []
    content {
      disk_size_gib = tostring(broker_capacity_config.value)
    }
  }

  dynamic "rebalance_config" {
    for_each = local.rebalance_mode != null ? [local.rebalance_mode] : []
    content {
      mode = rebalance_config.value
    }
  }

  # Declared (even empty) means sent: an empty tls_config block is how an
  # earlier mTLS configuration is cleared, and an undeclared one leaves
  # Google's current configuration alone (the block is Optional+Computed).
  dynamic "tls_config" {
    for_each = var.spec.tls_config != null ? [var.spec.tls_config] : []
    content {
      ssl_principal_mapping_rules = tls_config.value.ssl_principal_mapping_rules != "" ? tls_config.value.ssl_principal_mapping_rules : null

      dynamic "trust_config" {
        for_each = length(tls_config.value.ca_pools) > 0 ? [tls_config.value.ca_pools] : []
        content {
          dynamic "cas_configs" {
            for_each = trust_config.value
            content {
              ca_pool = cas_configs.value
            }
          }
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
