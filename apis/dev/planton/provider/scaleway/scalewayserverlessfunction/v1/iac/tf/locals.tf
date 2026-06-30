locals {
  # ── Resource naming ───────────────────────────────────────────────────

  name        = var.metadata.name
  region      = var.spec.region
  description = coalesce(var.spec.description, "")

  # ── Environment variables ─────────────────────────────────────────────
  #
  # Convert repeated name-value messages to maps for the Scaleway API.
  # The proto spec uses repeated messages (Kubernetes-style) for sort
  # order stability and future valueFrom extension. The Terraform
  # provider requires map(string) inputs.

  env_vars_map = {
    for ev in var.spec.env.variables : ev.name => ev.value
  }

  secret_env_vars_map = {
    for ev in var.spec.env.secrets : ev.name => ev.value
  }

  # ── Function configuration ────────────────────────────────────────────

  runtime         = var.spec.runtime
  handler         = var.spec.handler
  privacy         = var.spec.privacy
  memory_limit    = coalesce(var.spec.memory_limit_mb, 256)
  min_scale       = coalesce(var.spec.min_scale, 0)
  max_scale       = coalesce(var.spec.max_scale, 20)
  timeout         = coalesce(var.spec.timeout_seconds, 300)
  http_option     = coalesce(var.spec.http_option, "enabled")
  sandbox         = var.spec.sandbox

  # ── Code deployment ───────────────────────────────────────────────────

  zip_file = var.spec.zip_file
  zip_hash = var.spec.zip_hash
  deploy   = var.spec.zip_file != "" ? true : null

  # ── Networking ────────────────────────────────────────────────────────

  private_network_id = try(var.spec.private_network_id.value, null)

  # ── Cron triggers ─────────────────────────────────────────────────────
  #
  # Build a map keyed by trigger name (or index-based fallback) for
  # use with for_each.

  cron_triggers = {
    for idx, trigger in var.spec.cron_triggers :
    coalesce(trigger.name, "cron-${idx}") => trigger
  }

  # ── Tags ──────────────────────────────────────────────────────────────
  #
  # Standard Planton metadata tags as flat "key=value" strings.

  standard_tags = concat(
    [
      "planton-ai_resource=true",
      "planton-ai_resource-name=${var.metadata.name}",
      "planton-ai_resource-kind=ScalewayServerlessFunction",
    ],
    var.metadata.org != null ? ["planton-ai_organization=${var.metadata.org}"] : [],
    var.metadata.env != null ? ["planton-ai_environment=${var.metadata.env}"] : [],
    var.metadata.id != null ? ["planton-ai_resource-id=${var.metadata.id}"] : [],
  )
}
