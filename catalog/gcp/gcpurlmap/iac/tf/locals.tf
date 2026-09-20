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

  # The scope selector. An empty region builds the global URL map; a region
  # name builds the regional one. Exactly one of the two resources in
  # main.tf exists (count guards), and outputs.tf picks whichever was
  # created.
  is_regional = var.spec.region != null && var.spec.region != ""

  # The cloud-side name defaults to metadata.name when the spec leaves
  # url_map_name empty — the same naming basis every kind uses.
  url_map_name = (
    var.spec.url_map_name != null && var.spec.url_map_name != ""
    ? var.spec.url_map_name
    : var.metadata.name
  )

  description = var.spec.description != "" ? var.spec.description : null

  default_service = var.spec.default_service != "" ? var.spec.default_service : null

  default_url_redirect = var.spec.default_url_redirect == null ? null : {
    host_redirect          = var.spec.default_url_redirect.host_redirect != "" ? var.spec.default_url_redirect.host_redirect : null
    https_redirect         = var.spec.default_url_redirect.https_redirect
    path_redirect          = var.spec.default_url_redirect.path_redirect != "" ? var.spec.default_url_redirect.path_redirect : null
    prefix_redirect        = var.spec.default_url_redirect.prefix_redirect != "" ? var.spec.default_url_redirect.prefix_redirect : null
    redirect_response_code = var.spec.default_url_redirect.redirect_response_code != "" ? var.spec.default_url_redirect.redirect_response_code : null
    strip_query            = var.spec.default_url_redirect.strip_query
  }

  # Route-action sub-policies (timeout, retry, mirror, CORS, fault
  # injection, stream duration, CDN cache) pass through as typed objects;
  # empty-vs-null conversion for their leaves happens at render time in
  # main.tf, keeping this file to the string/name normalization concern.
  default_route_action = var.spec.default_route_action == null ? null : {
    weighted_backend_services = [
      for w in var.spec.default_route_action.weighted_backend_services : {
        backend_service = w.backend_service
        weight          = w.weight
        header_action   = w.header_action
      }
    ]
    url_rewrite = var.spec.default_route_action.url_rewrite == null ? null : {
      host_rewrite        = try(var.spec.default_route_action.url_rewrite.host_rewrite, "") != "" ? var.spec.default_route_action.url_rewrite.host_rewrite : null
      path_prefix_rewrite = try(var.spec.default_route_action.url_rewrite.path_prefix_rewrite, "") != "" ? var.spec.default_route_action.url_rewrite.path_prefix_rewrite : null
    }
    timeout                = var.spec.default_route_action.timeout
    retry_policy           = var.spec.default_route_action.retry_policy
    request_mirror_policy  = var.spec.default_route_action.request_mirror_policy
    cors_policy            = var.spec.default_route_action.cors_policy
    fault_injection_policy = var.spec.default_route_action.fault_injection_policy
    max_stream_duration    = var.spec.default_route_action.max_stream_duration
    cache_policy           = var.spec.default_route_action.cache_policy
  }

  default_custom_error_response_policy = var.spec.default_custom_error_response_policy == null ? null : {
    error_service = try(var.spec.default_custom_error_response_policy.error_service, "") != "" ? var.spec.default_custom_error_response_policy.error_service : null
    error_response_rules = [
      for rule in var.spec.default_custom_error_response_policy.error_response_rules : {
        match_response_codes   = rule.match_response_codes
        override_response_code = try(rule.override_response_code, null) != 0 ? rule.override_response_code : null
        path                   = rule.path
      }
    ]
  }

  header_action = var.spec.header_action == null ? null : {
    request_headers_to_add = [
      for h in var.spec.header_action.request_headers_to_add : {
        header_name  = h.header_name
        header_value = h.header_value
        replace      = h.replace
      }
    ]
    request_headers_to_remove = length(var.spec.header_action.request_headers_to_remove) > 0 ? var.spec.header_action.request_headers_to_remove : null
    response_headers_to_add = [
      for h in var.spec.header_action.response_headers_to_add : {
        header_name  = h.header_name
        header_value = h.header_value
        replace      = h.replace
      }
    ]
    response_headers_to_remove = length(var.spec.header_action.response_headers_to_remove) > 0 ? var.spec.header_action.response_headers_to_remove : null
  }

  host_rules = [
    for rule in var.spec.host_rules : {
      hosts        = rule.hosts
      path_matcher = rule.path_matcher
      description  = rule.description != "" ? rule.description : null
    }
  ]

  path_matchers = [
    for matcher in var.spec.path_matchers : {
      name            = matcher.name
      description     = matcher.description != "" ? matcher.description : null
      default_service = try(matcher.default_service, "") != "" ? matcher.default_service : null
      default_url_redirect = matcher.default_url_redirect == null ? null : {
        host_redirect          = matcher.default_url_redirect.host_redirect != "" ? matcher.default_url_redirect.host_redirect : null
        https_redirect         = matcher.default_url_redirect.https_redirect
        path_redirect          = matcher.default_url_redirect.path_redirect != "" ? matcher.default_url_redirect.path_redirect : null
        prefix_redirect        = matcher.default_url_redirect.prefix_redirect != "" ? matcher.default_url_redirect.prefix_redirect : null
        redirect_response_code = matcher.default_url_redirect.redirect_response_code != "" ? matcher.default_url_redirect.redirect_response_code : null
        strip_query            = matcher.default_url_redirect.strip_query
      }
      default_route_action = matcher.default_route_action == null ? null : {
        weighted_backend_services = [
          for w in matcher.default_route_action.weighted_backend_services : {
            backend_service = w.backend_service
            weight          = w.weight
            header_action   = w.header_action
          }
        ]
        # path_template_rewrite is honored here by the regional map alone (the
        # spec CEL rejects it on a global map); the global resource block
        # never reads the key.
        url_rewrite = matcher.default_route_action.url_rewrite == null ? null : {
          host_rewrite          = try(matcher.default_route_action.url_rewrite.host_rewrite, "") != "" ? matcher.default_route_action.url_rewrite.host_rewrite : null
          path_prefix_rewrite   = try(matcher.default_route_action.url_rewrite.path_prefix_rewrite, "") != "" ? matcher.default_route_action.url_rewrite.path_prefix_rewrite : null
          path_template_rewrite = try(matcher.default_route_action.url_rewrite.path_template_rewrite, "") != "" ? matcher.default_route_action.url_rewrite.path_template_rewrite : null
        }
        timeout                = matcher.default_route_action.timeout
        retry_policy           = matcher.default_route_action.retry_policy
        request_mirror_policy  = matcher.default_route_action.request_mirror_policy
        cors_policy            = matcher.default_route_action.cors_policy
        fault_injection_policy = matcher.default_route_action.fault_injection_policy
        max_stream_duration    = matcher.default_route_action.max_stream_duration
        cache_policy           = matcher.default_route_action.cache_policy
      }
      default_custom_error_response_policy = matcher.default_custom_error_response_policy == null ? null : {
        error_service = try(matcher.default_custom_error_response_policy.error_service, "") != "" ? matcher.default_custom_error_response_policy.error_service : null
        error_response_rules = [
          for rule in matcher.default_custom_error_response_policy.error_response_rules : {
            match_response_codes   = rule.match_response_codes
            override_response_code = try(rule.override_response_code, null) != 0 ? rule.override_response_code : null
            path                   = rule.path
          }
        ]
      }
      header_action = matcher.header_action == null ? null : {
        request_headers_to_add = [
          for h in matcher.header_action.request_headers_to_add : {
            header_name  = h.header_name
            header_value = h.header_value
            replace      = h.replace
          }
        ]
        request_headers_to_remove = length(matcher.header_action.request_headers_to_remove) > 0 ? matcher.header_action.request_headers_to_remove : null
        response_headers_to_add = [
          for h in matcher.header_action.response_headers_to_add : {
            header_name  = h.header_name
            header_value = h.header_value
            replace      = h.replace
          }
        ]
        response_headers_to_remove = length(matcher.header_action.response_headers_to_remove) > 0 ? matcher.header_action.response_headers_to_remove : null
      }
      path_rules = [
        for rule in matcher.path_rules : {
          paths   = rule.paths
          service = try(rule.service, "") != "" ? rule.service : null
          url_redirect = rule.url_redirect == null ? null : {
            host_redirect          = rule.url_redirect.host_redirect != "" ? rule.url_redirect.host_redirect : null
            https_redirect         = rule.url_redirect.https_redirect
            path_redirect          = rule.url_redirect.path_redirect != "" ? rule.url_redirect.path_redirect : null
            prefix_redirect        = rule.url_redirect.prefix_redirect != "" ? rule.url_redirect.prefix_redirect : null
            redirect_response_code = rule.url_redirect.redirect_response_code != "" ? rule.url_redirect.redirect_response_code : null
            strip_query            = rule.url_redirect.strip_query
          }
          route_action = rule.route_action == null ? null : {
            weighted_backend_services = [
              for w in rule.route_action.weighted_backend_services : {
                backend_service = w.backend_service
                weight          = w.weight
                header_action   = w.header_action
              }
            ]
            url_rewrite = rule.route_action.url_rewrite == null ? null : {
              host_rewrite        = try(rule.route_action.url_rewrite.host_rewrite, "") != "" ? rule.route_action.url_rewrite.host_rewrite : null
              path_prefix_rewrite = try(rule.route_action.url_rewrite.path_prefix_rewrite, "") != "" ? rule.route_action.url_rewrite.path_prefix_rewrite : null
            }
            timeout                = rule.route_action.timeout
            retry_policy           = rule.route_action.retry_policy
            request_mirror_policy  = rule.route_action.request_mirror_policy
            cors_policy            = rule.route_action.cors_policy
            fault_injection_policy = rule.route_action.fault_injection_policy
            max_stream_duration    = rule.route_action.max_stream_duration
            cache_policy           = rule.route_action.cache_policy
          }
          custom_error_response_policy = rule.custom_error_response_policy == null ? null : {
            error_service = try(rule.custom_error_response_policy.error_service, "") != "" ? rule.custom_error_response_policy.error_service : null
            error_response_rules = [
              for err in rule.custom_error_response_policy.error_response_rules : {
                match_response_codes   = err.match_response_codes
                override_response_code = try(err.override_response_code, null) != 0 ? err.override_response_code : null
                path                   = err.path
              }
            ]
          }
        }
      ]
      route_rules = [
        for rule in matcher.route_rules : {
          priority = rule.priority
          service  = try(rule.service, "") != "" ? rule.service : null
          url_redirect = rule.url_redirect == null ? null : {
            host_redirect          = rule.url_redirect.host_redirect != "" ? rule.url_redirect.host_redirect : null
            https_redirect         = rule.url_redirect.https_redirect
            path_redirect          = rule.url_redirect.path_redirect != "" ? rule.url_redirect.path_redirect : null
            prefix_redirect        = rule.url_redirect.prefix_redirect != "" ? rule.url_redirect.prefix_redirect : null
            redirect_response_code = rule.url_redirect.redirect_response_code != "" ? rule.url_redirect.redirect_response_code : null
            strip_query            = rule.url_redirect.strip_query
          }
          route_action = rule.route_action == null ? null : {
            weighted_backend_services = [
              for w in rule.route_action.weighted_backend_services : {
                backend_service = w.backend_service
                weight          = w.weight
                header_action   = w.header_action
              }
            ]
            url_rewrite = rule.route_action.url_rewrite == null ? null : {
              host_rewrite          = try(rule.route_action.url_rewrite.host_rewrite, "") != "" ? rule.route_action.url_rewrite.host_rewrite : null
              path_prefix_rewrite   = try(rule.route_action.url_rewrite.path_prefix_rewrite, "") != "" ? rule.route_action.url_rewrite.path_prefix_rewrite : null
              path_template_rewrite = try(rule.route_action.url_rewrite.path_template_rewrite, "") != "" ? rule.route_action.url_rewrite.path_template_rewrite : null
            }
            timeout                = rule.route_action.timeout
            retry_policy           = rule.route_action.retry_policy
            request_mirror_policy  = rule.route_action.request_mirror_policy
            cors_policy            = rule.route_action.cors_policy
            fault_injection_policy = rule.route_action.fault_injection_policy
            max_stream_duration    = rule.route_action.max_stream_duration
            cache_policy           = rule.route_action.cache_policy
          }
          header_action = rule.header_action == null ? null : {
            request_headers_to_add = [
              for h in rule.header_action.request_headers_to_add : {
                header_name  = h.header_name
                header_value = h.header_value
                replace      = h.replace
              }
            ]
            request_headers_to_remove = length(rule.header_action.request_headers_to_remove) > 0 ? rule.header_action.request_headers_to_remove : null
            response_headers_to_add = [
              for h in rule.header_action.response_headers_to_add : {
                header_name  = h.header_name
                header_value = h.header_value
                replace      = h.replace
              }
            ]
            response_headers_to_remove = length(rule.header_action.response_headers_to_remove) > 0 ? rule.header_action.response_headers_to_remove : null
          }
          custom_error_response_policy = rule.custom_error_response_policy == null ? null : {
            error_service = try(rule.custom_error_response_policy.error_service, "") != "" ? rule.custom_error_response_policy.error_service : null
            error_response_rules = [
              for err in rule.custom_error_response_policy.error_response_rules : {
                match_response_codes   = err.match_response_codes
                override_response_code = try(err.override_response_code, null) != 0 ? err.override_response_code : null
                path                   = err.path
              }
            ]
          }
          match_rules = [
            for match in rule.match_rules : {
              prefix_match        = match.prefix_match != "" ? match.prefix_match : null
              full_path_match     = match.full_path_match != "" ? match.full_path_match : null
              regex_match         = match.regex_match != "" ? match.regex_match : null
              path_template_match = match.path_template_match != "" ? match.path_template_match : null
              ignore_case         = match.ignore_case
              header_matches = [
                for hm in match.header_matches : {
                  header_name   = hm.header_name
                  exact_match   = try(hm.exact_match, "") != "" ? hm.exact_match : null
                  prefix_match  = try(hm.prefix_match, "") != "" ? hm.prefix_match : null
                  suffix_match  = try(hm.suffix_match, "") != "" ? hm.suffix_match : null
                  regex_match   = try(hm.regex_match, "") != "" ? hm.regex_match : null
                  present_match = hm.present_match
                  invert_match  = hm.invert_match
                  range_match = hm.range_match == null ? null : {
                    range_start = hm.range_match.range_start
                    range_end   = hm.range_match.range_end
                  }
                }
              ]
              query_parameter_matches = [
                for qm in match.query_parameter_matches : {
                  name          = qm.name
                  exact_match   = try(qm.exact_match, "") != "" ? qm.exact_match : null
                  present_match = qm.present_match
                  regex_match   = try(qm.regex_match, "") != "" ? qm.regex_match : null
                }
              ]
              metadata_filters = [
                for mf in match.metadata_filters : {
                  filter_match_criteria = mf.filter_match_criteria
                  filter_labels = [
                    for fl in mf.filter_labels : {
                      name  = fl.name
                      value = fl.value
                    }
                  ]
                }
              ]
            }
          ]
        }
      ]
    }
  ]

  tests = [
    for test in var.spec.tests : {
      host                            = test.host
      path                            = test.path
      service                         = try(test.service, "") != "" ? test.service : null
      description                     = test.description != "" ? test.description : null
      expected_output_url             = test.expected_output_url != "" ? test.expected_output_url : null
      expected_redirect_response_code = try(test.expected_redirect_response_code, null) != 0 ? test.expected_redirect_response_code : null
      headers = [
        for h in test.headers : {
          name  = h.name
          value = h.value
        }
      ]
    }
  ]
}
