###############################################################################
# Google Cloud Function (Gen 2)
###############################################################################
resource "google_cloudfunctions2_function" "function" {
  name     = local.function_name
  project  = var.spec.project_id
  location = var.spec.region

  # Build configuration
  build_config {
    runtime     = var.spec.build_config.runtime
    entry_point = var.spec.build_config.entry_point

    # Source from GCS bucket
    source {
      storage_source {
        bucket     = var.spec.build_config.source.bucket
        object     = var.spec.build_config.source.object
        generation = var.spec.build_config.source.generation
      }
    }

    # Build-time environment variables
    environment_variables = var.spec.build_config.build_environment_variables
  }

  # Service configuration
  service_config {
    # Memory and timeout
    available_memory = "${local.available_memory_mb}M"
    timeout_seconds  = local.timeout_seconds

    # Concurrency and scaling
    max_instance_request_concurrency = local.max_instance_request_concurrency
    min_instance_count               = local.min_instance_count
    max_instance_count               = local.max_instance_count

    # Service account (runtime identity)
    service_account_email = try(local.service_config.service_account_email, null)

    # Environment variables
    environment_variables = try(local.service_config.environment_variables, null)

    # Secret Manager integration
    dynamic "secret_environment_variables" {
      for_each = local.secret_environment_variables
      content {
        key        = secret_environment_variables.value.key
        project_id = secret_environment_variables.value.project_id
        secret     = secret_environment_variables.value.secret
        version    = secret_environment_variables.value.version
      }
    }

    # VPC connector for private resource access
    vpc_connector                 = try(local.service_config.vpc_connector, null)
    vpc_connector_egress_settings = try(local.service_config.vpc_connector, null) != null ? local.vpc_egress_settings : null

    # Ingress settings (network access control)
    ingress_settings = local.ingress_settings

    # Authentication settings (if allow_unauthenticated is explicitly set to true)
    all_traffic_on_latest_revision = true
  }

  # Resource labels
  labels = local.final_gcp_labels

  # Event trigger configuration (only for event-driven functions)
  dynamic "event_trigger" {
    for_each = !local.is_http_trigger && var.spec.trigger != null && var.spec.trigger.event_trigger != null ? [1] : []
    content {
      # Event type (CloudEvents format)
      event_type = var.spec.trigger.event_trigger.event_type

      # Pub/Sub topic (if specified)
      pubsub_topic = var.spec.trigger.event_trigger.pubsub_topic

      # Trigger region (defaults to function region if not specified)
      trigger_region = coalesce(
        var.spec.trigger.event_trigger.trigger_region,
        var.spec.region
      )

      # Retry policy
      retry_policy = local.event_retry_policy

      # Service account for Eventarc trigger
      service_account_email = var.spec.trigger.event_trigger.service_account_email

      # Event filters
      dynamic "event_filters" {
        for_each = var.spec.trigger.event_trigger.event_filters != null ? var.spec.trigger.event_trigger.event_filters : []
        content {
          attribute = event_filters.value.attribute
          value     = event_filters.value.value
          operator  = event_filters.value.operator
        }
      }
    }
  }

  # Lifecycle: prevent accidental deletion
  lifecycle {
    # Allow recreation if source code changes
    create_before_destroy = false
  }
}

###############################################################################
# IAM Policy Binding for Public Access (if allow_unauthenticated is true)
###############################################################################
resource "google_cloud_run_service_iam_member" "public_invoker" {
  # Only create this resource if:
  # 1. Function is HTTP-triggered (not event-driven)
  # 2. allow_unauthenticated is explicitly set to true
  count = local.is_http_trigger && try(local.service_config.allow_unauthenticated, false) ? 1 : 0

  project  = var.spec.project_id
  location = var.spec.region
  service  = google_cloudfunctions2_function.function.name

  role   = "roles/run.invoker"
  member = "allUsers"
}

