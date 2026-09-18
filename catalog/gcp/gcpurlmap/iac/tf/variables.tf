variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpUrlMap specification"
  type = object({
    # The GCP project that owns the URL map.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing it destroys and recreates the URL map.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the URL map in GCP. Must be 1-63 characters: lowercase letters,
    # digits, and hyphens; must start with a letter and end with a letter or
    # digit. If not specified, defaults to metadata.name.
    # Immutable: changing it destroys and recreates the URL map, briefly
    # breaking every target proxy that references the old self_link.
    url_map_name = optional(string, "")

    # What this URL map fronts and how it routes — write it for the operator
    # reading a routing incident later. Mutable.
    description = optional(string, "")

    # The default target when no host/path rule matches — a backend service or
    # backend bucket. Reference a GcpBackendService or GcpBackendBucket, or
    # provide a self-link directly. Exactly one of default_service,
    # default_url_redirect, or default_route_action must be set. Mutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    default_service = optional(string, "")

    # Redirect unmatched requests instead of serving them (e.g. an
    # apex-to-www or http-to-https redirect as the catch-all). Mutually
    # exclusive with default_service and default_route_action.
    default_url_redirect = optional(object({
      # Replace the host in the redirect Location. Empty keeps the request host.
      host_redirect = optional(string, "")

      # Redirect to HTTPS (scheme becomes https). The standard http→https
      # upgrade. Default false.
      https_redirect = optional(bool, false)

      # Replace the entire path with this value. Mutually exclusive with
      # prefix_redirect. Empty keeps the request path.
      path_redirect = optional(string, "")

      # Replace the matched path prefix with this value, keeping the remainder.
      # Mutually exclusive with path_redirect.
      prefix_redirect = optional(string, "")

      # The HTTP redirect status code: FOUND (302), MOVED_PERMANENTLY_DEFAULT
      # (301), PERMANENT_REDIRECT (308), SEE_OTHER (303), or TEMPORARY_REDIRECT
      # (307). Empty uses the GCP default (MOVED_PERMANENTLY_DEFAULT).
      redirect_response_code = optional(string, "")

      # Drop the query string from the redirect Location. Default false (the query
      # string is preserved).
      strip_query = optional(bool, false)
    }))

    # Advanced default handling: weight traffic across several backends, rewrite
    # URLs, set timeouts/retries. Mutually exclusive with default_service and
    # default_url_redirect (its weighted_backend_services is the third arm of
    # the default-target choice).
    default_route_action = optional(object({
      # Split traffic across multiple backend services by weight — the mechanism
      # for weighted canary and blue/green rollouts. The weights are relative; a
      # backend's share is its weight over the sum of weights.
      weighted_backend_services = optional(list(object({
        # The backend service receiving this share of traffic. Reference a
        # GcpBackendService or provide a self-link directly.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        backend_service = string

        # Relative weight of this backend (0-1000). Its share is weight over the sum
        # of all weights in the split; 0 drains this backend from the split.
        weight = optional(number, 0)

        # Header mutations applied ONLY to the share of traffic sent to this
        # backend — e.g. tag canary responses with an identifying header so
        # clients and dashboards can tell which arm served them. Applied after
        # the URL-map-level and route-level header actions.
        header_action = optional(object({
          # Headers to add to the request before it reaches the backend.
          request_headers_to_add = optional(list(object({
            # The header name (e.g. "X-Client-Geo").
            header_name = string

            # The header value. May use load-balancer variables (e.g. "{client_region}").
            header_value = string

            # Replace an existing header of the same name (true) or append to it
            # (false).
            replace = optional(bool, false)
          })), [])

          # Header names to strip from the request before it reaches the backend.
          request_headers_to_remove = optional(list(string), [])

          # Headers to add to the response before it returns to the client.
          response_headers_to_add = optional(list(object({
            # The header name (e.g. "X-Client-Geo").
            header_name = string

            # The header value. May use load-balancer variables (e.g. "{client_region}").
            header_value = string

            # Replace an existing header of the same name (true) or append to it
            # (false).
            replace = optional(bool, false)
          })), [])

          # Header names to strip from the response before it returns to the client.
          response_headers_to_remove = optional(list(string), [])
        }))
      })), [])

      # Rewrite the host and/or path before forwarding to the backend.
      url_rewrite = optional(object({
        # Replace the request Host header with this value before forwarding.
        host_rewrite = optional(string, "")

        # Replace the matched path prefix with this value. Mutually exclusive with
        # path_template_rewrite.
        path_prefix_rewrite = optional(string, "")

        # Rewrite the path using a template that references named path variables
        # captured by a route rule's path_template_match (e.g. "/v2/{country}").
        # Honored only inside a route_rule's route_action — GCP rejects it in
        # default and path-rule route actions. Mutually exclusive with
        # path_prefix_rewrite.
        path_template_rewrite = optional(string, "")
      }))

      # Total time budget for the request, INCLUDING all retries (from the
      # first byte of the request to the last byte of the response). Pair with
      # retry_policy.per_try_timeout: per-try bounds one attempt, this bounds
      # the whole exchange. Unset uses the backend service's own timeout.
      # Not permitted when the route targets a redirect.
      timeout = optional(object({
        # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
        seconds = optional(number)

        # Fraction of a second at nanosecond resolution (0 to 999,999,999).
        # Durations under one second use seconds = 0 and a positive nanos.
        nanos = optional(number)
      }))

      # Retry failed requests to the backend. Which failures count is chosen
      # by retry_conditions; how long each attempt may run by per_try_timeout.
      # Retries consume the overall timeout's budget — they never extend it.
      retry_policy = optional(object({
        # Number of allowed retries (GCP defaults to 1 when unset/0). Each retry
        # still spends the route's overall timeout budget.
        num_retries = optional(number, 0)

        # Which failures trigger a retry. 5xx (any 5xx or no response at all),
        # gateway-error (502/503/504 only), connect-failure, retriable-4xx
        # (currently only 409), refused-stream, and the gRPC status conditions
        # cancelled, deadline-exceeded, resource-exhausted, unavailable.
        retry_conditions = optional(list(string), [])

        # Time budget for EACH retry attempt (the route's timeout bounds the
        # whole exchange across attempts). Unset uses the overall timeout for
        # every attempt.
        per_try_timeout = optional(object({
          # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
          seconds = optional(number)

          # Fraction of a second at nanosecond resolution (0 to 999,999,999).
          # Durations under one second use seconds = 0 and a positive nanos.
          nanos = optional(number)
        }))
      }))

      # Mirror every matched request to a second backend service, fire-and-
      # forget (responses from the mirror are discarded; the client sees only
      # the primary's response). Useful for shadow-testing a new stack with
      # production traffic — the mirror backend must be sized for the full
      # mirrored load.
      request_mirror_policy = optional(object({
        # The backend service receiving the mirrored copy of every matched
        # request. Reference a GcpBackendService or provide a self-link
        # directly. Responses from this backend are discarded.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        backend_service = string
      }))

      # Answer cross-origin (CORS) preflights and stamp CORS headers at the
      # load balancer, before requests reach the backend.
      cors_policy = optional(object({
        # Sets Access-Control-Allow-Credentials: allow requests carrying
        # credentials (cookies, authorization headers). Default false.
        allow_credentials = optional(bool, false)

        # Headers the client may send (Access-Control-Allow-Headers).
        allow_headers = optional(list(string), [])

        # Methods the client may use (Access-Control-Allow-Methods), e.g.
        # ["GET", "POST", "OPTIONS"].
        allow_methods = optional(list(string), [])

        # Regular expressions matching allowed origins; a request origin is
        # allowed when it matches any listed regex or exact allow_origins entry.
        allow_origin_regexes = optional(list(string), [])

        # Exact origins allowed, e.g. "https://app.example.com".
        allow_origins = optional(list(string), [])

        # Disable this CORS policy without deleting its configuration (true
        # turns the policy off). Default false — the policy is in effect.
        disabled = optional(bool, false)

        # Headers the browser may read from the response
        # (Access-Control-Expose-Headers).
        expose_headers = optional(list(string), [])

        # Seconds a browser may cache the preflight response
        # (Access-Control-Max-Age).
        max_age = optional(number, 0)
      }))

      # Deliberately inject failures (aborts with a chosen status, fixed
      # delays) into a percentage of matched requests — chaos/resilience
      # testing of the CLIENTS of this route. Never leave enabled on a
      # production default route: the injected failures are real to callers.
      fault_injection_policy = optional(object({
        # Abort a percentage of requests with a fixed HTTP status, before they
        # reach the backend.
        abort = optional(object({
          # HTTP status returned to aborted requests (200-599).
          http_status = optional(number, 0)

          # Percentage of requests aborted (0.0-100.0).
          percentage = optional(number, 0)
        }))

        # Delay a percentage of requests by a fixed duration before forwarding.
        delay = optional(object({
          # How long delayed requests are held before forwarding.
          fixed_delay = optional(object({
            # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
            seconds = optional(number)

            # Fraction of a second at nanosecond resolution (0 to 999,999,999).
            # Durations under one second use seconds = 0 and a positive nanos.
            nanos = optional(number)
          }))

          # Percentage of requests delayed (0.0-100.0).
          percentage = optional(number, 0)
        }))
      }))

      # Upper bound on how long a STREAM on this route may stay open
      # (gRPC/long-poll streams — distinct from timeout, which bounds a
      # request/response exchange). Live API truth: GCP rejects this field
      # unless the URL map's backend service uses the INTERNAL_SELF_MANAGED
      # (Traffic Director) load-balancing scheme ("Max stream duration is
      # only supported when UrlMap is used with BackendService whose Load
      # Balancing Scheme is INTERNAL_SELF_MANAGED") — leave it unset on
      # external application load balancers.
      max_stream_duration = optional(object({
        # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
        seconds = optional(number)

        # Fraction of a second at nanosecond resolution (0 to 999,999,999).
        # Durations under one second use seconds = 0 and a positive nanos.
        nanos = optional(number)
      }))

      # Cloud CDN caching for the routes using this action — overrides the
      # backend service's cdn_policy for matching traffic only. Takes effect
      # only when the target backend service (or bucket) has CDN enabled;
      # GCP ignores it otherwise.
      cache_policy = optional(object({
        # What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
        # content types and honors origin cache headers for the rest;
        # USE_ORIGIN_HEADERS caches only what the backends explicitly mark
        # cacheable (TTL fields must be unset — the origin controls lifetimes);
        # FORCE_CACHE_ALL caches everything, ignoring origin headers (never
        # combine with private or per-user content).
        cache_mode = optional(string, "")

        # Requests carrying any of these headers bypass the cache and go
        # straight to the origin (e.g. a debug or authorization header).
        cache_bypass_request_header_names = optional(list(string), [])

        # Cache negative responses (404, 410, ...) so repeated misses do not
        # hammer the backends. Pair with negative_caching_policy to set
        # per-status TTLs.
        negative_caching = optional(bool, false)

        # Collapse concurrent cache-fill requests for the same key into one
        # origin fetch. Default true on GCP's side.
        request_coalescing = optional(bool, false)

        # What forms the cache key — trim protocol, host, query string, or
        # select specific query parameters, headers, and cookies.
        cache_key_policy = optional(object({
          # Query parameters EXCLUDED from the cache key (everything else is
          # included). Mutually exclusive with included_query_parameters.
          excluded_query_parameters = optional(list(string), [])

          # Include the request host in the cache key (distinct hosts cache
          # separately).
          include_host = optional(bool, false)

          # Include the protocol (http/https) in the cache key.
          include_protocol = optional(bool, false)

          # Include the entire query string in the cache key. When false, the
          # included/excluded parameter lists refine what participates.
          include_query_string = optional(bool, false)

          # Cookie names whose values join the cache key.
          included_cookie_names = optional(list(string), [])

          # Header names whose values join the cache key.
          included_header_names = optional(list(string), [])

          # Query parameters INCLUDED in the cache key (everything else is
          # ignored). Mutually exclusive with excluded_query_parameters.
          included_query_parameters = optional(list(string), [])
        }))

        # Max lifetime in browsers and downstream caches (sets the max-age
        # clients see). Not respected with cache_mode USE_ORIGIN_HEADERS.
        client_ttl = optional(object({
          # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
          seconds = optional(number)

          # Fraction of a second at nanosecond resolution (0 to 999,999,999).
          # Durations under one second use seconds = 0 and a positive nanos.
          nanos = optional(number)
        }))

        # Edge-cache lifetime for responses without their own cache headers.
        # Not respected with cache_mode USE_ORIGIN_HEADERS.
        default_ttl = optional(object({
          # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
          seconds = optional(number)

          # Fraction of a second at nanosecond resolution (0 to 999,999,999).
          # Durations under one second use seconds = 0 and a positive nanos.
          nanos = optional(number)
        }))

        # Upper bound on any edge-cache lifetime, capping even origin-supplied
        # max-age. Not respected with cache_mode USE_ORIGIN_HEADERS or
        # FORCE_CACHE_ALL.
        max_ttl = optional(object({
          # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
          seconds = optional(number)

          # Fraction of a second at nanosecond resolution (0 to 999,999,999).
          # Durations under one second use seconds = 0 and a positive nanos.
          nanos = optional(number)
        }))

        # How long stale content may still be served while revalidating in the
        # background (up to 1 day).
        serve_while_stale = optional(object({
          # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
          seconds = optional(number)

          # Fraction of a second at nanosecond resolution (0 to 999,999,999).
          # Durations under one second use seconds = 0 and a positive nanos.
          nanos = optional(number)
        }))

        # Per-status TTLs for cached negative responses; only meaningful with
        # negative_caching enabled.
        negative_caching_policy = optional(list(object({
          # The HTTP status code this TTL applies to. GCP accepts only 300, 301,
          # 302, 307, 308, 404, 405, 410, 421, 451, and 501, each at most once.
          code = optional(number, 0)

          # How long responses with this status are cached.
          ttl = optional(object({
            # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
            seconds = optional(number)

            # Fraction of a second at nanosecond resolution (0 to 999,999,999).
            # Durations under one second use seconds = 0 and a positive nanos.
            nanos = optional(number)
          }))
        })), [])
      }))
    }))

    # Return a custom error page (from a backend bucket) for chosen response
    # codes at the top level. Global external Application Load Balancers only.
    default_custom_error_response_policy = optional(object({
      # The backend bucket serving the error pages. Reference a GcpBackendBucket
      # or provide a self-link directly.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      error_service = optional(string, "")

      # Rules mapping response codes to the error page and (optionally) an
      # overridden status code.
      error_response_rules = optional(list(object({
        # Response codes this rule matches: exact ("404", "503") or a class ("4xx",
        # "5xx"). At least one is required.
        match_response_codes = list(string)

        # Override the response code returned to the client (e.g. serve a 200 with a
        # maintenance page). Empty keeps the original code.
        override_response_code = optional(number, 0)

        # Path within the error backend bucket to serve for matched codes (e.g.
        # "/errors/404.html").
        path = string
      })), [])
    }))

    # Headers added to or removed from every request/response at the URL-map
    # level, before any per-route header action. Mutable.
    header_action = optional(object({
      # Headers to add to the request before it reaches the backend.
      request_headers_to_add = optional(list(object({
        # The header name (e.g. "X-Client-Geo").
        header_name = string

        # The header value. May use load-balancer variables (e.g. "{client_region}").
        header_value = string

        # Replace an existing header of the same name (true) or append to it
        # (false).
        replace = optional(bool, false)
      })), [])

      # Header names to strip from the request before it reaches the backend.
      request_headers_to_remove = optional(list(string), [])

      # Headers to add to the response before it returns to the client.
      response_headers_to_add = optional(list(object({
        # The header name (e.g. "X-Client-Geo").
        header_name = string

        # The header value. May use load-balancer variables (e.g. "{client_region}").
        header_value = string

        # Replace an existing header of the same name (true) or append to it
        # (false).
        replace = optional(bool, false)
      })), [])

      # Header names to strip from the response before it returns to the client.
      response_headers_to_remove = optional(list(string), [])
    }))

    # Map request Host headers to named path matchers. Each host rule points a
    # set of hosts (with optional wildcards) at one path_matcher by name.
    # Mutable.
    host_rules = optional(list(object({
      # Hosts this rule matches. A "*" wildcard may lead a domain
      # (e.g. "*.example.com") or match everything ("*"). At least one required.
      hosts = list(string)

      # The name of the path_matcher (in path_matchers) that handles these hosts.
      path_matcher = string

      # What this host rule covers — write it for the operator reading the routing
      # table later.
      description = optional(string, "")
    })), [])

    # The named path matchers host rules point at — each owns the path-level
    # routing (path_rules, route_rules, and its own default). Mutable.
    path_matchers = optional(list(object({
      # The path matcher's name, referenced by host_rules.path_matcher.
      name = string

      # The default target when no path_rule or route_rule matches — a backend
      # service or backend bucket. Reference a GcpBackendService or
      # GcpBackendBucket, or provide a self-link. Set exactly one of
      # default_service, default_url_redirect, or default_route_action.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      default_service = optional(string, "")

      # Redirect as the path matcher's default instead of serving.
      default_url_redirect = optional(object({
        # Replace the host in the redirect Location. Empty keeps the request host.
        host_redirect = optional(string, "")

        # Redirect to HTTPS (scheme becomes https). The standard http→https
        # upgrade. Default false.
        https_redirect = optional(bool, false)

        # Replace the entire path with this value. Mutually exclusive with
        # prefix_redirect. Empty keeps the request path.
        path_redirect = optional(string, "")

        # Replace the matched path prefix with this value, keeping the remainder.
        # Mutually exclusive with path_redirect.
        prefix_redirect = optional(string, "")

        # The HTTP redirect status code: FOUND (302), MOVED_PERMANENTLY_DEFAULT
        # (301), PERMANENT_REDIRECT (308), SEE_OTHER (303), or TEMPORARY_REDIRECT
        # (307). Empty uses the GCP default (MOVED_PERMANENTLY_DEFAULT).
        redirect_response_code = optional(string, "")

        # Drop the query string from the redirect Location. Default false (the query
        # string is preserved).
        strip_query = optional(bool, false)
      }))

      # Advanced default handling (weighted split / rewrite) for the path matcher.
      default_route_action = optional(object({
        # Split traffic across multiple backend services by weight — the mechanism
        # for weighted canary and blue/green rollouts. The weights are relative; a
        # backend's share is its weight over the sum of weights.
        weighted_backend_services = optional(list(object({
          # The backend service receiving this share of traffic. Reference a
          # GcpBackendService or provide a self-link directly.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          backend_service = string

          # Relative weight of this backend (0-1000). Its share is weight over the sum
          # of all weights in the split; 0 drains this backend from the split.
          weight = optional(number, 0)

          # Header mutations applied ONLY to the share of traffic sent to this
          # backend — e.g. tag canary responses with an identifying header so
          # clients and dashboards can tell which arm served them. Applied after
          # the URL-map-level and route-level header actions.
          header_action = optional(object({
            # Headers to add to the request before it reaches the backend.
            request_headers_to_add = optional(list(object({
              # The header name (e.g. "X-Client-Geo").
              header_name = string

              # The header value. May use load-balancer variables (e.g. "{client_region}").
              header_value = string

              # Replace an existing header of the same name (true) or append to it
              # (false).
              replace = optional(bool, false)
            })), [])

            # Header names to strip from the request before it reaches the backend.
            request_headers_to_remove = optional(list(string), [])

            # Headers to add to the response before it returns to the client.
            response_headers_to_add = optional(list(object({
              # The header name (e.g. "X-Client-Geo").
              header_name = string

              # The header value. May use load-balancer variables (e.g. "{client_region}").
              header_value = string

              # Replace an existing header of the same name (true) or append to it
              # (false).
              replace = optional(bool, false)
            })), [])

            # Header names to strip from the response before it returns to the client.
            response_headers_to_remove = optional(list(string), [])
          }))
        })), [])

        # Rewrite the host and/or path before forwarding to the backend.
        url_rewrite = optional(object({
          # Replace the request Host header with this value before forwarding.
          host_rewrite = optional(string, "")

          # Replace the matched path prefix with this value. Mutually exclusive with
          # path_template_rewrite.
          path_prefix_rewrite = optional(string, "")

          # Rewrite the path using a template that references named path variables
          # captured by a route rule's path_template_match (e.g. "/v2/{country}").
          # Honored only inside a route_rule's route_action — GCP rejects it in
          # default and path-rule route actions. Mutually exclusive with
          # path_prefix_rewrite.
          path_template_rewrite = optional(string, "")
        }))

        # Total time budget for the request, INCLUDING all retries (from the
        # first byte of the request to the last byte of the response). Pair with
        # retry_policy.per_try_timeout: per-try bounds one attempt, this bounds
        # the whole exchange. Unset uses the backend service's own timeout.
        # Not permitted when the route targets a redirect.
        timeout = optional(object({
          # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
          seconds = optional(number)

          # Fraction of a second at nanosecond resolution (0 to 999,999,999).
          # Durations under one second use seconds = 0 and a positive nanos.
          nanos = optional(number)
        }))

        # Retry failed requests to the backend. Which failures count is chosen
        # by retry_conditions; how long each attempt may run by per_try_timeout.
        # Retries consume the overall timeout's budget — they never extend it.
        retry_policy = optional(object({
          # Number of allowed retries (GCP defaults to 1 when unset/0). Each retry
          # still spends the route's overall timeout budget.
          num_retries = optional(number, 0)

          # Which failures trigger a retry. 5xx (any 5xx or no response at all),
          # gateway-error (502/503/504 only), connect-failure, retriable-4xx
          # (currently only 409), refused-stream, and the gRPC status conditions
          # cancelled, deadline-exceeded, resource-exhausted, unavailable.
          retry_conditions = optional(list(string), [])

          # Time budget for EACH retry attempt (the route's timeout bounds the
          # whole exchange across attempts). Unset uses the overall timeout for
          # every attempt.
          per_try_timeout = optional(object({
            # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
            seconds = optional(number)

            # Fraction of a second at nanosecond resolution (0 to 999,999,999).
            # Durations under one second use seconds = 0 and a positive nanos.
            nanos = optional(number)
          }))
        }))

        # Mirror every matched request to a second backend service, fire-and-
        # forget (responses from the mirror are discarded; the client sees only
        # the primary's response). Useful for shadow-testing a new stack with
        # production traffic — the mirror backend must be sized for the full
        # mirrored load.
        request_mirror_policy = optional(object({
          # The backend service receiving the mirrored copy of every matched
          # request. Reference a GcpBackendService or provide a self-link
          # directly. Responses from this backend are discarded.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          backend_service = string
        }))

        # Answer cross-origin (CORS) preflights and stamp CORS headers at the
        # load balancer, before requests reach the backend.
        cors_policy = optional(object({
          # Sets Access-Control-Allow-Credentials: allow requests carrying
          # credentials (cookies, authorization headers). Default false.
          allow_credentials = optional(bool, false)

          # Headers the client may send (Access-Control-Allow-Headers).
          allow_headers = optional(list(string), [])

          # Methods the client may use (Access-Control-Allow-Methods), e.g.
          # ["GET", "POST", "OPTIONS"].
          allow_methods = optional(list(string), [])

          # Regular expressions matching allowed origins; a request origin is
          # allowed when it matches any listed regex or exact allow_origins entry.
          allow_origin_regexes = optional(list(string), [])

          # Exact origins allowed, e.g. "https://app.example.com".
          allow_origins = optional(list(string), [])

          # Disable this CORS policy without deleting its configuration (true
          # turns the policy off). Default false — the policy is in effect.
          disabled = optional(bool, false)

          # Headers the browser may read from the response
          # (Access-Control-Expose-Headers).
          expose_headers = optional(list(string), [])

          # Seconds a browser may cache the preflight response
          # (Access-Control-Max-Age).
          max_age = optional(number, 0)
        }))

        # Deliberately inject failures (aborts with a chosen status, fixed
        # delays) into a percentage of matched requests — chaos/resilience
        # testing of the CLIENTS of this route. Never leave enabled on a
        # production default route: the injected failures are real to callers.
        fault_injection_policy = optional(object({
          # Abort a percentage of requests with a fixed HTTP status, before they
          # reach the backend.
          abort = optional(object({
            # HTTP status returned to aborted requests (200-599).
            http_status = optional(number, 0)

            # Percentage of requests aborted (0.0-100.0).
            percentage = optional(number, 0)
          }))

          # Delay a percentage of requests by a fixed duration before forwarding.
          delay = optional(object({
            # How long delayed requests are held before forwarding.
            fixed_delay = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))

            # Percentage of requests delayed (0.0-100.0).
            percentage = optional(number, 0)
          }))
        }))

        # Upper bound on how long a STREAM on this route may stay open
        # (gRPC/long-poll streams — distinct from timeout, which bounds a
        # request/response exchange). Live API truth: GCP rejects this field
        # unless the URL map's backend service uses the INTERNAL_SELF_MANAGED
        # (Traffic Director) load-balancing scheme ("Max stream duration is
        # only supported when UrlMap is used with BackendService whose Load
        # Balancing Scheme is INTERNAL_SELF_MANAGED") — leave it unset on
        # external application load balancers.
        max_stream_duration = optional(object({
          # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
          seconds = optional(number)

          # Fraction of a second at nanosecond resolution (0 to 999,999,999).
          # Durations under one second use seconds = 0 and a positive nanos.
          nanos = optional(number)
        }))

        # Cloud CDN caching for the routes using this action — overrides the
        # backend service's cdn_policy for matching traffic only. Takes effect
        # only when the target backend service (or bucket) has CDN enabled;
        # GCP ignores it otherwise.
        cache_policy = optional(object({
          # What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
          # content types and honors origin cache headers for the rest;
          # USE_ORIGIN_HEADERS caches only what the backends explicitly mark
          # cacheable (TTL fields must be unset — the origin controls lifetimes);
          # FORCE_CACHE_ALL caches everything, ignoring origin headers (never
          # combine with private or per-user content).
          cache_mode = optional(string, "")

          # Requests carrying any of these headers bypass the cache and go
          # straight to the origin (e.g. a debug or authorization header).
          cache_bypass_request_header_names = optional(list(string), [])

          # Cache negative responses (404, 410, ...) so repeated misses do not
          # hammer the backends. Pair with negative_caching_policy to set
          # per-status TTLs.
          negative_caching = optional(bool, false)

          # Collapse concurrent cache-fill requests for the same key into one
          # origin fetch. Default true on GCP's side.
          request_coalescing = optional(bool, false)

          # What forms the cache key — trim protocol, host, query string, or
          # select specific query parameters, headers, and cookies.
          cache_key_policy = optional(object({
            # Query parameters EXCLUDED from the cache key (everything else is
            # included). Mutually exclusive with included_query_parameters.
            excluded_query_parameters = optional(list(string), [])

            # Include the request host in the cache key (distinct hosts cache
            # separately).
            include_host = optional(bool, false)

            # Include the protocol (http/https) in the cache key.
            include_protocol = optional(bool, false)

            # Include the entire query string in the cache key. When false, the
            # included/excluded parameter lists refine what participates.
            include_query_string = optional(bool, false)

            # Cookie names whose values join the cache key.
            included_cookie_names = optional(list(string), [])

            # Header names whose values join the cache key.
            included_header_names = optional(list(string), [])

            # Query parameters INCLUDED in the cache key (everything else is
            # ignored). Mutually exclusive with excluded_query_parameters.
            included_query_parameters = optional(list(string), [])
          }))

          # Max lifetime in browsers and downstream caches (sets the max-age
          # clients see). Not respected with cache_mode USE_ORIGIN_HEADERS.
          client_ttl = optional(object({
            # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
            seconds = optional(number)

            # Fraction of a second at nanosecond resolution (0 to 999,999,999).
            # Durations under one second use seconds = 0 and a positive nanos.
            nanos = optional(number)
          }))

          # Edge-cache lifetime for responses without their own cache headers.
          # Not respected with cache_mode USE_ORIGIN_HEADERS.
          default_ttl = optional(object({
            # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
            seconds = optional(number)

            # Fraction of a second at nanosecond resolution (0 to 999,999,999).
            # Durations under one second use seconds = 0 and a positive nanos.
            nanos = optional(number)
          }))

          # Upper bound on any edge-cache lifetime, capping even origin-supplied
          # max-age. Not respected with cache_mode USE_ORIGIN_HEADERS or
          # FORCE_CACHE_ALL.
          max_ttl = optional(object({
            # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
            seconds = optional(number)

            # Fraction of a second at nanosecond resolution (0 to 999,999,999).
            # Durations under one second use seconds = 0 and a positive nanos.
            nanos = optional(number)
          }))

          # How long stale content may still be served while revalidating in the
          # background (up to 1 day).
          serve_while_stale = optional(object({
            # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
            seconds = optional(number)

            # Fraction of a second at nanosecond resolution (0 to 999,999,999).
            # Durations under one second use seconds = 0 and a positive nanos.
            nanos = optional(number)
          }))

          # Per-status TTLs for cached negative responses; only meaningful with
          # negative_caching enabled.
          negative_caching_policy = optional(list(object({
            # The HTTP status code this TTL applies to. GCP accepts only 300, 301,
            # 302, 307, 308, 404, 405, 410, 421, 451, and 501, each at most once.
            code = optional(number, 0)

            # How long responses with this status are cached.
            ttl = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))
          })), [])
        }))
      }))

      # Custom error pages for this path matcher's default. Global external ALBs
      # only.
      default_custom_error_response_policy = optional(object({
        # The backend bucket serving the error pages. Reference a GcpBackendBucket
        # or provide a self-link directly.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        error_service = optional(string, "")

        # Rules mapping response codes to the error page and (optionally) an
        # overridden status code.
        error_response_rules = optional(list(object({
          # Response codes this rule matches: exact ("404", "503") or a class ("4xx",
          # "5xx"). At least one is required.
          match_response_codes = list(string)

          # Override the response code returned to the client (e.g. serve a 200 with a
          # maintenance page). Empty keeps the original code.
          override_response_code = optional(number, 0)

          # Path within the error backend bucket to serve for matched codes (e.g.
          # "/errors/404.html").
          path = string
        })), [])
      }))

      # What this path matcher covers.
      description = optional(string, "")

      # Header mutations applied to all traffic through this path matcher.
      header_action = optional(object({
        # Headers to add to the request before it reaches the backend.
        request_headers_to_add = optional(list(object({
          # The header name (e.g. "X-Client-Geo").
          header_name = string

          # The header value. May use load-balancer variables (e.g. "{client_region}").
          header_value = string

          # Replace an existing header of the same name (true) or append to it
          # (false).
          replace = optional(bool, false)
        })), [])

        # Header names to strip from the request before it reaches the backend.
        request_headers_to_remove = optional(list(string), [])

        # Headers to add to the response before it returns to the client.
        response_headers_to_add = optional(list(object({
          # The header name (e.g. "X-Client-Geo").
          header_name = string

          # The header value. May use load-balancer variables (e.g. "{client_region}").
          header_value = string

          # Replace an existing header of the same name (true) or append to it
          # (false).
          replace = optional(bool, false)
        })), [])

        # Header names to strip from the response before it returns to the client.
        response_headers_to_remove = optional(list(string), [])
      }))

      # Longest-prefix path rules: each maps a set of path patterns to a service,
      # redirect, or route action. Evaluated after route_rules.
      path_rules = optional(list(object({
        # Path patterns this rule matches. Each must start with "/" and may end with
        # a single "*" wildcard (e.g. "/api/*"). At least one required.
        paths = list(string)

        # The target when a path matches — a backend service or backend bucket.
        # Reference a GcpBackendService or GcpBackendBucket, or provide a self-link.
        # Set exactly one of service, url_redirect, or route_action.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service = optional(string, "")

        # Advanced handling (weighted split / rewrite) for matched paths.
        route_action = optional(object({
          # Split traffic across multiple backend services by weight — the mechanism
          # for weighted canary and blue/green rollouts. The weights are relative; a
          # backend's share is its weight over the sum of weights.
          weighted_backend_services = optional(list(object({
            # The backend service receiving this share of traffic. Reference a
            # GcpBackendService or provide a self-link directly.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            backend_service = string

            # Relative weight of this backend (0-1000). Its share is weight over the sum
            # of all weights in the split; 0 drains this backend from the split.
            weight = optional(number, 0)

            # Header mutations applied ONLY to the share of traffic sent to this
            # backend — e.g. tag canary responses with an identifying header so
            # clients and dashboards can tell which arm served them. Applied after
            # the URL-map-level and route-level header actions.
            header_action = optional(object({
              # Headers to add to the request before it reaches the backend.
              request_headers_to_add = optional(list(object({
                # The header name (e.g. "X-Client-Geo").
                header_name = string

                # The header value. May use load-balancer variables (e.g. "{client_region}").
                header_value = string

                # Replace an existing header of the same name (true) or append to it
                # (false).
                replace = optional(bool, false)
              })), [])

              # Header names to strip from the request before it reaches the backend.
              request_headers_to_remove = optional(list(string), [])

              # Headers to add to the response before it returns to the client.
              response_headers_to_add = optional(list(object({
                # The header name (e.g. "X-Client-Geo").
                header_name = string

                # The header value. May use load-balancer variables (e.g. "{client_region}").
                header_value = string

                # Replace an existing header of the same name (true) or append to it
                # (false).
                replace = optional(bool, false)
              })), [])

              # Header names to strip from the response before it returns to the client.
              response_headers_to_remove = optional(list(string), [])
            }))
          })), [])

          # Rewrite the host and/or path before forwarding to the backend.
          url_rewrite = optional(object({
            # Replace the request Host header with this value before forwarding.
            host_rewrite = optional(string, "")

            # Replace the matched path prefix with this value. Mutually exclusive with
            # path_template_rewrite.
            path_prefix_rewrite = optional(string, "")

            # Rewrite the path using a template that references named path variables
            # captured by a route rule's path_template_match (e.g. "/v2/{country}").
            # Honored only inside a route_rule's route_action — GCP rejects it in
            # default and path-rule route actions. Mutually exclusive with
            # path_prefix_rewrite.
            path_template_rewrite = optional(string, "")
          }))

          # Total time budget for the request, INCLUDING all retries (from the
          # first byte of the request to the last byte of the response). Pair with
          # retry_policy.per_try_timeout: per-try bounds one attempt, this bounds
          # the whole exchange. Unset uses the backend service's own timeout.
          # Not permitted when the route targets a redirect.
          timeout = optional(object({
            # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
            seconds = optional(number)

            # Fraction of a second at nanosecond resolution (0 to 999,999,999).
            # Durations under one second use seconds = 0 and a positive nanos.
            nanos = optional(number)
          }))

          # Retry failed requests to the backend. Which failures count is chosen
          # by retry_conditions; how long each attempt may run by per_try_timeout.
          # Retries consume the overall timeout's budget — they never extend it.
          retry_policy = optional(object({
            # Number of allowed retries (GCP defaults to 1 when unset/0). Each retry
            # still spends the route's overall timeout budget.
            num_retries = optional(number, 0)

            # Which failures trigger a retry. 5xx (any 5xx or no response at all),
            # gateway-error (502/503/504 only), connect-failure, retriable-4xx
            # (currently only 409), refused-stream, and the gRPC status conditions
            # cancelled, deadline-exceeded, resource-exhausted, unavailable.
            retry_conditions = optional(list(string), [])

            # Time budget for EACH retry attempt (the route's timeout bounds the
            # whole exchange across attempts). Unset uses the overall timeout for
            # every attempt.
            per_try_timeout = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))
          }))

          # Mirror every matched request to a second backend service, fire-and-
          # forget (responses from the mirror are discarded; the client sees only
          # the primary's response). Useful for shadow-testing a new stack with
          # production traffic — the mirror backend must be sized for the full
          # mirrored load.
          request_mirror_policy = optional(object({
            # The backend service receiving the mirrored copy of every matched
            # request. Reference a GcpBackendService or provide a self-link
            # directly. Responses from this backend are discarded.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            backend_service = string
          }))

          # Answer cross-origin (CORS) preflights and stamp CORS headers at the
          # load balancer, before requests reach the backend.
          cors_policy = optional(object({
            # Sets Access-Control-Allow-Credentials: allow requests carrying
            # credentials (cookies, authorization headers). Default false.
            allow_credentials = optional(bool, false)

            # Headers the client may send (Access-Control-Allow-Headers).
            allow_headers = optional(list(string), [])

            # Methods the client may use (Access-Control-Allow-Methods), e.g.
            # ["GET", "POST", "OPTIONS"].
            allow_methods = optional(list(string), [])

            # Regular expressions matching allowed origins; a request origin is
            # allowed when it matches any listed regex or exact allow_origins entry.
            allow_origin_regexes = optional(list(string), [])

            # Exact origins allowed, e.g. "https://app.example.com".
            allow_origins = optional(list(string), [])

            # Disable this CORS policy without deleting its configuration (true
            # turns the policy off). Default false — the policy is in effect.
            disabled = optional(bool, false)

            # Headers the browser may read from the response
            # (Access-Control-Expose-Headers).
            expose_headers = optional(list(string), [])

            # Seconds a browser may cache the preflight response
            # (Access-Control-Max-Age).
            max_age = optional(number, 0)
          }))

          # Deliberately inject failures (aborts with a chosen status, fixed
          # delays) into a percentage of matched requests — chaos/resilience
          # testing of the CLIENTS of this route. Never leave enabled on a
          # production default route: the injected failures are real to callers.
          fault_injection_policy = optional(object({
            # Abort a percentage of requests with a fixed HTTP status, before they
            # reach the backend.
            abort = optional(object({
              # HTTP status returned to aborted requests (200-599).
              http_status = optional(number, 0)

              # Percentage of requests aborted (0.0-100.0).
              percentage = optional(number, 0)
            }))

            # Delay a percentage of requests by a fixed duration before forwarding.
            delay = optional(object({
              # How long delayed requests are held before forwarding.
              fixed_delay = optional(object({
                # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
                seconds = optional(number)

                # Fraction of a second at nanosecond resolution (0 to 999,999,999).
                # Durations under one second use seconds = 0 and a positive nanos.
                nanos = optional(number)
              }))

              # Percentage of requests delayed (0.0-100.0).
              percentage = optional(number, 0)
            }))
          }))

          # Upper bound on how long a STREAM on this route may stay open
          # (gRPC/long-poll streams — distinct from timeout, which bounds a
          # request/response exchange). Live API truth: GCP rejects this field
          # unless the URL map's backend service uses the INTERNAL_SELF_MANAGED
          # (Traffic Director) load-balancing scheme ("Max stream duration is
          # only supported when UrlMap is used with BackendService whose Load
          # Balancing Scheme is INTERNAL_SELF_MANAGED") — leave it unset on
          # external application load balancers.
          max_stream_duration = optional(object({
            # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
            seconds = optional(number)

            # Fraction of a second at nanosecond resolution (0 to 999,999,999).
            # Durations under one second use seconds = 0 and a positive nanos.
            nanos = optional(number)
          }))

          # Cloud CDN caching for the routes using this action — overrides the
          # backend service's cdn_policy for matching traffic only. Takes effect
          # only when the target backend service (or bucket) has CDN enabled;
          # GCP ignores it otherwise.
          cache_policy = optional(object({
            # What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
            # content types and honors origin cache headers for the rest;
            # USE_ORIGIN_HEADERS caches only what the backends explicitly mark
            # cacheable (TTL fields must be unset — the origin controls lifetimes);
            # FORCE_CACHE_ALL caches everything, ignoring origin headers (never
            # combine with private or per-user content).
            cache_mode = optional(string, "")

            # Requests carrying any of these headers bypass the cache and go
            # straight to the origin (e.g. a debug or authorization header).
            cache_bypass_request_header_names = optional(list(string), [])

            # Cache negative responses (404, 410, ...) so repeated misses do not
            # hammer the backends. Pair with negative_caching_policy to set
            # per-status TTLs.
            negative_caching = optional(bool, false)

            # Collapse concurrent cache-fill requests for the same key into one
            # origin fetch. Default true on GCP's side.
            request_coalescing = optional(bool, false)

            # What forms the cache key — trim protocol, host, query string, or
            # select specific query parameters, headers, and cookies.
            cache_key_policy = optional(object({
              # Query parameters EXCLUDED from the cache key (everything else is
              # included). Mutually exclusive with included_query_parameters.
              excluded_query_parameters = optional(list(string), [])

              # Include the request host in the cache key (distinct hosts cache
              # separately).
              include_host = optional(bool, false)

              # Include the protocol (http/https) in the cache key.
              include_protocol = optional(bool, false)

              # Include the entire query string in the cache key. When false, the
              # included/excluded parameter lists refine what participates.
              include_query_string = optional(bool, false)

              # Cookie names whose values join the cache key.
              included_cookie_names = optional(list(string), [])

              # Header names whose values join the cache key.
              included_header_names = optional(list(string), [])

              # Query parameters INCLUDED in the cache key (everything else is
              # ignored). Mutually exclusive with excluded_query_parameters.
              included_query_parameters = optional(list(string), [])
            }))

            # Max lifetime in browsers and downstream caches (sets the max-age
            # clients see). Not respected with cache_mode USE_ORIGIN_HEADERS.
            client_ttl = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))

            # Edge-cache lifetime for responses without their own cache headers.
            # Not respected with cache_mode USE_ORIGIN_HEADERS.
            default_ttl = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))

            # Upper bound on any edge-cache lifetime, capping even origin-supplied
            # max-age. Not respected with cache_mode USE_ORIGIN_HEADERS or
            # FORCE_CACHE_ALL.
            max_ttl = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))

            # How long stale content may still be served while revalidating in the
            # background (up to 1 day).
            serve_while_stale = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))

            # Per-status TTLs for cached negative responses; only meaningful with
            # negative_caching enabled.
            negative_caching_policy = optional(list(object({
              # The HTTP status code this TTL applies to. GCP accepts only 300, 301,
              # 302, 307, 308, 404, 405, 410, 421, 451, and 501, each at most once.
              code = optional(number, 0)

              # How long responses with this status are cached.
              ttl = optional(object({
                # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
                seconds = optional(number)

                # Fraction of a second at nanosecond resolution (0 to 999,999,999).
                # Durations under one second use seconds = 0 and a positive nanos.
                nanos = optional(number)
              }))
            })), [])
          }))
        }))

        # Redirect matched paths instead of serving them.
        url_redirect = optional(object({
          # Replace the host in the redirect Location. Empty keeps the request host.
          host_redirect = optional(string, "")

          # Redirect to HTTPS (scheme becomes https). The standard http→https
          # upgrade. Default false.
          https_redirect = optional(bool, false)

          # Replace the entire path with this value. Mutually exclusive with
          # prefix_redirect. Empty keeps the request path.
          path_redirect = optional(string, "")

          # Replace the matched path prefix with this value, keeping the remainder.
          # Mutually exclusive with path_redirect.
          prefix_redirect = optional(string, "")

          # The HTTP redirect status code: FOUND (302), MOVED_PERMANENTLY_DEFAULT
          # (301), PERMANENT_REDIRECT (308), SEE_OTHER (303), or TEMPORARY_REDIRECT
          # (307). Empty uses the GCP default (MOVED_PERMANENTLY_DEFAULT).
          redirect_response_code = optional(string, "")

          # Drop the query string from the redirect Location. Default false (the query
          # string is preserved).
          strip_query = optional(bool, false)
        }))

        # Custom error pages for matched paths. Global external ALBs only.
        custom_error_response_policy = optional(object({
          # The backend bucket serving the error pages. Reference a GcpBackendBucket
          # or provide a self-link directly.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          error_service = optional(string, "")

          # Rules mapping response codes to the error page and (optionally) an
          # overridden status code.
          error_response_rules = optional(list(object({
            # Response codes this rule matches: exact ("404", "503") or a class ("4xx",
            # "5xx"). At least one is required.
            match_response_codes = list(string)

            # Override the response code returned to the client (e.g. serve a 200 with a
            # maintenance page). Empty keeps the original code.
            override_response_code = optional(number, 0)

            # Path within the error backend bucket to serve for matched codes (e.g.
            # "/errors/404.html").
            path = string
          })), [])
        }))
      })), [])

      # Priority-ordered route rules with rich header/query/path matching.
      # Evaluated before path_rules. A path matcher uses either path_rules or
      # route_rules, not both.
      route_rules = optional(list(object({
        # Evaluation priority (0 to 2147483647); lower numbers are evaluated first
        # and must be unique within the path matcher. Proto3 int32 has no presence,
        # so required is omitted — priority 0 is valid and must not be rejected as
        # "unset".
        priority = optional(number, 0)

        # The target backend service when this rule matches. Reference a
        # GcpBackendService or provide a self-link. Route rules target backend
        # services only (not buckets). Set exactly one of service, url_redirect, or
        # route_action.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service = optional(string, "")

        # Conditions a request must meet for this rule to fire (path/header/query
        # matching). At least one match rule is required.
        match_rules = list(object({
          # Match when the path starts with this prefix (e.g. "/api"). Mutually
          # exclusive with full_path_match, regex_match, and path_template_match.
          prefix_match = optional(string, "")

          # Match when the path equals this exactly. Mutually exclusive with the other
          # path matchers.
          full_path_match = optional(string, "")

          # Match the path against this regular expression. Mutually exclusive with
          # the other path matchers.
          regex_match = optional(string, "")

          # Match the path against a wildcard template capturing named variables
          # (e.g. "/v1/{country}/**"), usable by a route_action's
          # path_template_rewrite. Mutually exclusive with the other path matchers.
          path_template_match = optional(string, "")

          # Case-insensitive path matching. Default false.
          ignore_case = optional(bool, false)

          # Match on request headers — all listed header matches must hold.
          header_matches = optional(list(object({
            # The header name to match.
            header_name = string

            # Match when the header equals this exactly.
            exact_match = optional(string, "")

            # Match when the header starts with this.
            prefix_match = optional(string, "")

            # Match when the header ends with this.
            suffix_match = optional(string, "")

            # Match the header against this regular expression.
            regex_match = optional(string, "")

            # Match when the header is present (any value).
            present_match = optional(bool, false)

            # Match when the header's integer value falls in a range.
            range_match = optional(object({
              # Inclusive lower bound.
              range_start = optional(number, 0)

              # Exclusive upper bound.
              range_end = optional(number, 0)
            }))

            # Invert the whole match — fire when the condition does NOT hold. Default
            # false.
            invert_match = optional(bool, false)
          })), [])

          # Match on query parameters — all listed matches must hold.
          query_parameter_matches = optional(list(object({
            # The query parameter name to match.
            name = string

            # Match when the parameter equals this exactly.
            exact_match = optional(string, "")

            # Match when the parameter is present (any value).
            present_match = optional(bool, false)

            # Match the parameter against this regular expression.
            regex_match = optional(string, "")
          })), [])

          # Traffic Director metadata filters (xDS node metadata matching).
          metadata_filters = optional(list(object({
            # How the labels combine: MATCH_ALL (every label must match) or MATCH_ANY
            # (at least one).
            filter_match_criteria = string

            # The xDS node metadata labels to match against (1-64 entries).
            filter_labels = list(object({
              # Label name.
              name = string

              # Label value.
              value = string
            }))
          })), [])
        }))

        # Advanced handling (weighted split / rewrite / retry) for matched requests.
        route_action = optional(object({
          # Split traffic across multiple backend services by weight — the mechanism
          # for weighted canary and blue/green rollouts. The weights are relative; a
          # backend's share is its weight over the sum of weights.
          weighted_backend_services = optional(list(object({
            # The backend service receiving this share of traffic. Reference a
            # GcpBackendService or provide a self-link directly.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            backend_service = string

            # Relative weight of this backend (0-1000). Its share is weight over the sum
            # of all weights in the split; 0 drains this backend from the split.
            weight = optional(number, 0)

            # Header mutations applied ONLY to the share of traffic sent to this
            # backend — e.g. tag canary responses with an identifying header so
            # clients and dashboards can tell which arm served them. Applied after
            # the URL-map-level and route-level header actions.
            header_action = optional(object({
              # Headers to add to the request before it reaches the backend.
              request_headers_to_add = optional(list(object({
                # The header name (e.g. "X-Client-Geo").
                header_name = string

                # The header value. May use load-balancer variables (e.g. "{client_region}").
                header_value = string

                # Replace an existing header of the same name (true) or append to it
                # (false).
                replace = optional(bool, false)
              })), [])

              # Header names to strip from the request before it reaches the backend.
              request_headers_to_remove = optional(list(string), [])

              # Headers to add to the response before it returns to the client.
              response_headers_to_add = optional(list(object({
                # The header name (e.g. "X-Client-Geo").
                header_name = string

                # The header value. May use load-balancer variables (e.g. "{client_region}").
                header_value = string

                # Replace an existing header of the same name (true) or append to it
                # (false).
                replace = optional(bool, false)
              })), [])

              # Header names to strip from the response before it returns to the client.
              response_headers_to_remove = optional(list(string), [])
            }))
          })), [])

          # Rewrite the host and/or path before forwarding to the backend.
          url_rewrite = optional(object({
            # Replace the request Host header with this value before forwarding.
            host_rewrite = optional(string, "")

            # Replace the matched path prefix with this value. Mutually exclusive with
            # path_template_rewrite.
            path_prefix_rewrite = optional(string, "")

            # Rewrite the path using a template that references named path variables
            # captured by a route rule's path_template_match (e.g. "/v2/{country}").
            # Honored only inside a route_rule's route_action — GCP rejects it in
            # default and path-rule route actions. Mutually exclusive with
            # path_prefix_rewrite.
            path_template_rewrite = optional(string, "")
          }))

          # Total time budget for the request, INCLUDING all retries (from the
          # first byte of the request to the last byte of the response). Pair with
          # retry_policy.per_try_timeout: per-try bounds one attempt, this bounds
          # the whole exchange. Unset uses the backend service's own timeout.
          # Not permitted when the route targets a redirect.
          timeout = optional(object({
            # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
            seconds = optional(number)

            # Fraction of a second at nanosecond resolution (0 to 999,999,999).
            # Durations under one second use seconds = 0 and a positive nanos.
            nanos = optional(number)
          }))

          # Retry failed requests to the backend. Which failures count is chosen
          # by retry_conditions; how long each attempt may run by per_try_timeout.
          # Retries consume the overall timeout's budget — they never extend it.
          retry_policy = optional(object({
            # Number of allowed retries (GCP defaults to 1 when unset/0). Each retry
            # still spends the route's overall timeout budget.
            num_retries = optional(number, 0)

            # Which failures trigger a retry. 5xx (any 5xx or no response at all),
            # gateway-error (502/503/504 only), connect-failure, retriable-4xx
            # (currently only 409), refused-stream, and the gRPC status conditions
            # cancelled, deadline-exceeded, resource-exhausted, unavailable.
            retry_conditions = optional(list(string), [])

            # Time budget for EACH retry attempt (the route's timeout bounds the
            # whole exchange across attempts). Unset uses the overall timeout for
            # every attempt.
            per_try_timeout = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))
          }))

          # Mirror every matched request to a second backend service, fire-and-
          # forget (responses from the mirror are discarded; the client sees only
          # the primary's response). Useful for shadow-testing a new stack with
          # production traffic — the mirror backend must be sized for the full
          # mirrored load.
          request_mirror_policy = optional(object({
            # The backend service receiving the mirrored copy of every matched
            # request. Reference a GcpBackendService or provide a self-link
            # directly. Responses from this backend are discarded.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            backend_service = string
          }))

          # Answer cross-origin (CORS) preflights and stamp CORS headers at the
          # load balancer, before requests reach the backend.
          cors_policy = optional(object({
            # Sets Access-Control-Allow-Credentials: allow requests carrying
            # credentials (cookies, authorization headers). Default false.
            allow_credentials = optional(bool, false)

            # Headers the client may send (Access-Control-Allow-Headers).
            allow_headers = optional(list(string), [])

            # Methods the client may use (Access-Control-Allow-Methods), e.g.
            # ["GET", "POST", "OPTIONS"].
            allow_methods = optional(list(string), [])

            # Regular expressions matching allowed origins; a request origin is
            # allowed when it matches any listed regex or exact allow_origins entry.
            allow_origin_regexes = optional(list(string), [])

            # Exact origins allowed, e.g. "https://app.example.com".
            allow_origins = optional(list(string), [])

            # Disable this CORS policy without deleting its configuration (true
            # turns the policy off). Default false — the policy is in effect.
            disabled = optional(bool, false)

            # Headers the browser may read from the response
            # (Access-Control-Expose-Headers).
            expose_headers = optional(list(string), [])

            # Seconds a browser may cache the preflight response
            # (Access-Control-Max-Age).
            max_age = optional(number, 0)
          }))

          # Deliberately inject failures (aborts with a chosen status, fixed
          # delays) into a percentage of matched requests — chaos/resilience
          # testing of the CLIENTS of this route. Never leave enabled on a
          # production default route: the injected failures are real to callers.
          fault_injection_policy = optional(object({
            # Abort a percentage of requests with a fixed HTTP status, before they
            # reach the backend.
            abort = optional(object({
              # HTTP status returned to aborted requests (200-599).
              http_status = optional(number, 0)

              # Percentage of requests aborted (0.0-100.0).
              percentage = optional(number, 0)
            }))

            # Delay a percentage of requests by a fixed duration before forwarding.
            delay = optional(object({
              # How long delayed requests are held before forwarding.
              fixed_delay = optional(object({
                # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
                seconds = optional(number)

                # Fraction of a second at nanosecond resolution (0 to 999,999,999).
                # Durations under one second use seconds = 0 and a positive nanos.
                nanos = optional(number)
              }))

              # Percentage of requests delayed (0.0-100.0).
              percentage = optional(number, 0)
            }))
          }))

          # Upper bound on how long a STREAM on this route may stay open
          # (gRPC/long-poll streams — distinct from timeout, which bounds a
          # request/response exchange). Live API truth: GCP rejects this field
          # unless the URL map's backend service uses the INTERNAL_SELF_MANAGED
          # (Traffic Director) load-balancing scheme ("Max stream duration is
          # only supported when UrlMap is used with BackendService whose Load
          # Balancing Scheme is INTERNAL_SELF_MANAGED") — leave it unset on
          # external application load balancers.
          max_stream_duration = optional(object({
            # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
            seconds = optional(number)

            # Fraction of a second at nanosecond resolution (0 to 999,999,999).
            # Durations under one second use seconds = 0 and a positive nanos.
            nanos = optional(number)
          }))

          # Cloud CDN caching for the routes using this action — overrides the
          # backend service's cdn_policy for matching traffic only. Takes effect
          # only when the target backend service (or bucket) has CDN enabled;
          # GCP ignores it otherwise.
          cache_policy = optional(object({
            # What gets cached. CACHE_ALL_STATIC (the GCP default) caches static
            # content types and honors origin cache headers for the rest;
            # USE_ORIGIN_HEADERS caches only what the backends explicitly mark
            # cacheable (TTL fields must be unset — the origin controls lifetimes);
            # FORCE_CACHE_ALL caches everything, ignoring origin headers (never
            # combine with private or per-user content).
            cache_mode = optional(string, "")

            # Requests carrying any of these headers bypass the cache and go
            # straight to the origin (e.g. a debug or authorization header).
            cache_bypass_request_header_names = optional(list(string), [])

            # Cache negative responses (404, 410, ...) so repeated misses do not
            # hammer the backends. Pair with negative_caching_policy to set
            # per-status TTLs.
            negative_caching = optional(bool, false)

            # Collapse concurrent cache-fill requests for the same key into one
            # origin fetch. Default true on GCP's side.
            request_coalescing = optional(bool, false)

            # What forms the cache key — trim protocol, host, query string, or
            # select specific query parameters, headers, and cookies.
            cache_key_policy = optional(object({
              # Query parameters EXCLUDED from the cache key (everything else is
              # included). Mutually exclusive with included_query_parameters.
              excluded_query_parameters = optional(list(string), [])

              # Include the request host in the cache key (distinct hosts cache
              # separately).
              include_host = optional(bool, false)

              # Include the protocol (http/https) in the cache key.
              include_protocol = optional(bool, false)

              # Include the entire query string in the cache key. When false, the
              # included/excluded parameter lists refine what participates.
              include_query_string = optional(bool, false)

              # Cookie names whose values join the cache key.
              included_cookie_names = optional(list(string), [])

              # Header names whose values join the cache key.
              included_header_names = optional(list(string), [])

              # Query parameters INCLUDED in the cache key (everything else is
              # ignored). Mutually exclusive with excluded_query_parameters.
              included_query_parameters = optional(list(string), [])
            }))

            # Max lifetime in browsers and downstream caches (sets the max-age
            # clients see). Not respected with cache_mode USE_ORIGIN_HEADERS.
            client_ttl = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))

            # Edge-cache lifetime for responses without their own cache headers.
            # Not respected with cache_mode USE_ORIGIN_HEADERS.
            default_ttl = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))

            # Upper bound on any edge-cache lifetime, capping even origin-supplied
            # max-age. Not respected with cache_mode USE_ORIGIN_HEADERS or
            # FORCE_CACHE_ALL.
            max_ttl = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))

            # How long stale content may still be served while revalidating in the
            # background (up to 1 day).
            serve_while_stale = optional(object({
              # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
              seconds = optional(number)

              # Fraction of a second at nanosecond resolution (0 to 999,999,999).
              # Durations under one second use seconds = 0 and a positive nanos.
              nanos = optional(number)
            }))

            # Per-status TTLs for cached negative responses; only meaningful with
            # negative_caching enabled.
            negative_caching_policy = optional(list(object({
              # The HTTP status code this TTL applies to. GCP accepts only 300, 301,
              # 302, 307, 308, 404, 405, 410, 421, 451, and 501, each at most once.
              code = optional(number, 0)

              # How long responses with this status are cached.
              ttl = optional(object({
                # Whole seconds (0 to 315,576,000,000 — GCP's int64 Duration bound).
                seconds = optional(number)

                # Fraction of a second at nanosecond resolution (0 to 999,999,999).
                # Durations under one second use seconds = 0 and a positive nanos.
                nanos = optional(number)
              }))
            })), [])
          }))
        }))

        # Redirect matched requests instead of serving them.
        url_redirect = optional(object({
          # Replace the host in the redirect Location. Empty keeps the request host.
          host_redirect = optional(string, "")

          # Redirect to HTTPS (scheme becomes https). The standard http→https
          # upgrade. Default false.
          https_redirect = optional(bool, false)

          # Replace the entire path with this value. Mutually exclusive with
          # prefix_redirect. Empty keeps the request path.
          path_redirect = optional(string, "")

          # Replace the matched path prefix with this value, keeping the remainder.
          # Mutually exclusive with path_redirect.
          prefix_redirect = optional(string, "")

          # The HTTP redirect status code: FOUND (302), MOVED_PERMANENTLY_DEFAULT
          # (301), PERMANENT_REDIRECT (308), SEE_OTHER (303), or TEMPORARY_REDIRECT
          # (307). Empty uses the GCP default (MOVED_PERMANENTLY_DEFAULT).
          redirect_response_code = optional(string, "")

          # Drop the query string from the redirect Location. Default false (the query
          # string is preserved).
          strip_query = optional(bool, false)
        }))

        # Header mutations applied to matched requests.
        header_action = optional(object({
          # Headers to add to the request before it reaches the backend.
          request_headers_to_add = optional(list(object({
            # The header name (e.g. "X-Client-Geo").
            header_name = string

            # The header value. May use load-balancer variables (e.g. "{client_region}").
            header_value = string

            # Replace an existing header of the same name (true) or append to it
            # (false).
            replace = optional(bool, false)
          })), [])

          # Header names to strip from the request before it reaches the backend.
          request_headers_to_remove = optional(list(string), [])

          # Headers to add to the response before it returns to the client.
          response_headers_to_add = optional(list(object({
            # The header name (e.g. "X-Client-Geo").
            header_name = string

            # The header value. May use load-balancer variables (e.g. "{client_region}").
            header_value = string

            # Replace an existing header of the same name (true) or append to it
            # (false).
            replace = optional(bool, false)
          })), [])

          # Header names to strip from the response before it returns to the client.
          response_headers_to_remove = optional(list(string), [])
        }))

        # Custom error pages for matched requests. Global external ALBs only.
        custom_error_response_policy = optional(object({
          # The backend bucket serving the error pages. Reference a GcpBackendBucket
          # or provide a self-link directly.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          error_service = optional(string, "")

          # Rules mapping response codes to the error page and (optionally) an
          # overridden status code.
          error_response_rules = optional(list(object({
            # Response codes this rule matches: exact ("404", "503") or a class ("4xx",
            # "5xx"). At least one is required.
            match_response_codes = list(string)

            # Override the response code returned to the client (e.g. serve a 200 with a
            # maintenance page). Empty keeps the original code.
            override_response_code = optional(number, 0)

            # Path within the error backend bucket to serve for matched codes (e.g.
            # "/errors/404.html").
            path = string
          })), [])
        }))
      })), [])
    })), [])

    # Routing self-tests evaluated by GCP at create/update time: each asserts
    # that a given host+path resolves to an expected service or redirect. A
    # failing test blocks the update — a guard against a routing change that
    # silently breaks a path. Mutable.
    tests = optional(list(object({
      # The request Host header the test sends.
      host = string

      # The request path the test sends.
      path = string

      # The backend service or backend bucket the request is expected to resolve
      # to. Reference a GcpBackendService or GcpBackendBucket, or provide a
      # self-link. Leave empty when asserting a redirect via
      # expected_redirect_response_code.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service = optional(string, "")

      # What this test guards — write it for whoever reads a failed-test error.
      description = optional(string, "")

      # The URL the request is expected to be redirected/rewritten to. Optional
      # when service is set.
      expected_output_url = optional(string, "")

      # The redirect status code the request is expected to produce. Cannot be set
      # together with service.
      expected_redirect_response_code = optional(number, 0)

      # Request headers the test sends.
      headers = optional(list(object({
        # Header name.
        name = string

        # Header value.
        value = string
      })), [])
    })), [])

    # What `terraform destroy` (or a stack teardown) may do to the URL map.
    # DELETE (the default when empty) allows deletion; PREVENT fails the
    # destroy outright — the URL map and anything orchestrating its teardown
    # stop there; ABANDON removes it from state without deleting it in GCP
    # (the map keeps serving, unmanaged). Client-side only — never sent to
    # the GCP API.
    deletion_policy = optional(string, "")
  })
}
