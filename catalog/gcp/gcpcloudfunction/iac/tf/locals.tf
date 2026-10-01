locals {
  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # Function name defaults to metadata.name.
  function_name = var.spec.function_name != "" ? var.spec.function_name : var.metadata.name

  # Platform attribution labels. User labels merge BENEATH them so the
  # platform's keys always win — the same order the Pulumi module applies.
  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = local.function_name
    "planton-ai_kind"     = "gcpcloudfunction"
  }

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "planton-ai_organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "planton-ai_environment" = var.metadata.env } : {}

  id_label = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "planton-ai_id" = var.metadata.id } : {}

  final_labels = merge(
    var.spec.labels,
    local.base_labels,
    local.org_label,
    local.env_label,
    local.id_label,
  )

  # Enum fields arrive from the converter as enum-name strings; empty means
  # "not set" and defers to the API default.
  is_http_trigger = var.spec.trigger == null || var.spec.trigger.trigger_type == "" || var.spec.trigger.trigger_type == "HTTP"

  service_config = var.spec.service_config

  # The Gen 2 API takes memory as a quantity string ("256M", "1Gi"); the
  # spec carries it verbatim. Empty defers to the API default (256M).
  available_memory = try(local.service_config.available_memory, "") != "" ? local.service_config.available_memory : null
  available_cpu    = try(local.service_config.available_cpu, "") != "" ? local.service_config.available_cpu : null

  vpc_connector = try(local.service_config.vpc_connector, "") != "" ? local.service_config.vpc_connector : null
  # Egress settings only make sense with a connector attached (sending them
  # without one is an API error) — but WITH a connector they are sent
  # explicitly on every apply, defaulting PRIVATE_RANGES_ONLY: the API
  # materializes that default server-side and echoes it into state, so an
  # omitted value re-plans as a perpetual remove-diff (live-caught by the
  # idempotency gate). Same explicit-send grain as direct_vpc_egress below.
  vpc_egress = local.vpc_connector != null ? (try(local.service_config.vpc_connector_egress_settings, "") != "" ? local.service_config.vpc_connector_egress_settings : "PRIVATE_RANGES_ONLY") : null

  ingress_settings = try(local.service_config.ingress_settings, "") != "" ? local.service_config.ingress_settings : null

  # Direct VPC egress (the connectorless path; mutually exclusive with
  # vpc_connector, enforced pre-deploy). The egress mode is sent
  # explicitly whenever the interface is present — the field is
  # Optional+Computed on the provider, so an ALL_TRAFFIC ->
  # PRIVATE_RANGES_ONLY spec edit would otherwise silently keep the old
  # live value. The API takes VPC_EGRESS_-prefixed values; the spec
  # carries the bare enum names.
  direct_vpc        = try(local.service_config.direct_vpc_network_interface, null)
  direct_vpc_egress = local.direct_vpc != null ? "VPC_EGRESS_${try(local.service_config.direct_vpc_egress, "") != "" ? local.service_config.direct_vpc_egress : "PRIVATE_RANGES_ONLY"}" : null

  # A Secret Manager entry naming a secret the author owns that omits its
  # project needs a concrete project id in the API payload (a stored
  # value's project comes from the secret secrets.tf created). Only when
  # the function itself also rides the ambient project (spec.project_id
  # empty) must the module LOOK UP that project — count-gating the data
  # source keeps every explicitly-scoped plan credential-free.
  needs_project_lookup = local.project_id == null && local.service_config != null && (
    length([for s in try(local.service_config.secret_environment_variables, []) : s if s.secret != "" && s.project_id == ""]) > 0 ||
    length([for v in try(local.service_config.secret_volumes, []) : v if v.project_id == ""]) > 0
  )

  # The project stamped on secret entries that omit their own.
  effective_secret_project = local.project_id != null ? local.project_id : (local.needs_project_lookup ? data.google_project.function[0].project_id : null)

  build_update_policy = var.spec.build_config.update_policy
}
