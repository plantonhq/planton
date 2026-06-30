# main.tf - GCP Cloud Run service provisioning

# Cloud Run v2 Service
resource "google_cloud_run_v2_service" "main" {
  project  = var.spec.project_id.value
  location = var.spec.region
  name     = local.service_name

  # Public access settings - disable invoker IAM check for unauthenticated access
  ingress = local.ingress

  labels = local.labels

  # Deletion protection at GCP resource level
  deletion_protection = var.spec.delete_protection

  template {
    # Service account
    service_account = local.service_account

    # Timeout
    timeout = local.timeout

    # Execution environment
    execution_environment = local.execution_environment

    # Max concurrent requests per instance
    max_instance_request_concurrency = var.spec.max_concurrency

    # Scaling configuration
    scaling {
      min_instance_count = var.spec.container.replicas.min
      max_instance_count = var.spec.container.replicas.max
    }

    # VPC access configuration (conditional)
    dynamic "vpc_access" {
      for_each = local.has_vpc_access ? [1] : []
      content {
        dynamic "network_interfaces" {
          for_each = local.vpc_network != null && local.vpc_network != "" ? [1] : []
          content {
            network    = local.vpc_network
            subnetwork = local.vpc_subnet
          }
        }
        egress = var.spec.vpc_access.egress != null ? var.spec.vpc_access.egress : null
      }
    }

    # Container configuration
    containers {
      image = local.image

      # Container port
      ports {
        container_port = local.port
      }

      # Resource limits
      resources {
        limits = {
          memory = local.memory
          cpu    = local.cpu
        }
      }

      # Environment variables (plain)
      dynamic "env" {
        for_each = local.env_vars
        content {
          name  = env.value.name
          value = env.value.value
        }
      }

      # Environment variables (secrets from Secret Manager)
      dynamic "env" {
        for_each = local.env_secrets
        content {
          name = env.value.name
          value_source {
            secret_key_ref {
              secret = env.value.secret
            }
          }
        }
      }
    }
  }

  # Lifecycle settings to prevent unnecessary recreations
  lifecycle {
    create_before_destroy = true
  }
}

# IAM policy for unauthenticated access (if enabled)
resource "google_cloud_run_v2_service_iam_member" "public_invoker" {
  count = var.spec.allow_unauthenticated ? 1 : 0

  project  = google_cloud_run_v2_service.main.project
  location = google_cloud_run_v2_service.main.location
  name     = google_cloud_run_v2_service.main.name

  role   = "roles/run.invoker"
  member = "allUsers"
}

# Custom DNS Domain Mapping (if DNS is enabled)
resource "google_cloud_run_domain_mapping" "main" {
  count = local.dns_enabled && length(local.dns_hostnames) > 0 ? 1 : 0

  location = google_cloud_run_v2_service.main.location
  name     = local.dns_hostnames[0]

  metadata {
    namespace = var.spec.project_id.value
    labels    = local.labels
  }

  spec {
    route_name = google_cloud_run_v2_service.main.name
  }
}

# DNS TXT record for domain verification (if DNS is enabled)
resource "google_dns_record_set" "domain_verification" {
  count = local.dns_enabled && length(local.dns_hostnames) > 0 ? 1 : 0

  managed_zone = local.dns_managed_zone
  name         = "${local.dns_hostnames[0]}."
  type         = "TXT"
  ttl          = 300

  rrdatas = [
    google_cloud_run_domain_mapping.main[0].status[0].resource_records[0].rrdata
  ]
}
