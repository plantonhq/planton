locals {
  # Derive a stable resource ID
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

  # The scope selector. An empty region builds the global backend service; a
  # region name builds the regional one. Exactly one of the two resources
  # in main.tf exists (count guards), and outputs.tf picks whichever was
  # created.
  is_regional = var.spec.region != null && var.spec.region != ""

  # The cloud-side name defaults to metadata.name when the spec leaves
  # backend_service_name empty — the same naming basis every kind uses.
  backend_service_name = (
    var.spec.backend_service_name != null && var.spec.backend_service_name != ""
    ? var.spec.backend_service_name
    : var.metadata.name
  )

  # Normalize "" -> null for optional strings the provider treats as
  # meaningfully absent. Each of these has a GCP API default that matches
  # the spec's proto default (protocol HTTP, session_affinity NONE), so
  # null and the middleware-applied default are behaviorally identical.
  protocol  = var.spec.protocol != "" ? var.spec.protocol : null
  port_name = var.spec.port_name != "" ? var.spec.port_name : null
  # With an HA policy the leader IS the routing decision, and the provider
  # rejects session_affinity beside it even at the spec's NONE default (the
  # manifest loader fills NONE before the module runs), so the affinity is
  # dropped entirely on that arm. The Pulumi module makes the same choice.
  session_affinity            = var.spec.session_affinity != "" && var.spec.ha_policy == null ? var.spec.session_affinity : null
  locality_lb_policy          = var.spec.locality_lb_policy != "" ? var.spec.locality_lb_policy : null
  compression_mode            = var.spec.compression_mode != "" ? var.spec.compression_mode : null
  security_policy             = var.spec.security_policy != "" ? var.spec.security_policy : null
  edge_security_policy        = var.spec.edge_security_policy != "" ? var.spec.edge_security_policy : null
  ip_address_selection_policy = var.spec.ip_address_selection_policy != "" ? var.spec.ip_address_selection_policy : null
  service_lb_policy           = var.spec.service_lb_policy != "" ? var.spec.service_lb_policy : null
  migration_state             = var.spec.external_managed_migration_state != "" ? var.spec.external_managed_migration_state : null

  # The scheme is the ONE exception to the "" -> null rule: the spec's
  # default is EXTERNAL (the classic global external ALB, or the external
  # passthrough NLB on a regional service), but the provider's own default
  # differs per scope (EXTERNAL_MANAGED globally, INTERNAL regionally), and
  # the scheme is immutable -- letting the provider decide would replace
  # every existing classic backend service the next time it was applied,
  # and would give a manifest a different meaning depending on where it
  # lives. The module therefore sends EXTERNAL itself when the spec leaves
  # the scheme empty, on BOTH scopes (the manifest defaults applier normally
  # fills it first; this guard covers every path that bypasses the applier).
  # The Pulumi module makes the same choice.
  load_balancing_scheme = var.spec.load_balancing_scheme != "" ? var.spec.load_balancing_scheme : "EXTERNAL"

  # Regional-only: the VPC network of a passthrough Network Load Balancer.
  network = var.spec.network != "" ? var.spec.network : null

  # The tfvars converter emits 0 for unset proto numbers; 0 is not a
  # meaningful value for these, so 0 -> null lets the API apply its
  # defaults (timeout 30s, draining 300s, affinity cookie session-scoped).
  timeout_sec                     = try(var.spec.timeout_sec, 0) != 0 ? var.spec.timeout_sec : null
  connection_draining_timeout_sec = try(var.spec.connection_draining_timeout_sec, 0) != 0 ? var.spec.connection_draining_timeout_sec : null
  affinity_cookie_ttl_sec         = try(var.spec.affinity_cookie_ttl_sec, 0) != 0 ? var.spec.affinity_cookie_ttl_sec : null
  migration_testing_percentage    = try(var.spec.external_managed_migration_testing_percentage, 0) != 0 ? var.spec.external_managed_migration_testing_percentage : null

  # An empty health_check means none (valid only for serverless/internet NEG
  # backends); the provider takes a set of at most one.
  health_checks = var.spec.health_check != "" ? [var.spec.health_check] : null

  # Per-backend normalization. balancing_mode "" -> null (API default
  # UTILIZATION). capacity_scaler passes through untouched: it is `optional`
  # in the proto, so the converter emits null when unset (API default 1.0)
  # and a literal 0 only when the operator explicitly set 0 — which is the
  # drain-this-backend semantics and must survive. The max_* dials are
  # plain proto numbers where 0 means unset, so 0 -> null lets the API
  # enforce which dial the balancing mode requires.
  backends = [
    for backend in var.spec.backends : {
      group                        = backend.group
      balancing_mode               = backend.balancing_mode != "" ? backend.balancing_mode : null
      capacity_scaler              = try(backend.capacity_scaler, null)
      description                  = try(backend.description, null) != "" ? backend.description : null
      max_connections              = try(backend.max_connections, 0) != 0 ? backend.max_connections : null
      max_connections_per_instance = try(backend.max_connections_per_instance, 0) != 0 ? backend.max_connections_per_instance : null
      max_connections_per_endpoint = try(backend.max_connections_per_endpoint, 0) != 0 ? backend.max_connections_per_endpoint : null
      max_rate                     = try(backend.max_rate, 0) != 0 ? backend.max_rate : null
      max_rate_per_instance        = try(backend.max_rate_per_instance, 0) != 0 ? backend.max_rate_per_instance : null
      max_rate_per_endpoint        = try(backend.max_rate_per_endpoint, 0) != 0 ? backend.max_rate_per_endpoint : null
      max_utilization              = try(backend.max_utilization, 0) != 0 ? backend.max_utilization : null
      preference                   = backend.preference != "" ? backend.preference : null
      custom_metrics               = try(backend.custom_metrics, [])
      # Regional-only (the failover pool of a passthrough NLB); the API
      # default is false, so only an explicit true is sent.
      failover = try(backend.failover, false) ? true : null
    }
  ]

  # The tfvars converter emits 0 for unset proto ints, but a real 0 TTL is
  # also meaningful to GCP ("don't cache"). The spec resolves the ambiguity
  # in favor of the overwhelmingly common intent: 0 = unset, letting the API
  # apply its own defaults. cache_mode governs which TTLs may be sent at all
  # (GCP rejects TTLs it would ignore) — the spec's CEL rules enforce that
  # before deploy, so no TTL-stripping logic is needed here.
  # request_coalescing is sent only when enabled: the API defaults coalescing
  # ON, so an explicit false would switch it off for a manifest that never
  # mentioned it.
  cdn_policy = var.spec.cdn_policy == null ? null : {
    cache_mode                      = try(var.spec.cdn_policy.cache_mode, "") != "" ? var.spec.cdn_policy.cache_mode : null
    client_ttl                      = try(var.spec.cdn_policy.client_ttl, 0) != 0 ? var.spec.cdn_policy.client_ttl : null
    default_ttl                     = try(var.spec.cdn_policy.default_ttl, 0) != 0 ? var.spec.cdn_policy.default_ttl : null
    max_ttl                         = try(var.spec.cdn_policy.max_ttl, 0) != 0 ? var.spec.cdn_policy.max_ttl : null
    negative_caching                = try(var.spec.cdn_policy.negative_caching, null)
    negative_caching_policy         = try(var.spec.cdn_policy.negative_caching_policy, [])
    serve_while_stale               = try(var.spec.cdn_policy.serve_while_stale, 0) != 0 ? var.spec.cdn_policy.serve_while_stale : null
    request_coalescing              = try(var.spec.cdn_policy.request_coalescing, false) ? true : null
    signed_url_cache_max_age_sec    = try(var.spec.cdn_policy.signed_url_cache_max_age_sec, 0) != 0 ? var.spec.cdn_policy.signed_url_cache_max_age_sec : null
    cache_key_policy                = try(var.spec.cdn_policy.cache_key_policy, null)
    bypass_cache_on_request_headers = try(var.spec.cdn_policy.bypass_cache_on_request_headers, [])
  }

  # IAP: only send the block when enabled or a custom client is configured.
  # The provider requires `enabled`; an all-empty block would needlessly
  # write `enabled = false` state.
  iap = var.spec.iap == null ? null : {
    enabled              = coalesce(var.spec.iap.enabled, false)
    oauth2_client_id     = try(var.spec.iap.oauth2_client_id, "") != "" ? var.spec.iap.oauth2_client_id : null
    oauth2_client_secret = try(var.spec.iap.oauth2_client_secret, "") != "" ? var.spec.iap.oauth2_client_secret : null
  }

  # Log config: sample_rate 0 -> null (unset; API default 1.0). A genuine
  # "log nothing" is enable = false.
  log_config = var.spec.log_config == null ? null : {
    enable          = coalesce(var.spec.log_config.enable, false)
    sample_rate     = try(var.spec.log_config.sample_rate, 0) != 0 ? var.spec.log_config.sample_rate : null
    optional_mode   = try(var.spec.log_config.optional_mode, "") != "" ? var.spec.log_config.optional_mode : null
    optional_fields = try(var.spec.log_config.optional_fields, [])
    # Header names whose values join each log entry (HTTP-family protocols).
    request_headers  = try(var.spec.log_config.request_headers, [])
    response_headers = try(var.spec.log_config.response_headers, [])
  }

  # Circuit breakers: 0 -> null so the API defaults (1024/1024/1024/-/3)
  # apply per-field.
  circuit_breakers = var.spec.circuit_breakers == null ? null : {
    max_connections             = try(var.spec.circuit_breakers.max_connections, 0) != 0 ? var.spec.circuit_breakers.max_connections : null
    max_pending_requests        = try(var.spec.circuit_breakers.max_pending_requests, 0) != 0 ? var.spec.circuit_breakers.max_pending_requests : null
    max_requests                = try(var.spec.circuit_breakers.max_requests, 0) != 0 ? var.spec.circuit_breakers.max_requests : null
    max_requests_per_connection = try(var.spec.circuit_breakers.max_requests_per_connection, 0) != 0 ? var.spec.circuit_breakers.max_requests_per_connection : null
    max_retries                 = try(var.spec.circuit_breakers.max_retries, 0) != 0 ? var.spec.circuit_breakers.max_retries : null
  }

  # Outlier detection: 0 -> null per field so GCP's own defaults apply;
  # duration sub-blocks pass through when present.
  outlier_detection = var.spec.outlier_detection == null ? null : {
    base_ejection_time                    = try(var.spec.outlier_detection.base_ejection_time, null)
    consecutive_errors                    = try(var.spec.outlier_detection.consecutive_errors, 0) != 0 ? var.spec.outlier_detection.consecutive_errors : null
    consecutive_gateway_failure           = try(var.spec.outlier_detection.consecutive_gateway_failure, 0) != 0 ? var.spec.outlier_detection.consecutive_gateway_failure : null
    enforcing_consecutive_errors          = try(var.spec.outlier_detection.enforcing_consecutive_errors, 0) != 0 ? var.spec.outlier_detection.enforcing_consecutive_errors : null
    enforcing_consecutive_gateway_failure = try(var.spec.outlier_detection.enforcing_consecutive_gateway_failure, 0) != 0 ? var.spec.outlier_detection.enforcing_consecutive_gateway_failure : null
    enforcing_success_rate                = try(var.spec.outlier_detection.enforcing_success_rate, 0) != 0 ? var.spec.outlier_detection.enforcing_success_rate : null
    interval                              = try(var.spec.outlier_detection.interval, null)
    max_ejection_percent                  = try(var.spec.outlier_detection.max_ejection_percent, 0) != 0 ? var.spec.outlier_detection.max_ejection_percent : null
    success_rate_minimum_hosts            = try(var.spec.outlier_detection.success_rate_minimum_hosts, 0) != 0 ? var.spec.outlier_detection.success_rate_minimum_hosts : null
    success_rate_request_volume           = try(var.spec.outlier_detection.success_rate_request_volume, 0) != 0 ? var.spec.outlier_detection.success_rate_request_volume : null
    success_rate_stdev_factor             = try(var.spec.outlier_detection.success_rate_stdev_factor, 0) != 0 ? var.spec.outlier_detection.success_rate_stdev_factor : null
  }

  # Consistent hash: "" -> null strings, 0 -> null ring size.
  consistent_hash = var.spec.consistent_hash == null ? null : {
    http_header_name  = try(var.spec.consistent_hash.http_header_name, "") != "" ? var.spec.consistent_hash.http_header_name : null
    minimum_ring_size = try(var.spec.consistent_hash.minimum_ring_size, 0) != 0 ? var.spec.consistent_hash.minimum_ring_size : null
    http_cookie       = try(var.spec.consistent_hash.http_cookie, null)
  }

  # Strong-affinity cookie: GCP requires a cookie configuration with
  # STRONG_COOKIE_AFFINITY; the mode is the statement and the spec block only
  # customizes it, so the block is rendered (GCP-defaulted when the spec has
  # none) whenever that mode is chosen. "" name/path -> null.
  strong_session_affinity_cookie = var.spec.strong_session_affinity_cookie == null && var.spec.session_affinity != "STRONG_COOKIE_AFFINITY" ? null : {
    name = try(var.spec.strong_session_affinity_cookie.name, "") != "" ? var.spec.strong_session_affinity_cookie.name : null
    path = try(var.spec.strong_session_affinity_cookie.path, "") != "" ? var.spec.strong_session_affinity_cookie.path : null
    ttl  = try(var.spec.strong_session_affinity_cookie.ttl, null)
  }

  # Security settings: only send sub-fields that are set; SigV4 block passes
  # through with "" -> null normalization.
  security_settings = var.spec.security_settings == null ? null : {
    client_tls_policy = try(var.spec.security_settings.client_tls_policy, "") != "" ? var.spec.security_settings.client_tls_policy : null
    subject_alt_names = try(var.spec.security_settings.subject_alt_names, [])
    aws_v4_authentication = var.spec.security_settings.aws_v4_authentication == null ? null : {
      access_key_id      = try(var.spec.security_settings.aws_v4_authentication.access_key_id, "") != "" ? var.spec.security_settings.aws_v4_authentication.access_key_id : null
      access_key         = try(var.spec.security_settings.aws_v4_authentication.access_key, "") != "" ? var.spec.security_settings.aws_v4_authentication.access_key : null
      access_key_version = try(var.spec.security_settings.aws_v4_authentication.access_key_version, "") != "" ? var.spec.security_settings.aws_v4_authentication.access_key_version : null
      origin_region      = try(var.spec.security_settings.aws_v4_authentication.origin_region, "") != "" ? var.spec.security_settings.aws_v4_authentication.origin_region : null
    }
  }

  # TLS settings: "" -> null; SAN entries keep exactly one arm (the proto
  # oneof guarantees the other is empty — map "" -> null so the provider
  # sees only the set arm).
  tls_settings = var.spec.tls_settings == null ? null : {
    authentication_config = try(var.spec.tls_settings.authentication_config, "") != "" ? var.spec.tls_settings.authentication_config : null
    sni                   = try(var.spec.tls_settings.sni, "") != "" ? var.spec.tls_settings.sni : null
    subject_alt_names = [
      for san in try(var.spec.tls_settings.subject_alt_names, []) : {
        dns_name                    = san.dns_name != "" ? san.dns_name : null
        uniform_resource_identifier = san.uniform_resource_identifier != "" ? san.uniform_resource_identifier : null
      }
    ]
  }

  # max_stream_duration: the provider takes seconds as a STRING (int64
  # format) — convert the spec's number; nanos stays a number.
  max_stream_duration = var.spec.max_stream_duration == null ? null : {
    seconds = tostring(coalesce(var.spec.max_stream_duration.seconds, 0))
    nanos   = try(var.spec.max_stream_duration.nanos, 0) != 0 ? var.spec.max_stream_duration.nanos : null
  }

  # ---- Regional-only policies (the passthrough Network Load Balancers) ----
  # Every field below carries presence in the spec (optional), so null means
  # "not set" and the API default stands; the spec's CEL keeps these blocks
  # off a global manifest, whose resource has no such arguments.

  # Failover policy: the three fields pass through with their presence
  # intact (Google requires at least one; the spec enforces it).
  failover_policy = var.spec.failover_policy == null ? null : {
    disable_connection_drain_on_failover = try(var.spec.failover_policy.disable_connection_drain_on_failover, null)
    drop_traffic_if_unhealthy            = try(var.spec.failover_policy.drop_traffic_if_unhealthy, null)
    failover_ratio                       = try(var.spec.failover_policy.failover_ratio, null)
  }

  # Connection tracking: the two mode strings carry spec defaults that
  # match Google's, so they are sent explicitly (a re-plan stays clean);
  # idle_timeout_sec is Optional+Computed on the API and sent only when set.
  connection_tracking_policy = var.spec.connection_tracking_policy == null ? null : {
    tracking_mode                                = try(var.spec.connection_tracking_policy.tracking_mode, "") != "" ? var.spec.connection_tracking_policy.tracking_mode : "PER_CONNECTION"
    connection_persistence_on_unhealthy_backends = try(var.spec.connection_tracking_policy.connection_persistence_on_unhealthy_backends, "") != "" ? var.spec.connection_tracking_policy.connection_persistence_on_unhealthy_backends : "DEFAULT_FOR_PROTOCOL"
    idle_timeout_sec                             = try(var.spec.connection_tracking_policy.idle_timeout_sec, null)
    enable_strong_affinity                       = try(var.spec.connection_tracking_policy.enable_strong_affinity, false) ? true : null
  }

  # HA policy: "" -> null on the strings; the leader block renders only when
  # declared.
  ha_policy = var.spec.ha_policy == null ? null : {
    fast_ip_move = try(var.spec.ha_policy.fast_ip_move, "") != "" ? var.spec.ha_policy.fast_ip_move : null
    leader = var.spec.ha_policy.leader == null ? null : {
      backend_group = try(var.spec.ha_policy.leader.backend_group, "") != "" ? var.spec.ha_policy.leader.backend_group : null
      instance      = try(var.spec.ha_policy.leader.network_endpoint.instance, "") != "" ? var.spec.ha_policy.leader.network_endpoint.instance : null
    }
  }

  # Zonal affinity: the mode carries a spec default matching Google's and
  # is sent explicitly; the ratio is sent only when set.
  zonal_affinity = var.spec.network_pass_through_lb_traffic_policy == null || var.spec.network_pass_through_lb_traffic_policy.zonal_affinity == null ? null : {
    spillover       = try(var.spec.network_pass_through_lb_traffic_policy.zonal_affinity.spillover, "") != "" ? var.spec.network_pass_through_lb_traffic_policy.zonal_affinity.spillover : "ZONAL_AFFINITY_DISABLED"
    spillover_ratio = try(var.spec.network_pass_through_lb_traffic_policy.zonal_affinity.spillover_ratio, null)
  }
}
