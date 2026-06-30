# Confluent Cloud Kafka Cluster
# https://registry.terraform.io/providers/confluentinc/confluent/latest/docs/resources/confluent_kafka_cluster

resource "confluent_kafka_cluster" "main" {
  display_name = local.display_name
  availability = var.spec.availability
  cloud        = var.spec.cloud
  region       = var.spec.region

  # Environment configuration
  environment {
    id = var.spec.environment_id
  }

  # Network configuration (if provided)
  dynamic "network" {
    for_each = local.has_network_config ? [1] : []
    content {
      id = var.spec.network_config.network_id
    }
  }

  # Cluster type-specific configuration
  # Only one of basic, standard, enterprise, or dedicated should be set

  dynamic "basic" {
    for_each = local.cluster_type == "BASIC" ? [1] : []
    content {}
  }

  dynamic "standard" {
    for_each = local.cluster_type == "STANDARD" ? [1] : []
    content {}
  }

  dynamic "enterprise" {
    for_each = local.cluster_type == "ENTERPRISE" ? [1] : []
    content {}
  }

  dynamic "dedicated" {
    for_each = local.cluster_type == "DEDICATED" ? [1] : []
    content {
      cku = var.spec.dedicated_config.cku
    }
  }

  # Lifecycle management
  lifecycle {
    # Prevent accidental deletion of production clusters
    prevent_destroy = false

    # Validate dedicated cluster configuration
    precondition {
      condition     = !(local.is_dedicated && !local.has_dedicated_config)
      error_message = "dedicated_config with cku is required when cluster_type is DEDICATED"
    }

    # Validate network configuration for basic/standard clusters
    precondition {
      condition     = !(local.has_network_config && (local.cluster_type == "BASIC" || local.cluster_type == "STANDARD"))
      error_message = "network_config is only supported for ENTERPRISE and DEDICATED cluster types"
    }
  }
}

