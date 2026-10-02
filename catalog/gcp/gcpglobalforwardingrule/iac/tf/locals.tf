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

  # The scope selector. An empty region builds the global forwarding rule;
  # a region name builds the regional one. Exactly one of the two resources
  # in main.tf exists (count guards), and outputs.tf picks whichever was
  # created.
  is_regional = var.spec.region != null && var.spec.region != ""

  # The cloud-side name defaults to metadata.name when the spec leaves
  # forwarding_rule_name empty — the same naming basis every kind uses.
  forwarding_rule_name = (
    var.spec.forwarding_rule_name != null && var.spec.forwarding_rule_name != ""
    ? var.spec.forwarding_rule_name
    : var.metadata.name
  )

  description = var.spec.description != "" ? var.spec.description : null

  # Exactly one of the two sinks is set (spec CEL): a proxy-based load
  # balancer names its target, a passthrough Network Load Balancer names its
  # backend service. Both arrive resolved to literals.
  target          = var.spec.target != "" ? var.spec.target : null
  backend_service = var.spec.backend_service != "" ? var.spec.backend_service : null

  ip_address = var.spec.ip_address != "" ? var.spec.ip_address : null

  # The middleware default (TCP) matches GCP's own default; null lets the
  # API compute TCP when unset.
  ip_protocol = var.spec.ip_protocol != "" ? var.spec.ip_protocol : null

  ip_version = var.spec.ip_version != "" ? var.spec.ip_version : null

  # The spec's NONE sentinel is the Private Service Connect form, which the
  # API expects as an EMPTY scheme — the one case where "" must be SENT
  # rather than treated as unset. Anything else passes through verbatim.
  # An empty spec value becomes EXTERNAL, never null, on BOTH scopes: the
  # spec's default is EXTERNAL (the classic global external ALB, or the
  # external passthrough NLB on a regional rule), the global provider's own
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

  # The three port forms are mutually exclusive (spec CEL); each is sent
  # only when chosen. ports and all_ports exist on the regional resource
  # alone.
  port_range = var.spec.port_range != "" ? var.spec.port_range : null
  ports      = length(var.spec.ports) > 0 ? var.spec.ports : null
  all_ports  = var.spec.all_ports ? true : null

  network    = var.spec.network != "" ? var.spec.network : null
  subnetwork = var.spec.subnetwork != "" ? var.spec.subnetwork : null

  # Empty keeps the API's computed default (PREMIUM) in charge; STANDARD is
  # a regional-rule value the spec CEL admits only with region set.
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

  # Regional-only levers. The booleans default false on the API, so only an
  # explicit true is sent (null lets the API compute the default and a
  # re-plan stays clean); recreate_closed_psc carries a provider default of
  # false and is sent explicitly as the catalog does for every
  # default-bearing lever.
  allow_global_access     = var.spec.allow_global_access ? true : null
  allow_psc_global_access = var.spec.allow_psc_global_access ? true : null
  is_mirroring_collector  = var.spec.is_mirroring_collector ? true : null
  service_label           = var.spec.service_label != "" ? var.spec.service_label : null
  ip_collection           = var.spec.ip_collection != "" ? var.spec.ip_collection : null
  source_ip_ranges        = length(var.spec.source_ip_ranges) > 0 ? var.spec.source_ip_ranges : null

  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
