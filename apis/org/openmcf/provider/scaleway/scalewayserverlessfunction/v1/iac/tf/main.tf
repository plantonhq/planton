# ScalewayServerlessFunction Terraform Module
#
# Composite resource: creates a Scaleway function namespace, the
# function itself, and optional cron triggers.
#
# Resources created:
#   - scaleway_function_namespace (1x) -- grouping container
#   - scaleway_function (1x) -- the serverless function
#   - scaleway_function_cron (0..Nx) -- optional scheduled triggers

# ── Function Namespace ──────────────────────────────────────────────────

resource "scaleway_function_namespace" "namespace" {
  name        = local.name
  description = local.description
  region      = local.region
  tags        = local.standard_tags
}

# ── Function ────────────────────────────────────────────────────────────

resource "scaleway_function" "function" {
  namespace_id = scaleway_function_namespace.namespace.id
  name         = local.name
  runtime      = local.runtime
  handler      = local.handler
  privacy      = local.privacy
  description  = local.description

  memory_limit = local.memory_limit
  min_scale    = local.min_scale
  max_scale    = local.max_scale
  timeout      = local.timeout
  http_option  = local.http_option

  environment_variables        = local.env_vars_map
  secret_environment_variables = local.secret_env_vars_map

  tags = local.standard_tags

  # Optional: execution environment.
  sandbox = local.sandbox != "" ? local.sandbox : null

  # Optional: zip-based code deployment.
  zip_file = local.zip_file != "" ? local.zip_file : null
  zip_hash = local.zip_hash != "" ? local.zip_hash : null
  deploy   = local.deploy

  # Optional: Private Network connectivity.
  private_network_id = local.private_network_id

  # Ignore changes to secret_environment_variables to prevent
  # unnecessary updates when secrets are managed externally.
  lifecycle {
    ignore_changes = [
      secret_environment_variables,
    ]
  }

  depends_on = [scaleway_function_namespace.namespace]
}

# ── Cron Triggers ───────────────────────────────────────────────────────

resource "scaleway_function_cron" "triggers" {
  for_each = local.cron_triggers

  function_id = scaleway_function.function.id
  name        = each.key
  schedule    = each.value.schedule
  args        = each.value.args

  depends_on = [scaleway_function.function]
}
