# Enable the Compute Engine API so a fresh project can host the backend
# service. disable_on_destroy is false: tearing down one backend service must
# never disable the API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A global Compute Engine backend service — the hub of the L7 load balancing
# family. It owns how traffic reaches a set of backends: the backend list,
# health checking, session affinity, Cloud CDN policy, IAP, Cloud Armor
# attachment, and logging. URL maps route host/path patterns here.
#
# name and project are immutable (ForceNew): changing either destroys and
# recreates the backend service, briefly breaking every URL map referencing
# the old self_link. Everything else — backends, CDN policy, affinity, IAP —
# updates in place, which is what makes this node the operational lever of a
# running load balancer.
#
# Two provider subtleties worth knowing when comparing plans to API calls:
# the provider applies security_policy and edge_security_policy via
# dedicated setSecurityPolicy/setEdgeSecurityPolicy sub-calls (not the main
# insert/patch body), and it strips max_utilization from any NEG backend
# because the API rejects it there. Neither changes declared state — both
# engines behave identically.
#
# Cross-field applicability (CDN only on external schemes, circuit breakers /
# max_stream_duration only on INTERNAL_SELF_MANAGED, consistent_hash only
# with MAGLEV/RING_HASH, cache-mode/TTL coherence) is enforced by the spec's
# CEL rules before deploy, so this module stays declarative.
resource "google_compute_backend_service" "this" {
  name        = local.backend_service_name
  project     = local.project_id
  description = var.spec.description

  protocol              = local.protocol
  load_balancing_scheme = local.load_balancing_scheme
  port_name             = local.port_name

  timeout_sec                     = local.timeout_sec
  connection_draining_timeout_sec = local.connection_draining_timeout_sec

  # At most one health check (a GCP-enforced cap — the spec models it
  # singular); none at all is valid only for serverless/internet NEG
  # backends.
  health_checks = local.health_checks

  session_affinity        = local.session_affinity
  affinity_cookie_ttl_sec = local.affinity_cookie_ttl_sec

  locality_lb_policy = local.locality_lb_policy

  enable_cdn       = var.spec.enable_cdn
  compression_mode = local.compression_mode

  # Attach by reference — the Cloud Armor policies are their own composable
  # nodes (GcpCloudArmorPolicy), never embedded here. security_policy
  # filters after the CDN cache; edge_security_policy before it.
  security_policy      = local.security_policy
  edge_security_policy = local.edge_security_policy

  custom_request_headers  = length(var.spec.custom_request_headers) > 0 ? var.spec.custom_request_headers : null
  custom_response_headers = length(var.spec.custom_response_headers) > 0 ? var.spec.custom_response_headers : null

  ip_address_selection_policy = local.ip_address_selection_policy
  service_lb_policy           = local.service_lb_policy

  external_managed_migration_state              = local.migration_state
  external_managed_migration_testing_percentage = local.migration_testing_percentage

  # The backends serving traffic. Instance groups and NEGs cannot be mixed
  # on one service (GCP rejects it); which max_* dial is required is decided
  # by each backend's balancing_mode and enforced by the spec pre-deploy.
  dynamic "backend" {
    for_each = local.backends
    content {
      group                        = backend.value.group
      balancing_mode               = backend.value.balancing_mode
      capacity_scaler              = backend.value.capacity_scaler
      description                  = backend.value.description
      max_connections              = backend.value.max_connections
      max_connections_per_instance = backend.value.max_connections_per_instance
      max_connections_per_endpoint = backend.value.max_connections_per_endpoint
      max_rate                     = backend.value.max_rate
      max_rate_per_instance        = backend.value.max_rate_per_instance
      max_rate_per_endpoint        = backend.value.max_rate_per_endpoint
      max_utilization              = backend.value.max_utilization
      preference                   = backend.value.preference

      dynamic "custom_metrics" {
        for_each = backend.value.custom_metrics
        content {
          name            = custom_metrics.value.name
          dry_run         = custom_metrics.value.dry_run
          max_utilization = custom_metrics.value.max_utilization
        }
      }
    }
  }

  dynamic "cdn_policy" {
    for_each = local.cdn_policy != null ? [local.cdn_policy] : []
    content {
      cache_mode                   = cdn_policy.value.cache_mode
      client_ttl                   = cdn_policy.value.client_ttl
      default_ttl                  = cdn_policy.value.default_ttl
      max_ttl                      = cdn_policy.value.max_ttl
      negative_caching             = cdn_policy.value.negative_caching
      serve_while_stale            = cdn_policy.value.serve_while_stale
      request_coalescing           = cdn_policy.value.request_coalescing
      signed_url_cache_max_age_sec = cdn_policy.value.signed_url_cache_max_age_sec

      dynamic "negative_caching_policy" {
        for_each = cdn_policy.value.negative_caching_policy
        content {
          code = negative_caching_policy.value.code
          # A 0 TTL is meaningful here (GCP treats 0 as don't-cache-this-
          # code), so pass it as-is.
          ttl = negative_caching_policy.value.ttl
        }
      }

      # The backend-service cache key is richer than a backend bucket's:
      # host, protocol, query handling, cookies, and headers all shape it.
      dynamic "cache_key_policy" {
        for_each = cdn_policy.value.cache_key_policy != null ? [cdn_policy.value.cache_key_policy] : []
        content {
          include_host           = cache_key_policy.value.include_host
          include_protocol       = cache_key_policy.value.include_protocol
          include_query_string   = cache_key_policy.value.include_query_string
          query_string_whitelist = cache_key_policy.value.query_string_whitelist
          query_string_blacklist = cache_key_policy.value.query_string_blacklist
          include_http_headers   = cache_key_policy.value.include_http_headers
          include_named_cookies  = cache_key_policy.value.include_named_cookies
        }
      }

      dynamic "bypass_cache_on_request_headers" {
        for_each = cdn_policy.value.bypass_cache_on_request_headers
        content {
          header_name = bypass_cache_on_request_headers.value.header_name
        }
      }
    }
  }

  # Identity-Aware Proxy: Google-identity authentication in front of the
  # backends. The client secret is secret material — never surfaced in
  # outputs; GCP itself only returns its SHA-256 after creation.
  dynamic "iap" {
    for_each = local.iap != null ? [local.iap] : []
    content {
      enabled              = iap.value.enabled
      oauth2_client_id     = iap.value.oauth2_client_id
      oauth2_client_secret = iap.value.oauth2_client_secret
    }
  }

  dynamic "log_config" {
    for_each = local.log_config != null ? [local.log_config] : []
    content {
      enable          = log_config.value.enable
      sample_rate     = log_config.value.sample_rate
      optional_mode   = log_config.value.optional_mode
      optional_fields = log_config.value.optional_fields

      # Each logged header is its own block in the provider; the spec
      # holds a flat name list.
      dynamic "request_headers" {
        for_each = log_config.value.request_headers
        content {
          header_name = request_headers.value
        }
      }
      dynamic "response_headers" {
        for_each = log_config.value.response_headers
        content {
          header_name = response_headers.value
        }
      }
    }
  }

  dynamic "strong_session_affinity_cookie" {
    for_each = local.strong_session_affinity_cookie != null ? [local.strong_session_affinity_cookie] : []
    content {
      name = strong_session_affinity_cookie.value.name
      path = strong_session_affinity_cookie.value.path

      dynamic "ttl" {
        for_each = strong_session_affinity_cookie.value.ttl != null ? [strong_session_affinity_cookie.value.ttl] : []
        content {
          seconds = coalesce(ttl.value.seconds, 0)
          nanos   = try(ttl.value.nanos, 0) != 0 ? ttl.value.nanos : null
        }
      }
    }
  }

  # Ordered Traffic Director policy list; each entry carries exactly one of
  # a built-in policy or a custom xDS policy (proto oneof upstream).
  dynamic "locality_lb_policies" {
    for_each = var.spec.locality_lb_policies
    content {
      dynamic "policy" {
        for_each = locality_lb_policies.value.policy != null ? [locality_lb_policies.value.policy] : []
        content {
          name = policy.value.name
        }
      }
      dynamic "custom_policy" {
        for_each = locality_lb_policies.value.custom_policy != null ? [locality_lb_policies.value.custom_policy] : []
        content {
          name = custom_policy.value.name
          data = custom_policy.value.data != "" ? custom_policy.value.data : null
        }
      }
    }
  }

  dynamic "consistent_hash" {
    for_each = local.consistent_hash != null ? [local.consistent_hash] : []
    content {
      http_header_name  = consistent_hash.value.http_header_name
      minimum_ring_size = consistent_hash.value.minimum_ring_size

      dynamic "http_cookie" {
        for_each = consistent_hash.value.http_cookie != null ? [consistent_hash.value.http_cookie] : []
        content {
          name = try(http_cookie.value.name, "") != "" ? http_cookie.value.name : null
          path = try(http_cookie.value.path, "") != "" ? http_cookie.value.path : null

          dynamic "ttl" {
            for_each = http_cookie.value.ttl != null ? [http_cookie.value.ttl] : []
            content {
              seconds = coalesce(ttl.value.seconds, 0)
              nanos   = try(ttl.value.nanos, 0) != 0 ? ttl.value.nanos : null
            }
          }
        }
      }
    }
  }

  dynamic "circuit_breakers" {
    for_each = local.circuit_breakers != null ? [local.circuit_breakers] : []
    content {
      max_connections             = circuit_breakers.value.max_connections
      max_pending_requests        = circuit_breakers.value.max_pending_requests
      max_requests                = circuit_breakers.value.max_requests
      max_requests_per_connection = circuit_breakers.value.max_requests_per_connection
      max_retries                 = circuit_breakers.value.max_retries
    }
  }

  dynamic "outlier_detection" {
    for_each = local.outlier_detection != null ? [local.outlier_detection] : []
    content {
      consecutive_errors                    = outlier_detection.value.consecutive_errors
      consecutive_gateway_failure           = outlier_detection.value.consecutive_gateway_failure
      enforcing_consecutive_errors          = outlier_detection.value.enforcing_consecutive_errors
      enforcing_consecutive_gateway_failure = outlier_detection.value.enforcing_consecutive_gateway_failure
      enforcing_success_rate                = outlier_detection.value.enforcing_success_rate
      max_ejection_percent                  = outlier_detection.value.max_ejection_percent
      success_rate_minimum_hosts            = outlier_detection.value.success_rate_minimum_hosts
      success_rate_request_volume           = outlier_detection.value.success_rate_request_volume
      success_rate_stdev_factor             = outlier_detection.value.success_rate_stdev_factor

      dynamic "base_ejection_time" {
        for_each = outlier_detection.value.base_ejection_time != null ? [outlier_detection.value.base_ejection_time] : []
        content {
          seconds = coalesce(base_ejection_time.value.seconds, 0)
          nanos   = try(base_ejection_time.value.nanos, 0) != 0 ? base_ejection_time.value.nanos : null
        }
      }

      dynamic "interval" {
        for_each = outlier_detection.value.interval != null ? [outlier_detection.value.interval] : []
        content {
          seconds = coalesce(interval.value.seconds, 0)
          nanos   = try(interval.value.nanos, 0) != 0 ? interval.value.nanos : null
        }
      }
    }
  }

  # The provider models Duration seconds as a STRING here (int64 format) —
  # locals.tf converts.
  dynamic "max_stream_duration" {
    for_each = local.max_stream_duration != null ? [local.max_stream_duration] : []
    content {
      seconds = max_stream_duration.value.seconds
      nanos   = max_stream_duration.value.nanos
    }
  }

  dynamic "security_settings" {
    for_each = local.security_settings != null ? [local.security_settings] : []
    content {
      client_tls_policy = security_settings.value.client_tls_policy
      subject_alt_names = security_settings.value.subject_alt_names

      # SigV4 origin signing: access_key is secret material — never
      # surfaced in outputs; GCP never returns it on reads.
      dynamic "aws_v4_authentication" {
        for_each = security_settings.value.aws_v4_authentication != null ? [security_settings.value.aws_v4_authentication] : []
        content {
          access_key_id      = aws_v4_authentication.value.access_key_id
          access_key         = aws_v4_authentication.value.access_key
          access_key_version = aws_v4_authentication.value.access_key_version
          origin_region      = aws_v4_authentication.value.origin_region
        }
      }
    }
  }

  dynamic "tls_settings" {
    for_each = local.tls_settings != null ? [local.tls_settings] : []
    content {
      authentication_config = tls_settings.value.authentication_config
      sni                   = tls_settings.value.sni

      dynamic "subject_alt_names" {
        for_each = tls_settings.value.subject_alt_names
        content {
          dns_name                    = subject_alt_names.value.dns_name
          uniform_resource_identifier = subject_alt_names.value.uniform_resource_identifier
        }
      }
    }
  }

  # Service-level ORCA metrics for WEIGHTED_ROUND_ROBIN.
  dynamic "custom_metrics" {
    for_each = var.spec.custom_metrics
    content {
      name    = custom_metrics.value.name
      dry_run = custom_metrics.value.dry_run
    }
  }

  # Create-time Resource Manager tags; the provider nests them in a params
  # block that must be omitted entirely when no tags are set.
  dynamic "params" {
    for_each = length(var.spec.resource_manager_tags) > 0 ? [1] : []
    content {
      resource_manager_tags = var.spec.resource_manager_tags
    }
  }

  # What destroy does to the routing target: DELETE (default), PREVENT
  # (refuse), or ABANDON (drop from state, keep serving).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  depends_on = [google_project_service.compute_api]
}

# Cloud CDN signed-URL keys — folded into this kind rather than modeled as a
# separate node: keys are never referenced by other resources, GCP caps them
# at 3 per service, and their lifecycle is the service's. Each key is
# immutable in GCP (add/delete only); changing a key's value forces
# replacement of that key resource, which is exactly the rotation semantics
# signed URLs need (add new key -> re-sign -> remove old).
resource "google_compute_backend_service_signed_url_key" "this" {
  for_each = { for signed_url_key in var.spec.signed_url_keys : signed_url_key.name => signed_url_key }

  name            = each.value.name
  key_value       = each.value.key_value # secret material; never surfaced in outputs
  backend_service = google_compute_backend_service.this.name
  project         = local.project_id

  # The kind-level deletion_policy fans out to every key — the keys have no
  # life apart from the service, so one switch governs both objects.
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null
}
