# Enable the Compute Engine API so a fresh project can host the URL map.
# disable_on_destroy is false: tearing down one URL map must never disable the
# API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# A global Compute Engine URL map — the L7 routing brain of a global external
# Application Load Balancer. Host rules map request Host headers to named path
# matchers; path matchers evaluate route_rules (priority-ordered, rich matching)
# then path_rules (longest prefix), then their own default; anything unmatched
# falls through to the URL map's top-level default.
#
# name and project are immutable (ForceNew): changing either destroys and
# recreates the map, briefly breaking every target proxy referencing the old
# self_link. Routing tables, header actions, and tests update in place.
#
# Cross-field exclusivity (one default target, path_rules XOR route_rules,
# redirect vs route_action, path_template_rewrite only in route rules) is
# enforced by the spec's CEL rules before deploy — no defensive logic here.
# route_action carries the full traffic-management surface at every site:
# weighted splits, rewrites, timeout/retry/mirror/CORS/fault-injection/
# stream-duration policies, and the route-scoped CDN cache_policy.
resource "google_compute_url_map" "this" {
  name        = local.url_map_name
  project     = local.project_id
  description = local.description

  # Client-side destroy stance (DELETE/PREVENT/ABANDON) — provider-level,
  # never sent to the GCP API. Empty falls back to the provider default
  # (DELETE).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : null

  default_service = local.default_service

  dynamic "default_url_redirect" {
    for_each = local.default_url_redirect != null ? [local.default_url_redirect] : []
    content {
      host_redirect          = default_url_redirect.value.host_redirect
      https_redirect         = default_url_redirect.value.https_redirect
      path_redirect          = default_url_redirect.value.path_redirect
      prefix_redirect        = default_url_redirect.value.prefix_redirect
      redirect_response_code = default_url_redirect.value.redirect_response_code
      strip_query            = default_url_redirect.value.strip_query
    }
  }

  dynamic "default_route_action" {
    for_each = local.default_route_action != null ? [local.default_route_action] : []
    content {
      dynamic "weighted_backend_services" {
        for_each = default_route_action.value.weighted_backend_services
        content {
          backend_service = weighted_backend_services.value.backend_service
          weight          = weighted_backend_services.value.weight

          dynamic "header_action" {
            for_each = weighted_backend_services.value.header_action != null ? [weighted_backend_services.value.header_action] : []
            content {
              dynamic "request_headers_to_add" {
                for_each = header_action.value.request_headers_to_add
                content {
                  header_name  = request_headers_to_add.value.header_name
                  header_value = request_headers_to_add.value.header_value
                  replace      = request_headers_to_add.value.replace
                }
              }
              request_headers_to_remove = length(header_action.value.request_headers_to_remove) > 0 ? header_action.value.request_headers_to_remove : null
              dynamic "response_headers_to_add" {
                for_each = header_action.value.response_headers_to_add
                content {
                  header_name  = response_headers_to_add.value.header_name
                  header_value = response_headers_to_add.value.header_value
                  replace      = response_headers_to_add.value.replace
                }
              }
              response_headers_to_remove = length(header_action.value.response_headers_to_remove) > 0 ? header_action.value.response_headers_to_remove : null
            }
          }
        }
      }

      dynamic "url_rewrite" {
        for_each = default_route_action.value.url_rewrite != null ? [default_route_action.value.url_rewrite] : []
        content {
          host_rewrite        = url_rewrite.value.host_rewrite
          path_prefix_rewrite = url_rewrite.value.path_prefix_rewrite
        }
      }

      dynamic "timeout" {
        for_each = default_route_action.value.timeout != null ? [default_route_action.value.timeout] : []
        content {
          seconds = coalesce(timeout.value.seconds, 0)
          nanos   = coalesce(timeout.value.nanos, 0) != 0 ? timeout.value.nanos : null
        }
      }

      dynamic "retry_policy" {
        for_each = default_route_action.value.retry_policy != null ? [default_route_action.value.retry_policy] : []
        content {
          num_retries      = retry_policy.value.num_retries != 0 ? retry_policy.value.num_retries : null
          retry_conditions = length(retry_policy.value.retry_conditions) > 0 ? retry_policy.value.retry_conditions : null

          dynamic "per_try_timeout" {
            for_each = retry_policy.value.per_try_timeout != null ? [retry_policy.value.per_try_timeout] : []
            content {
              seconds = coalesce(per_try_timeout.value.seconds, 0)
              nanos   = coalesce(per_try_timeout.value.nanos, 0) != 0 ? per_try_timeout.value.nanos : null
            }
          }
        }
      }

      dynamic "request_mirror_policy" {
        for_each = default_route_action.value.request_mirror_policy != null ? [default_route_action.value.request_mirror_policy] : []
        content {
          backend_service = request_mirror_policy.value.backend_service
        }
      }

      dynamic "cors_policy" {
        for_each = default_route_action.value.cors_policy != null ? [default_route_action.value.cors_policy] : []
        content {
          allow_credentials    = cors_policy.value.allow_credentials
          allow_headers        = length(cors_policy.value.allow_headers) > 0 ? cors_policy.value.allow_headers : null
          allow_methods        = length(cors_policy.value.allow_methods) > 0 ? cors_policy.value.allow_methods : null
          allow_origin_regexes = length(cors_policy.value.allow_origin_regexes) > 0 ? cors_policy.value.allow_origin_regexes : null
          allow_origins        = length(cors_policy.value.allow_origins) > 0 ? cors_policy.value.allow_origins : null
          disabled             = cors_policy.value.disabled
          expose_headers       = length(cors_policy.value.expose_headers) > 0 ? cors_policy.value.expose_headers : null
          max_age              = cors_policy.value.max_age != 0 ? cors_policy.value.max_age : null
        }
      }

      dynamic "fault_injection_policy" {
        for_each = default_route_action.value.fault_injection_policy != null ? [default_route_action.value.fault_injection_policy] : []
        content {
          dynamic "abort" {
            for_each = fault_injection_policy.value.abort != null ? [fault_injection_policy.value.abort] : []
            content {
              http_status = abort.value.http_status != 0 ? abort.value.http_status : null
              percentage  = abort.value.percentage
            }
          }
          dynamic "delay" {
            for_each = fault_injection_policy.value.delay != null ? [fault_injection_policy.value.delay] : []
            content {
              percentage = delay.value.percentage
              dynamic "fixed_delay" {
                for_each = delay.value.fixed_delay != null ? [delay.value.fixed_delay] : []
                content {
                  seconds = coalesce(fixed_delay.value.seconds, 0)
                  nanos   = coalesce(fixed_delay.value.nanos, 0) != 0 ? fixed_delay.value.nanos : null
                }
              }
            }
          }
        }
      }

      dynamic "max_stream_duration" {
        for_each = default_route_action.value.max_stream_duration != null ? [default_route_action.value.max_stream_duration] : []
        content {
          seconds = coalesce(max_stream_duration.value.seconds, 0)
          nanos   = coalesce(max_stream_duration.value.nanos, 0) != 0 ? max_stream_duration.value.nanos : null
        }
      }

      dynamic "cache_policy" {
        for_each = default_route_action.value.cache_policy != null ? [default_route_action.value.cache_policy] : []
        content {
          cache_mode                        = cache_policy.value.cache_mode != "" ? cache_policy.value.cache_mode : null
          cache_bypass_request_header_names = length(cache_policy.value.cache_bypass_request_header_names) > 0 ? cache_policy.value.cache_bypass_request_header_names : null
          negative_caching                  = cache_policy.value.negative_caching
          request_coalescing                = cache_policy.value.request_coalescing

          dynamic "cache_key_policy" {
            for_each = cache_policy.value.cache_key_policy != null ? [cache_policy.value.cache_key_policy] : []
            content {
              excluded_query_parameters = length(cache_key_policy.value.excluded_query_parameters) > 0 ? cache_key_policy.value.excluded_query_parameters : null
              include_host              = cache_key_policy.value.include_host
              include_protocol          = cache_key_policy.value.include_protocol
              include_query_string      = cache_key_policy.value.include_query_string
              included_cookie_names     = length(cache_key_policy.value.included_cookie_names) > 0 ? cache_key_policy.value.included_cookie_names : null
              included_header_names     = length(cache_key_policy.value.included_header_names) > 0 ? cache_key_policy.value.included_header_names : null
              included_query_parameters = length(cache_key_policy.value.included_query_parameters) > 0 ? cache_key_policy.value.included_query_parameters : null
            }
          }

          dynamic "client_ttl" {
            for_each = cache_policy.value.client_ttl != null ? [cache_policy.value.client_ttl] : []
            content {
              seconds = coalesce(client_ttl.value.seconds, 0)
              nanos   = coalesce(client_ttl.value.nanos, 0) != 0 ? client_ttl.value.nanos : null
            }
          }

          dynamic "default_ttl" {
            for_each = cache_policy.value.default_ttl != null ? [cache_policy.value.default_ttl] : []
            content {
              seconds = coalesce(default_ttl.value.seconds, 0)
              nanos   = coalesce(default_ttl.value.nanos, 0) != 0 ? default_ttl.value.nanos : null
            }
          }

          dynamic "max_ttl" {
            for_each = cache_policy.value.max_ttl != null ? [cache_policy.value.max_ttl] : []
            content {
              seconds = coalesce(max_ttl.value.seconds, 0)
              nanos   = coalesce(max_ttl.value.nanos, 0) != 0 ? max_ttl.value.nanos : null
            }
          }

          dynamic "serve_while_stale" {
            for_each = cache_policy.value.serve_while_stale != null ? [cache_policy.value.serve_while_stale] : []
            content {
              seconds = coalesce(serve_while_stale.value.seconds, 0)
              nanos   = coalesce(serve_while_stale.value.nanos, 0) != 0 ? serve_while_stale.value.nanos : null
            }
          }

          dynamic "negative_caching_policy" {
            for_each = cache_policy.value.negative_caching_policy
            content {
              code = negative_caching_policy.value.code != 0 ? negative_caching_policy.value.code : null
              dynamic "ttl" {
                for_each = negative_caching_policy.value.ttl != null ? [negative_caching_policy.value.ttl] : []
                content {
                  seconds = coalesce(ttl.value.seconds, 0)
                  nanos   = coalesce(ttl.value.nanos, 0) != 0 ? ttl.value.nanos : null
                }
              }
            }
          }
        }
      }
    }
  }

  dynamic "default_custom_error_response_policy" {
    for_each = local.default_custom_error_response_policy != null ? [local.default_custom_error_response_policy] : []
    content {
      error_service = default_custom_error_response_policy.value.error_service

      dynamic "error_response_rule" {
        for_each = default_custom_error_response_policy.value.error_response_rules
        content {
          match_response_codes   = error_response_rule.value.match_response_codes
          override_response_code = error_response_rule.value.override_response_code
          path                   = error_response_rule.value.path
        }
      }
    }
  }

  dynamic "header_action" {
    for_each = local.header_action != null ? [local.header_action] : []
    content {
      dynamic "request_headers_to_add" {
        for_each = header_action.value.request_headers_to_add
        content {
          header_name  = request_headers_to_add.value.header_name
          header_value = request_headers_to_add.value.header_value
          replace      = request_headers_to_add.value.replace
        }
      }
      request_headers_to_remove = header_action.value.request_headers_to_remove

      dynamic "response_headers_to_add" {
        for_each = header_action.value.response_headers_to_add
        content {
          header_name  = response_headers_to_add.value.header_name
          header_value = response_headers_to_add.value.header_value
          replace      = response_headers_to_add.value.replace
        }
      }
      response_headers_to_remove = header_action.value.response_headers_to_remove
    }
  }

  dynamic "host_rule" {
    for_each = local.host_rules
    content {
      hosts        = host_rule.value.hosts
      path_matcher = host_rule.value.path_matcher
      description  = host_rule.value.description
    }
  }

  dynamic "path_matcher" {
    for_each = local.path_matchers
    content {
      name            = path_matcher.value.name
      description     = path_matcher.value.description
      default_service = path_matcher.value.default_service

      dynamic "default_url_redirect" {
        for_each = path_matcher.value.default_url_redirect != null ? [path_matcher.value.default_url_redirect] : []
        content {
          host_redirect          = default_url_redirect.value.host_redirect
          https_redirect         = default_url_redirect.value.https_redirect
          path_redirect          = default_url_redirect.value.path_redirect
          prefix_redirect        = default_url_redirect.value.prefix_redirect
          redirect_response_code = default_url_redirect.value.redirect_response_code
          strip_query            = default_url_redirect.value.strip_query
        }
      }

      dynamic "default_route_action" {
        for_each = path_matcher.value.default_route_action != null ? [path_matcher.value.default_route_action] : []
        content {
          dynamic "weighted_backend_services" {
            for_each = default_route_action.value.weighted_backend_services
            content {
              backend_service = weighted_backend_services.value.backend_service
              weight          = weighted_backend_services.value.weight

              dynamic "header_action" {
                for_each = weighted_backend_services.value.header_action != null ? [weighted_backend_services.value.header_action] : []
                content {
                  dynamic "request_headers_to_add" {
                    for_each = header_action.value.request_headers_to_add
                    content {
                      header_name  = request_headers_to_add.value.header_name
                      header_value = request_headers_to_add.value.header_value
                      replace      = request_headers_to_add.value.replace
                    }
                  }
                  request_headers_to_remove = length(header_action.value.request_headers_to_remove) > 0 ? header_action.value.request_headers_to_remove : null
                  dynamic "response_headers_to_add" {
                    for_each = header_action.value.response_headers_to_add
                    content {
                      header_name  = response_headers_to_add.value.header_name
                      header_value = response_headers_to_add.value.header_value
                      replace      = response_headers_to_add.value.replace
                    }
                  }
                  response_headers_to_remove = length(header_action.value.response_headers_to_remove) > 0 ? header_action.value.response_headers_to_remove : null
                }
              }
            }
          }

          dynamic "url_rewrite" {
            for_each = default_route_action.value.url_rewrite != null ? [default_route_action.value.url_rewrite] : []
            content {
              host_rewrite        = url_rewrite.value.host_rewrite
              path_prefix_rewrite = url_rewrite.value.path_prefix_rewrite
            }
          }

          dynamic "timeout" {
            for_each = default_route_action.value.timeout != null ? [default_route_action.value.timeout] : []
            content {
              seconds = coalesce(timeout.value.seconds, 0)
              nanos   = coalesce(timeout.value.nanos, 0) != 0 ? timeout.value.nanos : null
            }
          }

          dynamic "retry_policy" {
            for_each = default_route_action.value.retry_policy != null ? [default_route_action.value.retry_policy] : []
            content {
              num_retries      = retry_policy.value.num_retries != 0 ? retry_policy.value.num_retries : null
              retry_conditions = length(retry_policy.value.retry_conditions) > 0 ? retry_policy.value.retry_conditions : null

              dynamic "per_try_timeout" {
                for_each = retry_policy.value.per_try_timeout != null ? [retry_policy.value.per_try_timeout] : []
                content {
                  seconds = coalesce(per_try_timeout.value.seconds, 0)
                  nanos   = coalesce(per_try_timeout.value.nanos, 0) != 0 ? per_try_timeout.value.nanos : null
                }
              }
            }
          }

          dynamic "request_mirror_policy" {
            for_each = default_route_action.value.request_mirror_policy != null ? [default_route_action.value.request_mirror_policy] : []
            content {
              backend_service = request_mirror_policy.value.backend_service
            }
          }

          dynamic "cors_policy" {
            for_each = default_route_action.value.cors_policy != null ? [default_route_action.value.cors_policy] : []
            content {
              allow_credentials    = cors_policy.value.allow_credentials
              allow_headers        = length(cors_policy.value.allow_headers) > 0 ? cors_policy.value.allow_headers : null
              allow_methods        = length(cors_policy.value.allow_methods) > 0 ? cors_policy.value.allow_methods : null
              allow_origin_regexes = length(cors_policy.value.allow_origin_regexes) > 0 ? cors_policy.value.allow_origin_regexes : null
              allow_origins        = length(cors_policy.value.allow_origins) > 0 ? cors_policy.value.allow_origins : null
              disabled             = cors_policy.value.disabled
              expose_headers       = length(cors_policy.value.expose_headers) > 0 ? cors_policy.value.expose_headers : null
              max_age              = cors_policy.value.max_age != 0 ? cors_policy.value.max_age : null
            }
          }

          dynamic "fault_injection_policy" {
            for_each = default_route_action.value.fault_injection_policy != null ? [default_route_action.value.fault_injection_policy] : []
            content {
              dynamic "abort" {
                for_each = fault_injection_policy.value.abort != null ? [fault_injection_policy.value.abort] : []
                content {
                  http_status = abort.value.http_status != 0 ? abort.value.http_status : null
                  percentage  = abort.value.percentage
                }
              }
              dynamic "delay" {
                for_each = fault_injection_policy.value.delay != null ? [fault_injection_policy.value.delay] : []
                content {
                  percentage = delay.value.percentage
                  dynamic "fixed_delay" {
                    for_each = delay.value.fixed_delay != null ? [delay.value.fixed_delay] : []
                    content {
                      seconds = coalesce(fixed_delay.value.seconds, 0)
                      nanos   = coalesce(fixed_delay.value.nanos, 0) != 0 ? fixed_delay.value.nanos : null
                    }
                  }
                }
              }
            }
          }

          dynamic "max_stream_duration" {
            for_each = default_route_action.value.max_stream_duration != null ? [default_route_action.value.max_stream_duration] : []
            content {
              seconds = coalesce(max_stream_duration.value.seconds, 0)
              nanos   = coalesce(max_stream_duration.value.nanos, 0) != 0 ? max_stream_duration.value.nanos : null
            }
          }

          dynamic "cache_policy" {
            for_each = default_route_action.value.cache_policy != null ? [default_route_action.value.cache_policy] : []
            content {
              cache_mode                        = cache_policy.value.cache_mode != "" ? cache_policy.value.cache_mode : null
              cache_bypass_request_header_names = length(cache_policy.value.cache_bypass_request_header_names) > 0 ? cache_policy.value.cache_bypass_request_header_names : null
              negative_caching                  = cache_policy.value.negative_caching
              request_coalescing                = cache_policy.value.request_coalescing

              dynamic "cache_key_policy" {
                for_each = cache_policy.value.cache_key_policy != null ? [cache_policy.value.cache_key_policy] : []
                content {
                  excluded_query_parameters = length(cache_key_policy.value.excluded_query_parameters) > 0 ? cache_key_policy.value.excluded_query_parameters : null
                  include_host              = cache_key_policy.value.include_host
                  include_protocol          = cache_key_policy.value.include_protocol
                  include_query_string      = cache_key_policy.value.include_query_string
                  included_cookie_names     = length(cache_key_policy.value.included_cookie_names) > 0 ? cache_key_policy.value.included_cookie_names : null
                  included_header_names     = length(cache_key_policy.value.included_header_names) > 0 ? cache_key_policy.value.included_header_names : null
                  included_query_parameters = length(cache_key_policy.value.included_query_parameters) > 0 ? cache_key_policy.value.included_query_parameters : null
                }
              }

              dynamic "client_ttl" {
                for_each = cache_policy.value.client_ttl != null ? [cache_policy.value.client_ttl] : []
                content {
                  seconds = coalesce(client_ttl.value.seconds, 0)
                  nanos   = coalesce(client_ttl.value.nanos, 0) != 0 ? client_ttl.value.nanos : null
                }
              }

              dynamic "default_ttl" {
                for_each = cache_policy.value.default_ttl != null ? [cache_policy.value.default_ttl] : []
                content {
                  seconds = coalesce(default_ttl.value.seconds, 0)
                  nanos   = coalesce(default_ttl.value.nanos, 0) != 0 ? default_ttl.value.nanos : null
                }
              }

              dynamic "max_ttl" {
                for_each = cache_policy.value.max_ttl != null ? [cache_policy.value.max_ttl] : []
                content {
                  seconds = coalesce(max_ttl.value.seconds, 0)
                  nanos   = coalesce(max_ttl.value.nanos, 0) != 0 ? max_ttl.value.nanos : null
                }
              }

              dynamic "serve_while_stale" {
                for_each = cache_policy.value.serve_while_stale != null ? [cache_policy.value.serve_while_stale] : []
                content {
                  seconds = coalesce(serve_while_stale.value.seconds, 0)
                  nanos   = coalesce(serve_while_stale.value.nanos, 0) != 0 ? serve_while_stale.value.nanos : null
                }
              }

              dynamic "negative_caching_policy" {
                for_each = cache_policy.value.negative_caching_policy
                content {
                  code = negative_caching_policy.value.code != 0 ? negative_caching_policy.value.code : null
                  dynamic "ttl" {
                    for_each = negative_caching_policy.value.ttl != null ? [negative_caching_policy.value.ttl] : []
                    content {
                      seconds = coalesce(ttl.value.seconds, 0)
                      nanos   = coalesce(ttl.value.nanos, 0) != 0 ? ttl.value.nanos : null
                    }
                  }
                }
              }
            }
          }
        }
      }

      dynamic "default_custom_error_response_policy" {
        for_each = path_matcher.value.default_custom_error_response_policy != null ? [path_matcher.value.default_custom_error_response_policy] : []
        content {
          error_service = default_custom_error_response_policy.value.error_service

          dynamic "error_response_rule" {
            for_each = default_custom_error_response_policy.value.error_response_rules
            content {
              match_response_codes   = error_response_rule.value.match_response_codes
              override_response_code = error_response_rule.value.override_response_code
              path                   = error_response_rule.value.path
            }
          }
        }
      }

      dynamic "header_action" {
        for_each = path_matcher.value.header_action != null ? [path_matcher.value.header_action] : []
        content {
          dynamic "request_headers_to_add" {
            for_each = header_action.value.request_headers_to_add
            content {
              header_name  = request_headers_to_add.value.header_name
              header_value = request_headers_to_add.value.header_value
              replace      = request_headers_to_add.value.replace
            }
          }
          request_headers_to_remove = header_action.value.request_headers_to_remove

          dynamic "response_headers_to_add" {
            for_each = header_action.value.response_headers_to_add
            content {
              header_name  = response_headers_to_add.value.header_name
              header_value = response_headers_to_add.value.header_value
              replace      = response_headers_to_add.value.replace
            }
          }
          response_headers_to_remove = header_action.value.response_headers_to_remove
        }
      }

      dynamic "path_rule" {
        for_each = path_matcher.value.path_rules
        content {
          paths   = path_rule.value.paths
          service = path_rule.value.service

          dynamic "url_redirect" {
            for_each = path_rule.value.url_redirect != null ? [path_rule.value.url_redirect] : []
            content {
              host_redirect          = url_redirect.value.host_redirect
              https_redirect         = url_redirect.value.https_redirect
              path_redirect          = url_redirect.value.path_redirect
              prefix_redirect        = url_redirect.value.prefix_redirect
              redirect_response_code = url_redirect.value.redirect_response_code
              strip_query            = url_redirect.value.strip_query
            }
          }

          dynamic "route_action" {
            for_each = path_rule.value.route_action != null ? [path_rule.value.route_action] : []
            content {
              dynamic "weighted_backend_services" {
                for_each = route_action.value.weighted_backend_services
                content {
                  backend_service = weighted_backend_services.value.backend_service
                  weight          = weighted_backend_services.value.weight

                  dynamic "header_action" {
                    for_each = weighted_backend_services.value.header_action != null ? [weighted_backend_services.value.header_action] : []
                    content {
                      dynamic "request_headers_to_add" {
                        for_each = header_action.value.request_headers_to_add
                        content {
                          header_name  = request_headers_to_add.value.header_name
                          header_value = request_headers_to_add.value.header_value
                          replace      = request_headers_to_add.value.replace
                        }
                      }
                      request_headers_to_remove = length(header_action.value.request_headers_to_remove) > 0 ? header_action.value.request_headers_to_remove : null
                      dynamic "response_headers_to_add" {
                        for_each = header_action.value.response_headers_to_add
                        content {
                          header_name  = response_headers_to_add.value.header_name
                          header_value = response_headers_to_add.value.header_value
                          replace      = response_headers_to_add.value.replace
                        }
                      }
                      response_headers_to_remove = length(header_action.value.response_headers_to_remove) > 0 ? header_action.value.response_headers_to_remove : null
                    }
                  }
                }
              }

              dynamic "url_rewrite" {
                for_each = route_action.value.url_rewrite != null ? [route_action.value.url_rewrite] : []
                content {
                  host_rewrite        = url_rewrite.value.host_rewrite
                  path_prefix_rewrite = url_rewrite.value.path_prefix_rewrite
                }
              }

              dynamic "timeout" {
                for_each = route_action.value.timeout != null ? [route_action.value.timeout] : []
                content {
                  seconds = coalesce(timeout.value.seconds, 0)
                  nanos   = coalesce(timeout.value.nanos, 0) != 0 ? timeout.value.nanos : null
                }
              }

              dynamic "retry_policy" {
                for_each = route_action.value.retry_policy != null ? [route_action.value.retry_policy] : []
                content {
                  num_retries      = retry_policy.value.num_retries != 0 ? retry_policy.value.num_retries : null
                  retry_conditions = length(retry_policy.value.retry_conditions) > 0 ? retry_policy.value.retry_conditions : null

                  dynamic "per_try_timeout" {
                    for_each = retry_policy.value.per_try_timeout != null ? [retry_policy.value.per_try_timeout] : []
                    content {
                      seconds = coalesce(per_try_timeout.value.seconds, 0)
                      nanos   = coalesce(per_try_timeout.value.nanos, 0) != 0 ? per_try_timeout.value.nanos : null
                    }
                  }
                }
              }

              dynamic "request_mirror_policy" {
                for_each = route_action.value.request_mirror_policy != null ? [route_action.value.request_mirror_policy] : []
                content {
                  backend_service = request_mirror_policy.value.backend_service
                }
              }

              dynamic "cors_policy" {
                for_each = route_action.value.cors_policy != null ? [route_action.value.cors_policy] : []
                content {
                  allow_credentials    = cors_policy.value.allow_credentials
                  allow_headers        = length(cors_policy.value.allow_headers) > 0 ? cors_policy.value.allow_headers : null
                  allow_methods        = length(cors_policy.value.allow_methods) > 0 ? cors_policy.value.allow_methods : null
                  allow_origin_regexes = length(cors_policy.value.allow_origin_regexes) > 0 ? cors_policy.value.allow_origin_regexes : null
                  allow_origins        = length(cors_policy.value.allow_origins) > 0 ? cors_policy.value.allow_origins : null
                  disabled             = cors_policy.value.disabled
                  expose_headers       = length(cors_policy.value.expose_headers) > 0 ? cors_policy.value.expose_headers : null
                  max_age              = cors_policy.value.max_age != 0 ? cors_policy.value.max_age : null
                }
              }

              dynamic "fault_injection_policy" {
                for_each = route_action.value.fault_injection_policy != null ? [route_action.value.fault_injection_policy] : []
                content {
                  dynamic "abort" {
                    for_each = fault_injection_policy.value.abort != null ? [fault_injection_policy.value.abort] : []
                    content {
                      http_status = abort.value.http_status != 0 ? abort.value.http_status : null
                      percentage  = abort.value.percentage
                    }
                  }
                  dynamic "delay" {
                    for_each = fault_injection_policy.value.delay != null ? [fault_injection_policy.value.delay] : []
                    content {
                      percentage = delay.value.percentage
                      dynamic "fixed_delay" {
                        for_each = delay.value.fixed_delay != null ? [delay.value.fixed_delay] : []
                        content {
                          seconds = coalesce(fixed_delay.value.seconds, 0)
                          nanos   = coalesce(fixed_delay.value.nanos, 0) != 0 ? fixed_delay.value.nanos : null
                        }
                      }
                    }
                  }
                }
              }

              dynamic "max_stream_duration" {
                for_each = route_action.value.max_stream_duration != null ? [route_action.value.max_stream_duration] : []
                content {
                  seconds = coalesce(max_stream_duration.value.seconds, 0)
                  nanos   = coalesce(max_stream_duration.value.nanos, 0) != 0 ? max_stream_duration.value.nanos : null
                }
              }

              dynamic "cache_policy" {
                for_each = route_action.value.cache_policy != null ? [route_action.value.cache_policy] : []
                content {
                  cache_mode                        = cache_policy.value.cache_mode != "" ? cache_policy.value.cache_mode : null
                  cache_bypass_request_header_names = length(cache_policy.value.cache_bypass_request_header_names) > 0 ? cache_policy.value.cache_bypass_request_header_names : null
                  negative_caching                  = cache_policy.value.negative_caching
                  request_coalescing                = cache_policy.value.request_coalescing

                  dynamic "cache_key_policy" {
                    for_each = cache_policy.value.cache_key_policy != null ? [cache_policy.value.cache_key_policy] : []
                    content {
                      excluded_query_parameters = length(cache_key_policy.value.excluded_query_parameters) > 0 ? cache_key_policy.value.excluded_query_parameters : null
                      include_host              = cache_key_policy.value.include_host
                      include_protocol          = cache_key_policy.value.include_protocol
                      include_query_string      = cache_key_policy.value.include_query_string
                      included_cookie_names     = length(cache_key_policy.value.included_cookie_names) > 0 ? cache_key_policy.value.included_cookie_names : null
                      included_header_names     = length(cache_key_policy.value.included_header_names) > 0 ? cache_key_policy.value.included_header_names : null
                      included_query_parameters = length(cache_key_policy.value.included_query_parameters) > 0 ? cache_key_policy.value.included_query_parameters : null
                    }
                  }

                  dynamic "client_ttl" {
                    for_each = cache_policy.value.client_ttl != null ? [cache_policy.value.client_ttl] : []
                    content {
                      seconds = coalesce(client_ttl.value.seconds, 0)
                      nanos   = coalesce(client_ttl.value.nanos, 0) != 0 ? client_ttl.value.nanos : null
                    }
                  }

                  dynamic "default_ttl" {
                    for_each = cache_policy.value.default_ttl != null ? [cache_policy.value.default_ttl] : []
                    content {
                      seconds = coalesce(default_ttl.value.seconds, 0)
                      nanos   = coalesce(default_ttl.value.nanos, 0) != 0 ? default_ttl.value.nanos : null
                    }
                  }

                  dynamic "max_ttl" {
                    for_each = cache_policy.value.max_ttl != null ? [cache_policy.value.max_ttl] : []
                    content {
                      seconds = coalesce(max_ttl.value.seconds, 0)
                      nanos   = coalesce(max_ttl.value.nanos, 0) != 0 ? max_ttl.value.nanos : null
                    }
                  }

                  dynamic "serve_while_stale" {
                    for_each = cache_policy.value.serve_while_stale != null ? [cache_policy.value.serve_while_stale] : []
                    content {
                      seconds = coalesce(serve_while_stale.value.seconds, 0)
                      nanos   = coalesce(serve_while_stale.value.nanos, 0) != 0 ? serve_while_stale.value.nanos : null
                    }
                  }

                  dynamic "negative_caching_policy" {
                    for_each = cache_policy.value.negative_caching_policy
                    content {
                      code = negative_caching_policy.value.code != 0 ? negative_caching_policy.value.code : null
                      dynamic "ttl" {
                        for_each = negative_caching_policy.value.ttl != null ? [negative_caching_policy.value.ttl] : []
                        content {
                          seconds = coalesce(ttl.value.seconds, 0)
                          nanos   = coalesce(ttl.value.nanos, 0) != 0 ? ttl.value.nanos : null
                        }
                      }
                    }
                  }
                }
              }
            }
          }

          dynamic "custom_error_response_policy" {
            for_each = path_rule.value.custom_error_response_policy != null ? [path_rule.value.custom_error_response_policy] : []
            content {
              error_service = custom_error_response_policy.value.error_service

              dynamic "error_response_rule" {
                for_each = custom_error_response_policy.value.error_response_rules
                content {
                  match_response_codes   = error_response_rule.value.match_response_codes
                  override_response_code = error_response_rule.value.override_response_code
                  path                   = error_response_rule.value.path
                }
              }
            }
          }
        }
      }

      dynamic "route_rules" {
        for_each = path_matcher.value.route_rules
        content {
          priority = route_rules.value.priority
          service  = route_rules.value.service

          dynamic "url_redirect" {
            for_each = route_rules.value.url_redirect != null ? [route_rules.value.url_redirect] : []
            content {
              host_redirect          = url_redirect.value.host_redirect
              https_redirect         = url_redirect.value.https_redirect
              path_redirect          = url_redirect.value.path_redirect
              prefix_redirect        = url_redirect.value.prefix_redirect
              redirect_response_code = url_redirect.value.redirect_response_code
              strip_query            = url_redirect.value.strip_query
            }
          }

          dynamic "route_action" {
            for_each = route_rules.value.route_action != null ? [route_rules.value.route_action] : []
            content {
              dynamic "weighted_backend_services" {
                for_each = route_action.value.weighted_backend_services
                content {
                  backend_service = weighted_backend_services.value.backend_service
                  weight          = weighted_backend_services.value.weight

                  dynamic "header_action" {
                    for_each = weighted_backend_services.value.header_action != null ? [weighted_backend_services.value.header_action] : []
                    content {
                      dynamic "request_headers_to_add" {
                        for_each = header_action.value.request_headers_to_add
                        content {
                          header_name  = request_headers_to_add.value.header_name
                          header_value = request_headers_to_add.value.header_value
                          replace      = request_headers_to_add.value.replace
                        }
                      }
                      request_headers_to_remove = length(header_action.value.request_headers_to_remove) > 0 ? header_action.value.request_headers_to_remove : null
                      dynamic "response_headers_to_add" {
                        for_each = header_action.value.response_headers_to_add
                        content {
                          header_name  = response_headers_to_add.value.header_name
                          header_value = response_headers_to_add.value.header_value
                          replace      = response_headers_to_add.value.replace
                        }
                      }
                      response_headers_to_remove = length(header_action.value.response_headers_to_remove) > 0 ? header_action.value.response_headers_to_remove : null
                    }
                  }
                }
              }

              dynamic "url_rewrite" {
                for_each = route_action.value.url_rewrite != null ? [route_action.value.url_rewrite] : []
                content {
                  host_rewrite          = url_rewrite.value.host_rewrite
                  path_prefix_rewrite   = url_rewrite.value.path_prefix_rewrite
                  path_template_rewrite = url_rewrite.value.path_template_rewrite
                }
              }

              dynamic "timeout" {
                for_each = route_action.value.timeout != null ? [route_action.value.timeout] : []
                content {
                  seconds = coalesce(timeout.value.seconds, 0)
                  nanos   = coalesce(timeout.value.nanos, 0) != 0 ? timeout.value.nanos : null
                }
              }

              dynamic "retry_policy" {
                for_each = route_action.value.retry_policy != null ? [route_action.value.retry_policy] : []
                content {
                  num_retries      = retry_policy.value.num_retries != 0 ? retry_policy.value.num_retries : null
                  retry_conditions = length(retry_policy.value.retry_conditions) > 0 ? retry_policy.value.retry_conditions : null

                  dynamic "per_try_timeout" {
                    for_each = retry_policy.value.per_try_timeout != null ? [retry_policy.value.per_try_timeout] : []
                    content {
                      seconds = coalesce(per_try_timeout.value.seconds, 0)
                      nanos   = coalesce(per_try_timeout.value.nanos, 0) != 0 ? per_try_timeout.value.nanos : null
                    }
                  }
                }
              }

              dynamic "request_mirror_policy" {
                for_each = route_action.value.request_mirror_policy != null ? [route_action.value.request_mirror_policy] : []
                content {
                  backend_service = request_mirror_policy.value.backend_service
                }
              }

              dynamic "cors_policy" {
                for_each = route_action.value.cors_policy != null ? [route_action.value.cors_policy] : []
                content {
                  allow_credentials    = cors_policy.value.allow_credentials
                  allow_headers        = length(cors_policy.value.allow_headers) > 0 ? cors_policy.value.allow_headers : null
                  allow_methods        = length(cors_policy.value.allow_methods) > 0 ? cors_policy.value.allow_methods : null
                  allow_origin_regexes = length(cors_policy.value.allow_origin_regexes) > 0 ? cors_policy.value.allow_origin_regexes : null
                  allow_origins        = length(cors_policy.value.allow_origins) > 0 ? cors_policy.value.allow_origins : null
                  disabled             = cors_policy.value.disabled
                  expose_headers       = length(cors_policy.value.expose_headers) > 0 ? cors_policy.value.expose_headers : null
                  max_age              = cors_policy.value.max_age != 0 ? cors_policy.value.max_age : null
                }
              }

              dynamic "fault_injection_policy" {
                for_each = route_action.value.fault_injection_policy != null ? [route_action.value.fault_injection_policy] : []
                content {
                  dynamic "abort" {
                    for_each = fault_injection_policy.value.abort != null ? [fault_injection_policy.value.abort] : []
                    content {
                      http_status = abort.value.http_status != 0 ? abort.value.http_status : null
                      percentage  = abort.value.percentage
                    }
                  }
                  dynamic "delay" {
                    for_each = fault_injection_policy.value.delay != null ? [fault_injection_policy.value.delay] : []
                    content {
                      percentage = delay.value.percentage
                      dynamic "fixed_delay" {
                        for_each = delay.value.fixed_delay != null ? [delay.value.fixed_delay] : []
                        content {
                          seconds = coalesce(fixed_delay.value.seconds, 0)
                          nanos   = coalesce(fixed_delay.value.nanos, 0) != 0 ? fixed_delay.value.nanos : null
                        }
                      }
                    }
                  }
                }
              }

              dynamic "max_stream_duration" {
                for_each = route_action.value.max_stream_duration != null ? [route_action.value.max_stream_duration] : []
                content {
                  seconds = coalesce(max_stream_duration.value.seconds, 0)
                  nanos   = coalesce(max_stream_duration.value.nanos, 0) != 0 ? max_stream_duration.value.nanos : null
                }
              }

              dynamic "cache_policy" {
                for_each = route_action.value.cache_policy != null ? [route_action.value.cache_policy] : []
                content {
                  cache_mode                        = cache_policy.value.cache_mode != "" ? cache_policy.value.cache_mode : null
                  cache_bypass_request_header_names = length(cache_policy.value.cache_bypass_request_header_names) > 0 ? cache_policy.value.cache_bypass_request_header_names : null
                  negative_caching                  = cache_policy.value.negative_caching
                  request_coalescing                = cache_policy.value.request_coalescing

                  dynamic "cache_key_policy" {
                    for_each = cache_policy.value.cache_key_policy != null ? [cache_policy.value.cache_key_policy] : []
                    content {
                      excluded_query_parameters = length(cache_key_policy.value.excluded_query_parameters) > 0 ? cache_key_policy.value.excluded_query_parameters : null
                      include_host              = cache_key_policy.value.include_host
                      include_protocol          = cache_key_policy.value.include_protocol
                      include_query_string      = cache_key_policy.value.include_query_string
                      included_cookie_names     = length(cache_key_policy.value.included_cookie_names) > 0 ? cache_key_policy.value.included_cookie_names : null
                      included_header_names     = length(cache_key_policy.value.included_header_names) > 0 ? cache_key_policy.value.included_header_names : null
                      included_query_parameters = length(cache_key_policy.value.included_query_parameters) > 0 ? cache_key_policy.value.included_query_parameters : null
                    }
                  }

                  dynamic "client_ttl" {
                    for_each = cache_policy.value.client_ttl != null ? [cache_policy.value.client_ttl] : []
                    content {
                      seconds = coalesce(client_ttl.value.seconds, 0)
                      nanos   = coalesce(client_ttl.value.nanos, 0) != 0 ? client_ttl.value.nanos : null
                    }
                  }

                  dynamic "default_ttl" {
                    for_each = cache_policy.value.default_ttl != null ? [cache_policy.value.default_ttl] : []
                    content {
                      seconds = coalesce(default_ttl.value.seconds, 0)
                      nanos   = coalesce(default_ttl.value.nanos, 0) != 0 ? default_ttl.value.nanos : null
                    }
                  }

                  dynamic "max_ttl" {
                    for_each = cache_policy.value.max_ttl != null ? [cache_policy.value.max_ttl] : []
                    content {
                      seconds = coalesce(max_ttl.value.seconds, 0)
                      nanos   = coalesce(max_ttl.value.nanos, 0) != 0 ? max_ttl.value.nanos : null
                    }
                  }

                  dynamic "serve_while_stale" {
                    for_each = cache_policy.value.serve_while_stale != null ? [cache_policy.value.serve_while_stale] : []
                    content {
                      seconds = coalesce(serve_while_stale.value.seconds, 0)
                      nanos   = coalesce(serve_while_stale.value.nanos, 0) != 0 ? serve_while_stale.value.nanos : null
                    }
                  }

                  dynamic "negative_caching_policy" {
                    for_each = cache_policy.value.negative_caching_policy
                    content {
                      code = negative_caching_policy.value.code != 0 ? negative_caching_policy.value.code : null
                      dynamic "ttl" {
                        for_each = negative_caching_policy.value.ttl != null ? [negative_caching_policy.value.ttl] : []
                        content {
                          seconds = coalesce(ttl.value.seconds, 0)
                          nanos   = coalesce(ttl.value.nanos, 0) != 0 ? ttl.value.nanos : null
                        }
                      }
                    }
                  }
                }
              }
            }
          }

          dynamic "header_action" {
            for_each = route_rules.value.header_action != null ? [route_rules.value.header_action] : []
            content {
              dynamic "request_headers_to_add" {
                for_each = header_action.value.request_headers_to_add
                content {
                  header_name  = request_headers_to_add.value.header_name
                  header_value = request_headers_to_add.value.header_value
                  replace      = request_headers_to_add.value.replace
                }
              }
              request_headers_to_remove = header_action.value.request_headers_to_remove

              dynamic "response_headers_to_add" {
                for_each = header_action.value.response_headers_to_add
                content {
                  header_name  = response_headers_to_add.value.header_name
                  header_value = response_headers_to_add.value.header_value
                  replace      = response_headers_to_add.value.replace
                }
              }
              response_headers_to_remove = header_action.value.response_headers_to_remove
            }
          }

          dynamic "custom_error_response_policy" {
            for_each = route_rules.value.custom_error_response_policy != null ? [route_rules.value.custom_error_response_policy] : []
            content {
              error_service = custom_error_response_policy.value.error_service

              dynamic "error_response_rule" {
                for_each = custom_error_response_policy.value.error_response_rules
                content {
                  match_response_codes   = error_response_rule.value.match_response_codes
                  override_response_code = error_response_rule.value.override_response_code
                  path                   = error_response_rule.value.path
                }
              }
            }
          }

          dynamic "match_rules" {
            for_each = route_rules.value.match_rules
            content {
              prefix_match        = match_rules.value.prefix_match
              full_path_match     = match_rules.value.full_path_match
              regex_match         = match_rules.value.regex_match
              path_template_match = match_rules.value.path_template_match
              ignore_case         = match_rules.value.ignore_case

              dynamic "header_matches" {
                for_each = match_rules.value.header_matches
                content {
                  header_name   = header_matches.value.header_name
                  exact_match   = header_matches.value.exact_match
                  prefix_match  = header_matches.value.prefix_match
                  suffix_match  = header_matches.value.suffix_match
                  regex_match   = header_matches.value.regex_match
                  present_match = header_matches.value.present_match
                  invert_match  = header_matches.value.invert_match

                  dynamic "range_match" {
                    for_each = header_matches.value.range_match != null ? [header_matches.value.range_match] : []
                    content {
                      range_start = range_match.value.range_start
                      range_end   = range_match.value.range_end
                    }
                  }
                }
              }

              dynamic "query_parameter_matches" {
                for_each = match_rules.value.query_parameter_matches
                content {
                  name          = query_parameter_matches.value.name
                  exact_match   = query_parameter_matches.value.exact_match
                  present_match = query_parameter_matches.value.present_match
                  regex_match   = query_parameter_matches.value.regex_match
                }
              }

              dynamic "metadata_filters" {
                for_each = match_rules.value.metadata_filters
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
            }
          }
        }
      }
    }
  }

  dynamic "test" {
    for_each = local.tests
    content {
      host                            = test.value.host
      path                            = test.value.path
      service                         = test.value.service
      description                     = test.value.description
      expected_output_url             = test.value.expected_output_url
      expected_redirect_response_code = test.value.expected_redirect_response_code

      dynamic "headers" {
        for_each = test.value.headers
        content {
          name  = headers.value.name
          value = headers.value.value
        }
      }
    }
  }

  depends_on = [google_project_service.compute_api]
}
