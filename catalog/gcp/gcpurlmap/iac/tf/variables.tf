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
  description = "Specification for the GCP Compute Engine global URL map"
  type = object({
    # The GCP project that owns the URL map. The CLI's tfvars converter
    # resolves StringValueOrRef fields to their literal string before the
    # module runs, so this arrives as a plain string.
    # If empty, the provider's default project is used (see locals.tf).
    project_id = optional(string, "")

    # Name of the URL map in GCP (RFC1035). Empty defaults to metadata.name
    # (see locals.tf). Immutable (ForceNew).
    url_map_name = optional(string, "")

    description = optional(string, "")

    # Exactly one default target is required (enforced by the spec's CEL
    # upstream). StringValueOrRef fields arrive as plain strings.
    default_service = optional(string, "")

    default_url_redirect = optional(object({
      host_redirect          = optional(string, "")
      https_redirect         = optional(bool, false)
      path_redirect          = optional(string, "")
      prefix_redirect        = optional(string, "")
      redirect_response_code = optional(string, "")
      strip_query            = optional(bool, false)
    }))

    default_route_action = optional(object({
      weighted_backend_services = optional(list(object({
        backend_service = string
        weight          = optional(number, 0)
        header_action = optional(object({
          request_headers_to_add = optional(list(object({
            header_name  = string
            header_value = string
            replace      = optional(bool, false)
          })), [])
          request_headers_to_remove = optional(list(string), [])
          response_headers_to_add = optional(list(object({
            header_name  = string
            header_value = string
            replace      = optional(bool, false)
          })), [])
          response_headers_to_remove = optional(list(string), [])
        }))
      })), [])
      url_rewrite = optional(object({
        host_rewrite        = optional(string, "")
        path_prefix_rewrite = optional(string, "")
      }))
      # Durations arrive as {seconds, nanos} pairs mirroring GCP's Duration
      # shape; zero-valued pairs are sent as-is (a present block is intent).
      timeout = optional(object({
        seconds = optional(number)
        nanos   = optional(number)
      }))
      retry_policy = optional(object({
        num_retries      = optional(number, 0)
        retry_conditions = optional(list(string), [])
        per_try_timeout = optional(object({
          seconds = optional(number)
          nanos   = optional(number)
        }))
      }))
      request_mirror_policy = optional(object({
        backend_service = string
      }))
      cors_policy = optional(object({
        allow_credentials    = optional(bool, false)
        allow_headers        = optional(list(string), [])
        allow_methods        = optional(list(string), [])
        allow_origin_regexes = optional(list(string), [])
        allow_origins        = optional(list(string), [])
        disabled             = optional(bool, false)
        expose_headers       = optional(list(string), [])
        max_age              = optional(number, 0)
      }))
      fault_injection_policy = optional(object({
        abort = optional(object({
          http_status = optional(number, 0)
          percentage  = optional(number, 0)
        }))
        delay = optional(object({
          fixed_delay = optional(object({
            seconds = optional(number)
            nanos   = optional(number)
          }))
          percentage = optional(number, 0)
        }))
      }))
      max_stream_duration = optional(object({
        seconds = optional(number)
        nanos   = optional(number)
      }))
      cache_policy = optional(object({
        cache_mode                        = optional(string, "")
        cache_bypass_request_header_names = optional(list(string), [])
        negative_caching                  = optional(bool, false)
        request_coalescing                = optional(bool, false)
        cache_key_policy = optional(object({
          excluded_query_parameters = optional(list(string), [])
          include_host              = optional(bool, false)
          include_protocol          = optional(bool, false)
          include_query_string      = optional(bool, false)
          included_cookie_names     = optional(list(string), [])
          included_header_names     = optional(list(string), [])
          included_query_parameters = optional(list(string), [])
        }))
        client_ttl        = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
        default_ttl       = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
        max_ttl           = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
        serve_while_stale = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
        negative_caching_policy = optional(list(object({
          code = optional(number, 0)
          ttl  = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
        })), [])
      }))
    }))

    default_custom_error_response_policy = optional(object({
      error_service = optional(string, "")
      error_response_rules = optional(list(object({
        match_response_codes   = list(string)
        override_response_code = optional(number)
        path                   = string
      })), [])
    }))

    header_action = optional(object({
      request_headers_to_add = optional(list(object({
        header_name  = string
        header_value = string
        replace      = optional(bool, false)
      })), [])
      request_headers_to_remove = optional(list(string), [])
      response_headers_to_add = optional(list(object({
        header_name  = string
        header_value = string
        replace      = optional(bool, false)
      })), [])
      response_headers_to_remove = optional(list(string), [])
    }))

    host_rules = optional(list(object({
      hosts        = list(string)
      path_matcher = string
      description  = optional(string, "")
    })), [])

    path_matchers = optional(list(object({
      name            = string
      description     = optional(string, "")
      default_service = optional(string, "")
      default_url_redirect = optional(object({
        host_redirect          = optional(string, "")
        https_redirect         = optional(bool, false)
        path_redirect          = optional(string, "")
        prefix_redirect        = optional(string, "")
        redirect_response_code = optional(string, "")
        strip_query            = optional(bool, false)
      }))
      default_route_action = optional(object({
        weighted_backend_services = optional(list(object({
          backend_service = string
          weight          = optional(number, 0)
          header_action = optional(object({
            request_headers_to_add = optional(list(object({
              header_name  = string
              header_value = string
              replace      = optional(bool, false)
            })), [])
            request_headers_to_remove = optional(list(string), [])
            response_headers_to_add = optional(list(object({
              header_name  = string
              header_value = string
              replace      = optional(bool, false)
            })), [])
            response_headers_to_remove = optional(list(string), [])
          }))
        })), [])
        url_rewrite = optional(object({
          host_rewrite        = optional(string, "")
          path_prefix_rewrite = optional(string, "")
        }))
        timeout = optional(object({
          seconds = optional(number)
          nanos   = optional(number)
        }))
        retry_policy = optional(object({
          num_retries      = optional(number, 0)
          retry_conditions = optional(list(string), [])
          per_try_timeout = optional(object({
            seconds = optional(number)
            nanos   = optional(number)
          }))
        }))
        request_mirror_policy = optional(object({
          backend_service = string
        }))
        cors_policy = optional(object({
          allow_credentials    = optional(bool, false)
          allow_headers        = optional(list(string), [])
          allow_methods        = optional(list(string), [])
          allow_origin_regexes = optional(list(string), [])
          allow_origins        = optional(list(string), [])
          disabled             = optional(bool, false)
          expose_headers       = optional(list(string), [])
          max_age              = optional(number, 0)
        }))
        fault_injection_policy = optional(object({
          abort = optional(object({
            http_status = optional(number, 0)
            percentage  = optional(number, 0)
          }))
          delay = optional(object({
            fixed_delay = optional(object({
              seconds = optional(number)
              nanos   = optional(number)
            }))
            percentage = optional(number, 0)
          }))
        }))
        max_stream_duration = optional(object({
          seconds = optional(number)
          nanos   = optional(number)
        }))
        cache_policy = optional(object({
          cache_mode                        = optional(string, "")
          cache_bypass_request_header_names = optional(list(string), [])
          negative_caching                  = optional(bool, false)
          request_coalescing                = optional(bool, false)
          cache_key_policy = optional(object({
            excluded_query_parameters = optional(list(string), [])
            include_host              = optional(bool, false)
            include_protocol          = optional(bool, false)
            include_query_string      = optional(bool, false)
            included_cookie_names     = optional(list(string), [])
            included_header_names     = optional(list(string), [])
            included_query_parameters = optional(list(string), [])
          }))
          client_ttl        = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
          default_ttl       = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
          max_ttl           = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
          serve_while_stale = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
          negative_caching_policy = optional(list(object({
            code = optional(number, 0)
            ttl  = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
          })), [])
        }))
      }))
      default_custom_error_response_policy = optional(object({
        error_service = optional(string, "")
        error_response_rules = optional(list(object({
          match_response_codes   = list(string)
          override_response_code = optional(number)
          path                   = string
        })), [])
      }))
      header_action = optional(object({
        request_headers_to_add = optional(list(object({
          header_name  = string
          header_value = string
          replace      = optional(bool, false)
        })), [])
        request_headers_to_remove = optional(list(string), [])
        response_headers_to_add = optional(list(object({
          header_name  = string
          header_value = string
          replace      = optional(bool, false)
        })), [])
        response_headers_to_remove = optional(list(string), [])
      }))
      path_rules = optional(list(object({
        paths   = list(string)
        service = optional(string, "")
        url_redirect = optional(object({
          host_redirect          = optional(string, "")
          https_redirect         = optional(bool, false)
          path_redirect          = optional(string, "")
          prefix_redirect        = optional(string, "")
          redirect_response_code = optional(string, "")
          strip_query            = optional(bool, false)
        }))
        route_action = optional(object({
          weighted_backend_services = optional(list(object({
            backend_service = string
            weight          = optional(number, 0)
            header_action = optional(object({
              request_headers_to_add = optional(list(object({
                header_name  = string
                header_value = string
                replace      = optional(bool, false)
              })), [])
              request_headers_to_remove = optional(list(string), [])
              response_headers_to_add = optional(list(object({
                header_name  = string
                header_value = string
                replace      = optional(bool, false)
              })), [])
              response_headers_to_remove = optional(list(string), [])
            }))
          })), [])
          url_rewrite = optional(object({
            host_rewrite        = optional(string, "")
            path_prefix_rewrite = optional(string, "")
          }))
          timeout = optional(object({
            seconds = optional(number)
            nanos   = optional(number)
          }))
          retry_policy = optional(object({
            num_retries      = optional(number, 0)
            retry_conditions = optional(list(string), [])
            per_try_timeout = optional(object({
              seconds = optional(number)
              nanos   = optional(number)
            }))
          }))
          request_mirror_policy = optional(object({
            backend_service = string
          }))
          cors_policy = optional(object({
            allow_credentials    = optional(bool, false)
            allow_headers        = optional(list(string), [])
            allow_methods        = optional(list(string), [])
            allow_origin_regexes = optional(list(string), [])
            allow_origins        = optional(list(string), [])
            disabled             = optional(bool, false)
            expose_headers       = optional(list(string), [])
            max_age              = optional(number, 0)
          }))
          fault_injection_policy = optional(object({
            abort = optional(object({
              http_status = optional(number, 0)
              percentage  = optional(number, 0)
            }))
            delay = optional(object({
              fixed_delay = optional(object({
                seconds = optional(number)
                nanos   = optional(number)
              }))
              percentage = optional(number, 0)
            }))
          }))
          max_stream_duration = optional(object({
            seconds = optional(number)
            nanos   = optional(number)
          }))
          cache_policy = optional(object({
            cache_mode                        = optional(string, "")
            cache_bypass_request_header_names = optional(list(string), [])
            negative_caching                  = optional(bool, false)
            request_coalescing                = optional(bool, false)
            cache_key_policy = optional(object({
              excluded_query_parameters = optional(list(string), [])
              include_host              = optional(bool, false)
              include_protocol          = optional(bool, false)
              include_query_string      = optional(bool, false)
              included_cookie_names     = optional(list(string), [])
              included_header_names     = optional(list(string), [])
              included_query_parameters = optional(list(string), [])
            }))
            client_ttl        = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
            default_ttl       = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
            max_ttl           = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
            serve_while_stale = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
            negative_caching_policy = optional(list(object({
              code = optional(number, 0)
              ttl  = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
            })), [])
          }))
        }))
        custom_error_response_policy = optional(object({
          error_service = optional(string, "")
          error_response_rules = optional(list(object({
            match_response_codes   = list(string)
            override_response_code = optional(number)
            path                   = string
          })), [])
        }))
      })), [])
      route_rules = optional(list(object({
        priority = number
        service  = optional(string, "")
        url_redirect = optional(object({
          host_redirect          = optional(string, "")
          https_redirect         = optional(bool, false)
          path_redirect          = optional(string, "")
          prefix_redirect        = optional(string, "")
          redirect_response_code = optional(string, "")
          strip_query            = optional(bool, false)
        }))
        route_action = optional(object({
          weighted_backend_services = optional(list(object({
            backend_service = string
            weight          = optional(number, 0)
            header_action = optional(object({
              request_headers_to_add = optional(list(object({
                header_name  = string
                header_value = string
                replace      = optional(bool, false)
              })), [])
              request_headers_to_remove = optional(list(string), [])
              response_headers_to_add = optional(list(object({
                header_name  = string
                header_value = string
                replace      = optional(bool, false)
              })), [])
              response_headers_to_remove = optional(list(string), [])
            }))
          })), [])
          url_rewrite = optional(object({
            host_rewrite          = optional(string, "")
            path_prefix_rewrite   = optional(string, "")
            path_template_rewrite = optional(string, "")
          }))
          timeout = optional(object({
            seconds = optional(number)
            nanos   = optional(number)
          }))
          retry_policy = optional(object({
            num_retries      = optional(number, 0)
            retry_conditions = optional(list(string), [])
            per_try_timeout = optional(object({
              seconds = optional(number)
              nanos   = optional(number)
            }))
          }))
          request_mirror_policy = optional(object({
            backend_service = string
          }))
          cors_policy = optional(object({
            allow_credentials    = optional(bool, false)
            allow_headers        = optional(list(string), [])
            allow_methods        = optional(list(string), [])
            allow_origin_regexes = optional(list(string), [])
            allow_origins        = optional(list(string), [])
            disabled             = optional(bool, false)
            expose_headers       = optional(list(string), [])
            max_age              = optional(number, 0)
          }))
          fault_injection_policy = optional(object({
            abort = optional(object({
              http_status = optional(number, 0)
              percentage  = optional(number, 0)
            }))
            delay = optional(object({
              fixed_delay = optional(object({
                seconds = optional(number)
                nanos   = optional(number)
              }))
              percentage = optional(number, 0)
            }))
          }))
          max_stream_duration = optional(object({
            seconds = optional(number)
            nanos   = optional(number)
          }))
          cache_policy = optional(object({
            cache_mode                        = optional(string, "")
            cache_bypass_request_header_names = optional(list(string), [])
            negative_caching                  = optional(bool, false)
            request_coalescing                = optional(bool, false)
            cache_key_policy = optional(object({
              excluded_query_parameters = optional(list(string), [])
              include_host              = optional(bool, false)
              include_protocol          = optional(bool, false)
              include_query_string      = optional(bool, false)
              included_cookie_names     = optional(list(string), [])
              included_header_names     = optional(list(string), [])
              included_query_parameters = optional(list(string), [])
            }))
            client_ttl        = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
            default_ttl       = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
            max_ttl           = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
            serve_while_stale = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
            negative_caching_policy = optional(list(object({
              code = optional(number, 0)
              ttl  = optional(object({ seconds = optional(number, 0), nanos = optional(number, 0) }))
            })), [])
          }))
        }))
        header_action = optional(object({
          request_headers_to_add = optional(list(object({
            header_name  = string
            header_value = string
            replace      = optional(bool, false)
          })), [])
          request_headers_to_remove = optional(list(string), [])
          response_headers_to_add = optional(list(object({
            header_name  = string
            header_value = string
            replace      = optional(bool, false)
          })), [])
          response_headers_to_remove = optional(list(string), [])
        }))
        custom_error_response_policy = optional(object({
          error_service = optional(string, "")
          error_response_rules = optional(list(object({
            match_response_codes   = list(string)
            override_response_code = optional(number)
            path                   = string
          })), [])
        }))
        match_rules = list(object({
          prefix_match        = optional(string, "")
          full_path_match     = optional(string, "")
          regex_match         = optional(string, "")
          path_template_match = optional(string, "")
          ignore_case         = optional(bool, false)
          header_matches = optional(list(object({
            header_name   = string
            exact_match   = optional(string, "")
            prefix_match  = optional(string, "")
            suffix_match  = optional(string, "")
            regex_match   = optional(string, "")
            present_match = optional(bool, false)
            invert_match  = optional(bool, false)
            range_match = optional(object({
              range_start = optional(number, 0)
              range_end   = optional(number, 0)
            }))
          })), [])
          query_parameter_matches = optional(list(object({
            name          = string
            exact_match   = optional(string, "")
            present_match = optional(bool, false)
            regex_match   = optional(string, "")
          })), [])
          metadata_filters = optional(list(object({
            filter_match_criteria = string
            filter_labels = list(object({
              name  = string
              value = string
            }))
          })), [])
        }))
      })), [])
    })), [])

    tests = optional(list(object({
      host                            = string
      path                            = string
      service                         = optional(string, "")
      description                     = optional(string, "")
      expected_output_url             = optional(string, "")
      expected_redirect_response_code = optional(number)
      headers = optional(list(object({
        name  = string
        value = string
      })), [])
    })), [])

    # Client-side destroy stance: DELETE (default), PREVENT (fail the
    # destroy), or ABANDON (drop from state, keep serving). Never sent to
    # the GCP API.
    deletion_policy = optional(string, "")
  })
}
