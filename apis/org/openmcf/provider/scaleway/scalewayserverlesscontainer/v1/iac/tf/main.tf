# ScalewayServerlessContainer Terraform Module
#
# Composite resource: creates a Scaleway container namespace, the
# container itself, and optional cron triggers.
#
# Resources created:
#   - scaleway_container_namespace (1x) -- grouping container
#   - scaleway_container (1x) -- the serverless container
#   - scaleway_container_cron (0..Nx) -- optional scheduled triggers

# ── Container Namespace ─────────────────────────────────────────────────

resource "scaleway_container_namespace" "namespace" {
  name        = local.name
  description = local.description
  region      = local.region
  tags        = local.standard_tags
}

# ── Container ───────────────────────────────────────────────────────────

resource "scaleway_container" "container" {
  namespace_id   = scaleway_container_namespace.namespace.id
  name           = local.name
  description    = local.description
  registry_image = local.registry_image
  privacy        = local.privacy
  port           = local.port
  protocol       = local.protocol

  memory_limit = local.memory_limit
  min_scale    = local.min_scale
  max_scale    = local.max_scale
  timeout      = local.timeout
  http_option  = local.http_option
  deploy       = local.deploy

  environment_variables        = local.env_vars_map
  secret_environment_variables = local.secret_env_vars_map

  tags = local.standard_tags

  # Optional: CPU limit.
  cpu_limit = local.cpu_limit > 0 ? local.cpu_limit : null

  # Optional: execution environment.
  sandbox = local.sandbox != "" ? local.sandbox : null

  # Optional: deployment trigger.
  registry_sha256 = local.registry_sha256 != "" ? local.registry_sha256 : null

  # Optional: local storage limit.
  local_storage_limit = local.local_storage_limit_mb > 0 ? local.local_storage_limit_mb : null

  # Optional: Private Network connectivity.
  private_network_id = local.private_network_id

  # Optional: command and args override.
  # Note: Terraform provider uses singular "command" (not "commands").
  command = length(var.spec.commands) > 0 ? var.spec.commands : null
  args    = length(var.spec.args) > 0 ? var.spec.args : null

  # Optional: health check.
  dynamic "health_check" {
    for_each = var.spec.health_check != null ? [var.spec.health_check] : []
    content {
      failure_threshold = health_check.value.failure_threshold
      interval          = "${health_check.value.interval_seconds}s"
      http {
        path = health_check.value.path
      }
    }
  }

  # Optional: scaling options.
  dynamic "scaling_option" {
    for_each = var.spec.scaling_option != null ? [var.spec.scaling_option] : []
    content {
      concurrent_requests_threshold = scaling_option.value.concurrent_requests_threshold > 0 ? scaling_option.value.concurrent_requests_threshold : null
      cpu_usage_threshold           = scaling_option.value.cpu_usage_threshold > 0 ? scaling_option.value.cpu_usage_threshold : null
      memory_usage_threshold        = scaling_option.value.memory_usage_threshold > 0 ? scaling_option.value.memory_usage_threshold : null
    }
  }

  # Ignore changes to secret_environment_variables to prevent
  # unnecessary updates when secrets are managed externally.
  lifecycle {
    ignore_changes = [
      secret_environment_variables,
    ]
  }

  depends_on = [scaleway_container_namespace.namespace]
}

# ── Cron Triggers ───────────────────────────────────────────────────────

resource "scaleway_container_cron" "triggers" {
  for_each = local.cron_triggers

  container_id = scaleway_container.container.id
  name         = each.key
  schedule     = each.value.schedule
  args         = each.value.args

  depends_on = [scaleway_container.container]
}
