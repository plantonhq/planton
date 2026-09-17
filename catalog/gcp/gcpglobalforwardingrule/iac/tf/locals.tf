locals {
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Honor the spec contract: an empty project_id falls back to the provider's
  # default project. Passing null (instead of "") lets the google provider
  # resolve its own project from configuration or the GOOGLE_PROJECT /
  # GOOGLE_CLOUD_PROJECT environment chain; an empty string would be sent
  # verbatim and rejected by the API.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null

  # The cloud-side name defaults to metadata.name when the spec leaves
  # forwarding_rule_name empty — the same naming basis every kind uses.
  forwarding_rule_name = (
    var.spec.forwarding_rule_name != null && var.spec.forwarding_rule_name != ""
    ? var.spec.forwarding_rule_name
    : var.metadata.name
  )

  description = var.spec.description != "" ? var.spec.description : null

  ip_address = var.spec.ip_address != "" ? var.spec.ip_address : null

  # The middleware default (TCP) matches GCP's own default; null lets the
  # API compute TCP when unset.
  ip_protocol = var.spec.ip_protocol != "" ? var.spec.ip_protocol : null

  ip_version = var.spec.ip_version != "" ? var.spec.ip_version : null

  # The spec's NONE sentinel is the Private Service Connect form, which the
  # API expects as an EMPTY scheme — the one case where "" must be SENT
  # rather than treated as unset. Anything else passes through verbatim.
  # An empty spec value becomes EXTERNAL, never null: the spec's default
  # is EXTERNAL (the classic global external ALB), the provider's own
  # default is EXTERNAL_MANAGED, and the scheme is immutable — letting the
  # provider decide would replace every existing classic frontend the next
  # time it was applied. The manifest defaults applier normally fills
  # EXTERNAL first; this guard covers every path that bypasses it. The
  # Pulumi module makes the same choice.
  load_balancing_scheme = (
    var.spec.load_balancing_scheme == "NONE"
    ? ""
    : (var.spec.load_balancing_scheme != "" ? var.spec.load_balancing_scheme : "EXTERNAL")
  )

  port_range = var.spec.port_range != "" ? var.spec.port_range : null

  network    = var.spec.network != "" ? var.spec.network : null
  subnetwork = var.spec.subnetwork != "" ? var.spec.subnetwork : null

  # Global rules are PREMIUM-only (spec CEL enforces it); null keeps the
  # API's computed default in charge.
  network_tier = var.spec.network_tier != "" ? var.spec.network_tier : null

  # Only meaningful for PSC; the API default (auto-create the DNS zone)
  # applies unless explicitly disabled.
  no_automate_dns_zone = var.spec.no_automate_dns_zone ? true : null

  labels = length(var.spec.labels) > 0 ? var.spec.labels : null

  migration_state = (
    var.spec.external_managed_backend_bucket_migration_state != ""
    ? var.spec.external_managed_backend_bucket_migration_state
    : null
  )

  migration_testing_percentage = (
    var.spec.external_managed_backend_bucket_migration_testing_percentage != 0
    ? var.spec.external_managed_backend_bucket_migration_testing_percentage
    : null
  )
}
