# Enable the Compute Engine API so a fresh project can host the rule.
# disable_on_destroy is false: tearing down one forwarding rule must never
# disable the API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A Compute Engine forwarding rule — the VIP node where traffic enters a
# load balancer (or, with the PSC form, where a VPC's private path to Google
# APIs / a producer service begins). It binds an IP address and port to a
# target proxy — or, for the passthrough Network Load Balancers, straight to
# a backend service; everything behind it (proxy → URL map → backend service
# → backends) is wiring.
#
# GCP models the global and regional forwarding rules as two API
# collections. They share the VIP surface (address, protocol, port range,
# scheme, network wiring, labels); the regional one adds the passthrough
# and PSC-consumer levers (backend_service, ports, all_ports, global
# access, service label, mirroring, BYOIP, source ranges) and lacks the
# Traffic Director and backend-bucket-migration levers. spec.region selects
# which resource below is created; the two blocks mirror each other so a
# manifest reads the same on either scope.
#
# target and labels update in place (setTarget is the zero-downtime frontend
# swap); every other field is immutable (ForceNew). The VIP itself survives
# recreation only when ip_address references a reserved static address —
# which is why production frontends reserve one.
resource "google_compute_global_forwarding_rule" "this" {
  count = local.is_regional ? 0 : 1

  name        = local.forwarding_rule_name
  project     = local.project_id
  description = local.description

  # Arrives resolved to a literal: a proxy self-link, a PSC bundle name
  # (all-apis / vpc-sc), or a service attachment URI.
  target = local.target

  # Null → GCP assigns an ephemeral IP. A GcpGlobalAddress ref resolves to
  # its literal IP (the API reads back the IP number, so passing the number
  # keeps state drift-free).
  ip_address = local.ip_address

  ip_protocol = local.ip_protocol
  ip_version  = local.ip_version

  # The spec's NONE sentinel maps to the API's empty scheme (Private
  # Service Connect); an empty spec value is sent as EXTERNAL explicitly
  # (see locals.tf) so the provider's own default never rebuilds a classic
  # frontend.
  load_balancing_scheme = local.load_balancing_scheme

  port_range = local.port_range

  network    = local.network
  subnetwork = local.subnetwork

  # Global rules are PREMIUM-only (spec CEL enforces it).
  network_tier = local.network_tier

  # Traffic Director xDS scoping (INTERNAL_SELF_MANAGED only).
  dynamic "metadata_filters" {
    for_each = var.spec.metadata_filters
    content {
      filter_match_criteria = metadata_filters.value.filter_match_criteria

      dynamic "filter_labels" {
        for_each = metadata_filters.value.filter_labels
        content {
          name  = filter_labels.value.name
          value = filter_labels.value.value
        }
      }
    }
  }

  # Service Directory registration for PSC-for-Google-APIs frontends. The
  # global resource registers a namespace and a region; the service field
  # is the regional rule's (spec CEL).
  dynamic "service_directory_registrations" {
    for_each = var.spec.service_directory_registration != null ? [var.spec.service_directory_registration] : []
    content {
      namespace                = service_directory_registrations.value.namespace != "" ? service_directory_registrations.value.namespace : null
      service_directory_region = service_directory_registrations.value.service_directory_region != "" ? service_directory_registrations.value.service_directory_region : null
    }
  }

  # Only meaningful for PSC; null keeps the API default (auto-create the
  # PSC DNS zone).
  no_automate_dns_zone = local.no_automate_dns_zone

  labels = local.labels

  # EXTERNAL → EXTERNAL_MANAGED backend-bucket canary migration.
  external_managed_backend_bucket_migration_state              = local.migration_state
  external_managed_backend_bucket_migration_testing_percentage = local.migration_testing_percentage

  # What destroy does to the frontend: DELETE (default), PREVENT (refuse),
  # or ABANDON (drop from state, keep serving traffic).
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]
}

# The regional twin: the front door of the regional external and internal
# Application Load Balancers (target = a regional proxy), of the internal
# and external passthrough Network Load Balancers (backend_service, no
# proxy), and of a Private Service Connect consumer endpoint (target = a
# service attachment, empty scheme).
resource "google_compute_forwarding_rule" "this" {
  count = local.is_regional ? 1 : 0

  name        = local.forwarding_rule_name
  project     = local.project_id
  region      = var.spec.region
  description = local.description

  # Exactly one of the two is set (spec CEL): a regional proxy self-link or
  # a service attachment URI, or the regional backend service of a
  # passthrough NLB.
  target          = local.target
  backend_service = local.backend_service

  # A regional GcpAddress reference (or a literal); null → ephemeral IP.
  ip_address = local.ip_address

  # L3_DEFAULT (every protocol) is admitted here alone; the spec pairs it
  # with all_ports.
  ip_protocol = local.ip_protocol
  ip_version  = local.ip_version

  # The regional provider default is EXTERNAL too, so the explicit send in
  # locals.tf changes nothing here; it keeps one contract on both scopes.
  load_balancing_scheme = local.load_balancing_scheme

  # One of three port forms (spec CEL); the unchosen ones stay null so the
  # API's own exclusivity never sees two.
  port_range = local.port_range
  ports      = local.ports
  all_ports  = local.all_ports

  # Required for the internal schemes and PSC, and for the regional
  # external ALB whose proxy-only subnet lives in the network; null lets
  # the API pick the default network where one applies.
  network    = local.network
  subnetwork = local.subnetwork

  # PREMIUM or STANDARD; null keeps the API's computed default (PREMIUM).
  network_tier = local.network_tier

  # Service Directory registration for a PSC consumer endpoint: namespace
  # and service (the region field is the global rule's, spec CEL).
  dynamic "service_directory_registrations" {
    for_each = var.spec.service_directory_registration != null ? [var.spec.service_directory_registration] : []
    content {
      namespace = service_directory_registrations.value.namespace != "" ? service_directory_registrations.value.namespace : null
      service   = service_directory_registrations.value.service != "" ? service_directory_registrations.value.service : null
    }
  }

  no_automate_dns_zone = local.no_automate_dns_zone

  labels = local.labels

  # The passthrough and PSC-consumer levers; each null when unset so the
  # API default stands and a re-plan stays clean. recreate_closed_psc
  # carries a provider default (false) and is sent explicitly.
  allow_global_access     = local.allow_global_access
  allow_psc_global_access = local.allow_psc_global_access
  service_label           = local.service_label
  is_mirroring_collector  = local.is_mirroring_collector
  ip_collection           = local.ip_collection
  recreate_closed_psc     = var.spec.recreate_closed_psc
  source_ip_ranges        = local.source_ip_ranges

  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]
}
