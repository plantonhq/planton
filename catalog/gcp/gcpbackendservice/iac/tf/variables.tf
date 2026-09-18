variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the GCP Compute Engine global backend service"
  type = object({
    # The GCP project that owns the backend service. The CLI's tfvars
    # converter resolves StringValueOrRef fields to their literal string
    # before the module runs, so this arrives as a plain string.
    # If empty, the provider's default project is used (see locals.tf).
    project_id = optional(string, "")

    # Name of the backend service in GCP (RFC1035). Empty defaults to
    # metadata.name (see locals.tf). Immutable (ForceNew).
    backend_service_name = optional(string, "")

    description = optional(string)

    # LB→backend protocol. Planton middleware applies the proto default
    # (HTTP); empty falls through to the same GCP API default.
    protocol = optional(string, "")

    # Which load balancer family this service serves. Middleware applies the
    # proto default (EXTERNAL); empty falls through to the same API default.
    load_balancing_scheme = optional(string, "")

    # Named port on the instance groups (EXTERNAL scheme + instance-group
    # backends).
    port_name = optional(string, "")

    # Backend response timeout and connection draining. Middleware applies
    # the proto defaults (30/300); null falls through to the API defaults.
    timeout_sec                     = optional(number)
    connection_draining_timeout_sec = optional(number)

    # Self-link of the health check (resolved from a GcpHealthCheck
    # reference or given directly). Empty means no health check — valid
    # only when every backend is an internet or serverless NEG.
    health_check = optional(string, "")

    # The backends serving traffic. group arrives as a plain string
    # (resolved reference or literal instance-group/NEG self-link).
    backends = optional(list(object({
      group                        = string
      balancing_mode               = optional(string, "")
      capacity_scaler              = optional(number)
      description                  = optional(string)
      max_connections              = optional(number)
      max_connections_per_instance = optional(number)
      max_connections_per_endpoint = optional(number)
      max_rate                     = optional(number)
      max_rate_per_instance        = optional(number)
      max_rate_per_endpoint        = optional(number)
      max_utilization              = optional(number)
      preference                   = optional(string, "")
      custom_metrics = optional(list(object({
        name            = string
        dry_run         = optional(bool, false)
        max_utilization = optional(number)
      })), [])
    })), [])

    # Session stickiness. Middleware applies the proto default (NONE);
    # empty falls through to the same API default.
    session_affinity        = optional(string, "")
    affinity_cookie_ttl_sec = optional(number)

    # Cookie for STRONG_COOKIE_AFFINITY (required with that mode — enforced
    # by the spec's CEL before deploy).
    strong_session_affinity_cookie = optional(object({
      name = optional(string, "")
      path = optional(string, "")
      ttl = optional(object({
        seconds = optional(number)
        nanos   = optional(number)
      }))
    }))

    # Within-group balancing algorithm and the ordered Traffic Director
    # policy list (custom xDS policies with fallbacks).
    locality_lb_policy = optional(string, "")
    locality_lb_policies = optional(list(object({
      policy        = optional(object({ name = string }))
      custom_policy = optional(object({ name = string, data = optional(string, "") }))
    })), [])

    # Consistent-hash parameters (INTERNAL_SELF_MANAGED + MAGLEV/RING_HASH
    # only — enforced by the spec's CEL before deploy).
    consistent_hash = optional(object({
      http_header_name  = optional(string, "")
      minimum_ring_size = optional(number)
      http_cookie = optional(object({
        name = optional(string, "")
        path = optional(string, "")
        ttl = optional(object({
          seconds = optional(number)
          nanos   = optional(number)
        }))
      }))
    }))

    # Cache at Google's edge with Cloud CDN. cdn_policy only takes effect
    # while this is true.
    enable_cdn = optional(bool, false)

    # Cloud CDN caching behavior. TTL fields left at 0 are treated as unset
    # so the GCP API applies its own defaults (see locals.tf).
    cdn_policy = optional(object({
      cache_mode                   = optional(string, "")
      client_ttl                   = optional(number)
      default_ttl                  = optional(number)
      max_ttl                      = optional(number)
      negative_caching             = optional(bool)
      negative_caching_policy      = optional(list(object({ code = number, ttl = optional(number) })), [])
      serve_while_stale            = optional(number)
      request_coalescing           = optional(bool)
      signed_url_cache_max_age_sec = optional(number)
      cache_key_policy = optional(object({
        include_host           = optional(bool, false)
        include_protocol       = optional(bool, false)
        include_query_string   = optional(bool, false)
        query_string_whitelist = optional(list(string), [])
        query_string_blacklist = optional(list(string), [])
        include_http_headers   = optional(list(string), [])
        include_named_cookies  = optional(list(string), [])
      }))
      bypass_cache_on_request_headers = optional(list(object({ header_name = string })), [])
    }))

    # Self-links of Cloud Armor policies (resolved from GcpCloudArmorPolicy
    # references or given directly): security_policy filters after the CDN
    # cache, edge_security_policy before it.
    security_policy      = optional(string, "")
    edge_security_policy = optional(string, "")

    # Identity-Aware Proxy. oauth2_client_secret is secret material — it
    # never appears in outputs.
    iap = optional(object({
      enabled              = optional(bool, false)
      oauth2_client_id     = optional(string, "")
      oauth2_client_secret = optional(string, "")
    }))

    # Request logging to Cloud Logging.
    log_config = optional(object({
      enable          = optional(bool, false)
      sample_rate     = optional(number)
      optional_mode   = optional(string, "")
      optional_fields = optional(list(string), [])
      # Header names whose values join each log entry (enable + HTTP-family
      # protocol required).
      request_headers  = optional(list(string), [])
      response_headers = optional(list(string), [])
    }))

    # Headers the load balancer adds, "Header-Name: value" form.
    custom_request_headers  = optional(list(string), [])
    custom_response_headers = optional(list(string), [])

    # Load-balancer response compression: AUTOMATIC or DISABLED (empty keeps
    # the GCP default of no compression).
    compression_mode = optional(string, "")

    # Traffic Director connection-volume circuit breakers
    # (INTERNAL_SELF_MANAGED only — enforced by the spec's CEL).
    circuit_breakers = optional(object({
      max_connections             = optional(number)
      max_pending_requests        = optional(number)
      max_requests                = optional(number)
      max_requests_per_connection = optional(number)
      max_retries                 = optional(number)
    }))

    # Passive health checking (INTERNAL_SELF_MANAGED / EXTERNAL_MANAGED
    # only — enforced by the spec's CEL). Zero-valued fields are treated as
    # unset so the API applies its own defaults (see locals.tf).
    outlier_detection = optional(object({
      base_ejection_time = optional(object({
        seconds = optional(number)
        nanos   = optional(number)
      }))
      consecutive_errors                    = optional(number)
      consecutive_gateway_failure           = optional(number)
      enforcing_consecutive_errors          = optional(number)
      enforcing_consecutive_gateway_failure = optional(number)
      enforcing_success_rate                = optional(number)
      interval = optional(object({
        seconds = optional(number)
        nanos   = optional(number)
      }))
      max_ejection_percent        = optional(number)
      success_rate_minimum_hosts  = optional(number)
      success_rate_request_volume = optional(number)
      success_rate_stdev_factor   = optional(number)
    }))

    # Default stream timeout (INTERNAL_SELF_MANAGED only — enforced by the
    # spec's CEL). The provider takes seconds as a string (int64 format).
    max_stream_duration = optional(object({
      seconds = optional(number)
      nanos   = optional(number)
    }))

    # Traffic Director mTLS + AWS SigV4 origin authentication. access_key is
    # secret material — it never appears in outputs.
    security_settings = optional(object({
      client_tls_policy = optional(string, "")
      subject_alt_names = optional(list(string), [])
      aws_v4_authentication = optional(object({
        access_key_id      = optional(string, "")
        access_key         = optional(string, "")
        access_key_version = optional(string, "")
        origin_region      = optional(string, "")
      }))
    }))

    # TLS parameters toward the backends (protocol SSL/HTTPS/HTTP2 only —
    # enforced by the spec's CEL). Each SAN entry sets exactly one of
    # dns_name / uniform_resource_identifier (proto oneof upstream).
    tls_settings = optional(object({
      authentication_config = optional(string, "")
      sni                   = optional(string, "")
      subject_alt_names = optional(list(object({
        dns_name                    = optional(string, "")
        uniform_resource_identifier = optional(string, "")
      })), [])
    }))

    # IPv4/IPv6 preference toward dual-stack backends.
    ip_address_selection_policy = optional(string, "")

    # EXTERNAL → EXTERNAL_MANAGED canary migration controls.
    external_managed_migration_state              = optional(string, "")
    external_managed_migration_testing_percentage = optional(number)

    # Service-level ORCA metrics for WEIGHTED_ROUND_ROBIN.
    custom_metrics = optional(list(object({
      name    = string
      dry_run = optional(bool, false)
    })), [])

    # Self-link of a networkservices ServiceLbPolicy (plain URL).
    service_lb_policy = optional(string, "")

    # Cloud CDN signed-URL keys (at most 3). Each key_value is secret
    # material — it never appears in outputs.
    signed_url_keys = optional(list(object({
      name      = string
      key_value = string
    })), [])

    # Resource Manager tags bound at create time (tagKeys/{id} ->
    # tagValues/{id}). Immutable.
    resource_manager_tags = optional(map(string), {})

    # DELETE (default), PREVENT, or ABANDON — one switch governing destroy
    # for the backend service AND its signed-URL keys.
    deletion_policy = optional(string, "")
  })

  # NOTE: never guard optional strings with coalesce() here — HCL's coalesce
  # skips empty strings as well as nulls, so coalesce("", "") errors and the
  # validation fails on a legitimately-empty value.
  validation {
    condition     = try(var.spec.backend_service_name, "") == "" || can(regex("^[a-z]([-a-z0-9]{0,61}[a-z0-9])?$", var.spec.backend_service_name))
    error_message = "backend_service_name must be RFC1035-compliant: 1-63 lowercase letters, digits, or hyphens."
  }

  validation {
    condition     = contains(["", "HTTP", "HTTPS", "HTTP2", "H2C", "TCP", "SSL", "UDP", "GRPC"], var.spec.protocol)
    error_message = "protocol must be one of HTTP, HTTPS, HTTP2, H2C, TCP, SSL, UDP, or GRPC."
  }

  validation {
    condition     = contains(["", "EXTERNAL", "EXTERNAL_MANAGED", "INTERNAL_MANAGED", "INTERNAL_SELF_MANAGED"], var.spec.load_balancing_scheme)
    error_message = "load_balancing_scheme must be one of EXTERNAL, EXTERNAL_MANAGED, INTERNAL_MANAGED, or INTERNAL_SELF_MANAGED."
  }

  validation {
    condition     = contains(["", "AUTOMATIC", "DISABLED"], var.spec.compression_mode)
    error_message = "compression_mode must be AUTOMATIC or DISABLED."
  }

  # HCL's && does not short-circuit, so the nullable bool is guarded with
  # coalesce — Cloud CDN only fronts external load balancers.
  validation {
    condition     = !(coalesce(var.spec.enable_cdn, false) && contains(["INTERNAL_MANAGED", "INTERNAL_SELF_MANAGED"], var.spec.load_balancing_scheme))
    error_message = "Cloud CDN can only be enabled on external backend services (scheme EXTERNAL or EXTERNAL_MANAGED)."
  }

  validation {
    condition     = alltrue([for backend in var.spec.backends : length(backend.group) > 0])
    error_message = "every backend must set group — the instance-group or NEG self-link."
  }

  validation {
    condition     = length(var.spec.signed_url_keys) <= 3
    error_message = "at most 3 signed-URL keys are supported per backend service."
  }
}
