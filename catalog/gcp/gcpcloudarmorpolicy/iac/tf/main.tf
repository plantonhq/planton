# Enable the Compute Engine API so a fresh project can host security
# policies. disable_on_destroy is false: tearing down one policy must never
# disable the API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# GCP models the global and regional security policies as two API
# collections. spec.region selects which resource below is created; exactly
# one exists (count guards) and outputs.tf picks whichever was built.
#
# The GLOBAL policy: the WAF in front of a global external Application Load
# Balancer's backend services and backend buckets. The default-rule
# contract: creating with NO rules lets the API add a default "allow all"
# rule at priority 2147483647 automatically; providing ANY rules requires
# the set to include that default explicitly (the spec enforces it
# pre-deploy, mirroring the API's own rejection).
resource "google_compute_security_policy" "this" {
  count = local.is_regional ? 0 : 1

  name        = local.policy_name
  project     = local.project_id
  description = var.spec.description != "" ? var.spec.description : null
  type        = var.spec.type != "" ? var.spec.type : null

  labels = local.final_labels

  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]

  # Adaptive Protection (CAAP): Layer 7 DDoS detection with optional
  # per-granularity detection/auto-deploy threshold overrides.
  dynamic "adaptive_protection_config" {
    for_each = var.spec.adaptive_protection_config != null ? [var.spec.adaptive_protection_config] : []
    content {
      layer_7_ddos_defense_config {
        enable          = adaptive_protection_config.value.enable_layer_7_ddos_defense
        rule_visibility = adaptive_protection_config.value.rule_visibility != "" ? adaptive_protection_config.value.rule_visibility : null

        dynamic "threshold_configs" {
          for_each = adaptive_protection_config.value.threshold_configs
          content {
            name                                    = threshold_configs.value.name
            auto_deploy_confidence_threshold        = threshold_configs.value.auto_deploy_confidence_threshold
            auto_deploy_impacted_baseline_threshold = threshold_configs.value.auto_deploy_impacted_baseline_threshold
            auto_deploy_load_threshold              = threshold_configs.value.auto_deploy_load_threshold
            auto_deploy_expiration_sec              = threshold_configs.value.auto_deploy_expiration_sec
            detection_absolute_qps                  = threshold_configs.value.detection_absolute_qps
            detection_load_threshold                = threshold_configs.value.detection_load_threshold
            detection_relative_to_baseline_qps      = threshold_configs.value.detection_relative_to_baseline_qps

            dynamic "traffic_granularity_configs" {
              for_each = threshold_configs.value.traffic_granularity_configs
              content {
                type                     = traffic_granularity_configs.value.type
                value                    = traffic_granularity_configs.value.value != "" ? traffic_granularity_configs.value.value : null
                enable_each_unique_value = traffic_granularity_configs.value.enable_each_unique_value ? true : null
              }
            }
          }
        }
      }
    }
  }

  # Advanced options: JSON body parsing (with custom content types),
  # logging verbosity, and true-client-IP resolution headers.
  dynamic "advanced_options_config" {
    for_each = var.spec.advanced_options_config != null ? [var.spec.advanced_options_config] : []
    content {
      json_parsing            = advanced_options_config.value.json_parsing != "" ? advanced_options_config.value.json_parsing : null
      log_level               = advanced_options_config.value.log_level != "" ? advanced_options_config.value.log_level : null
      user_ip_request_headers = length(advanced_options_config.value.user_ip_request_headers) > 0 ? advanced_options_config.value.user_ip_request_headers : null

      # How much of each request body the WAF inspects (8KB default).
      request_body_inspection_size = advanced_options_config.value.request_body_inspection_size != "" ? advanced_options_config.value.request_body_inspection_size : null

      dynamic "json_custom_config" {
        for_each = advanced_options_config.value.json_custom_config != null ? [advanced_options_config.value.json_custom_config] : []
        content {
          content_types = json_custom_config.value.content_types
        }
      }
    }
  }

  # Policy-level reCAPTCHA site key for GOOGLE_RECAPTCHA redirect rules.
  dynamic "recaptcha_options_config" {
    for_each = var.spec.recaptcha_options_config != null ? [var.spec.recaptcha_options_config] : []
    content {
      redirect_site_key = recaptcha_options_config.value.redirect_site_key
    }
  }

  # Security rules, evaluated in priority order (lowest number first).
  dynamic "rule" {
    for_each = var.spec.rules
    content {
      action      = rule.value.action
      priority    = rule.value.priority
      description = rule.value.description != "" ? rule.value.description : null
      preview     = rule.value.preview ? true : null

      # Match condition: versioned_expr + config OR a CEL expr (with
      # optional reCAPTCHA site-key options).
      match {
        versioned_expr = rule.value.match.versioned_expr != "" ? rule.value.match.versioned_expr : null

        dynamic "config" {
          for_each = rule.value.match.versioned_expr != "" ? [1] : []
          content {
            src_ip_ranges = rule.value.match.src_ip_ranges
          }
        }

        dynamic "expr" {
          for_each = rule.value.match.expression != "" ? [1] : []
          content {
            expression = rule.value.match.expression
          }
        }

        dynamic "expr_options" {
          for_each = rule.value.match.expr_options != null ? [rule.value.match.expr_options] : []
          content {
            recaptcha_options {
              action_token_site_keys  = length(expr_options.value.action_token_site_keys) > 0 ? expr_options.value.action_token_site_keys : null
              session_token_site_keys = length(expr_options.value.session_token_site_keys) > 0 ? expr_options.value.session_token_site_keys : null
            }
          }
        }
      }

      # Rate limit options (for throttle and rate_based_ban actions).
      dynamic "rate_limit_options" {
        for_each = rule.value.rate_limit_options != null ? [rule.value.rate_limit_options] : []
        content {
          conform_action      = rate_limit_options.value.conform_action
          exceed_action       = rate_limit_options.value.exceed_action
          enforce_on_key      = rate_limit_options.value.enforce_on_key != "" ? rate_limit_options.value.enforce_on_key : null
          enforce_on_key_name = rate_limit_options.value.enforce_on_key_name != "" ? rate_limit_options.value.enforce_on_key_name : null
          ban_duration_sec    = rate_limit_options.value.ban_duration_sec > 0 ? rate_limit_options.value.ban_duration_sec : null

          # Composite rate-limit key: component values concatenate to form
          # the key requests are counted against.
          dynamic "enforce_on_key_configs" {
            for_each = rate_limit_options.value.enforce_on_key_configs
            content {
              enforce_on_key_type = enforce_on_key_configs.value.enforce_on_key_type
              enforce_on_key_name = enforce_on_key_configs.value.enforce_on_key_name != "" ? enforce_on_key_configs.value.enforce_on_key_name : null
            }
          }

          rate_limit_threshold {
            count        = rate_limit_options.value.rate_limit_threshold.count
            interval_sec = rate_limit_options.value.rate_limit_threshold.interval_sec
          }

          dynamic "ban_threshold" {
            for_each = rate_limit_options.value.ban_threshold != null ? [rate_limit_options.value.ban_threshold] : []
            content {
              count        = ban_threshold.value.count
              interval_sec = ban_threshold.value.interval_sec
            }
          }

          dynamic "exceed_redirect_options" {
            for_each = rate_limit_options.value.exceed_redirect_options != null ? [rate_limit_options.value.exceed_redirect_options] : []
            content {
              type   = exceed_redirect_options.value.type
              target = exceed_redirect_options.value.target != "" ? exceed_redirect_options.value.target : null
            }
          }
        }
      }

      # Redirect options (for redirect actions).
      dynamic "redirect_options" {
        for_each = rule.value.redirect_options != null ? [rule.value.redirect_options] : []
        content {
          type   = redirect_options.value.type
          target = redirect_options.value.target != "" ? redirect_options.value.target : null
        }
      }

      # Header action (add custom headers to matching requests).
      dynamic "header_action" {
        for_each = rule.value.header_action != null ? [rule.value.header_action] : []
        content {
          dynamic "request_headers_to_adds" {
            for_each = header_action.value.request_headers_to_adds
            content {
              header_name  = request_headers_to_adds.value.header_name
              header_value = request_headers_to_adds.value.header_value != "" ? request_headers_to_adds.value.header_value : null
            }
          }
        }
      }

      # Preconfigured WAF rule exclusions.
      dynamic "preconfigured_waf_config" {
        for_each = rule.value.preconfigured_waf_config != null ? [rule.value.preconfigured_waf_config] : []
        content {
          dynamic "exclusion" {
            for_each = preconfigured_waf_config.value.exclusions
            content {
              target_rule_set = exclusion.value.target_rule_set
              target_rule_ids = length(exclusion.value.target_rule_ids) > 0 ? exclusion.value.target_rule_ids : null

              dynamic "request_header" {
                for_each = exclusion.value.request_headers
                content {
                  operator = request_header.value.operator
                  value    = request_header.value.value != "" ? request_header.value.value : null
                }
              }

              dynamic "request_cookie" {
                for_each = exclusion.value.request_cookies
                content {
                  operator = request_cookie.value.operator
                  value    = request_cookie.value.value != "" ? request_cookie.value.value : null
                }
              }

              dynamic "request_uri" {
                for_each = exclusion.value.request_uris
                content {
                  operator = request_uri.value.operator
                  value    = request_uri.value.value != "" ? request_uri.value.value : null
                }
              }

              dynamic "request_query_param" {
                for_each = exclusion.value.request_query_params
                content {
                  operator = request_query_param.value.operator
                  value    = request_query_param.value.value != "" ? request_query_param.value.value : null
                }
              }
            }
          }
        }
      }
    }
  }
}

# The REGIONAL policy: the WAF a regional backend service attaches (regional
# external and internal Application Load Balancers) or, as a
# CLOUD_ARMOR_NETWORK policy, the packet filter in front of the region's
# passthrough Network Load Balancers, protocol forwarding, and public-IP
# VMs. The regional collection carries no labels, Adaptive Protection,
# reCAPTCHA options, request-body inspection size, redirect action, or
# header injection -- the spec rejects each when region is set, so none is
# wired here. Google compares the rule list as a set, so rule order in the
# manifest never shows as a diff.
resource "google_compute_region_security_policy" "this" {
  count = local.is_regional ? 1 : 0

  name        = local.policy_name
  project     = local.project_id
  region      = var.spec.region
  description = var.spec.description != "" ? var.spec.description : null
  type        = var.spec.type != "" ? var.spec.type : null

  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]

  # Advanced options shared with the global arm: JSON body parsing (with
  # custom content types), logging verbosity, and true-client-IP headers.
  dynamic "advanced_options_config" {
    for_each = var.spec.advanced_options_config != null ? [var.spec.advanced_options_config] : []
    content {
      json_parsing            = advanced_options_config.value.json_parsing != "" ? advanced_options_config.value.json_parsing : null
      log_level               = advanced_options_config.value.log_level != "" ? advanced_options_config.value.log_level : null
      user_ip_request_headers = length(advanced_options_config.value.user_ip_request_headers) > 0 ? advanced_options_config.value.user_ip_request_headers : null

      dynamic "json_custom_config" {
        for_each = advanced_options_config.value.json_custom_config != null ? [advanced_options_config.value.json_custom_config] : []
        content {
          content_types = json_custom_config.value.content_types
        }
      }
    }
  }

  # Network DDoS protection level (CLOUD_ARMOR_NETWORK policies).
  dynamic "ddos_protection_config" {
    for_each = var.spec.ddos_protection_config != null ? [var.spec.ddos_protection_config] : []
    content {
      ddos_protection = ddos_protection_config.value.ddos_protection
    }
  }

  # Custom packet fields the network rules match on. offset and size are
  # tri-state in the spec: unset sends null so Google applies its default,
  # while an explicit 0 offset is a real byte position and is sent as 0.
  dynamic "user_defined_fields" {
    for_each = var.spec.user_defined_fields
    content {
      name   = user_defined_fields.value.name != "" ? user_defined_fields.value.name : null
      base   = user_defined_fields.value.base
      offset = user_defined_fields.value.offset
      size   = user_defined_fields.value.size
      mask   = user_defined_fields.value.mask != "" ? user_defined_fields.value.mask : null
    }
  }

  # Security rules, evaluated in priority order (lowest number first). Each
  # rule matches through exactly one arm: match (HTTP attributes) or
  # network_match (packet headers, CLOUD_ARMOR_NETWORK policies).
  dynamic "rules" {
    for_each = var.spec.rules
    content {
      action      = rules.value.action
      priority    = rules.value.priority
      description = rules.value.description != "" ? rules.value.description : null
      preview     = rules.value.preview ? true : null

      dynamic "match" {
        for_each = rules.value.match != null ? [rules.value.match] : []
        content {
          versioned_expr = match.value.versioned_expr != "" ? match.value.versioned_expr : null

          dynamic "config" {
            for_each = match.value.versioned_expr != "" ? [1] : []
            content {
              src_ip_ranges = match.value.src_ip_ranges
            }
          }

          dynamic "expr" {
            for_each = match.value.expression != "" ? [1] : []
            content {
              expression = match.value.expression
            }
          }
        }
      }

      # Packet-level match: every listed field must match; an empty list
      # matches any value and is sent as null so Google treats the field as
      # unconstrained.
      dynamic "network_match" {
        for_each = rules.value.network_match != null ? [rules.value.network_match] : []
        content {
          src_ip_ranges    = length(network_match.value.src_ip_ranges) > 0 ? network_match.value.src_ip_ranges : null
          dest_ip_ranges   = length(network_match.value.dest_ip_ranges) > 0 ? network_match.value.dest_ip_ranges : null
          ip_protocols     = length(network_match.value.ip_protocols) > 0 ? network_match.value.ip_protocols : null
          src_ports        = length(network_match.value.src_ports) > 0 ? network_match.value.src_ports : null
          dest_ports       = length(network_match.value.dest_ports) > 0 ? network_match.value.dest_ports : null
          src_region_codes = length(network_match.value.src_region_codes) > 0 ? network_match.value.src_region_codes : null
          src_asns         = length(network_match.value.src_asns) > 0 ? network_match.value.src_asns : null

          dynamic "user_defined_fields" {
            for_each = network_match.value.user_defined_fields
            content {
              name   = user_defined_fields.value.name
              values = user_defined_fields.value.values
            }
          }
        }
      }

      # Rate limit options (throttle and rate_based_ban). The regional
      # collection exceeds to deny(STATUS) only -- no redirect arm.
      dynamic "rate_limit_options" {
        for_each = rules.value.rate_limit_options != null ? [rules.value.rate_limit_options] : []
        content {
          conform_action      = rate_limit_options.value.conform_action
          exceed_action       = rate_limit_options.value.exceed_action
          enforce_on_key      = rate_limit_options.value.enforce_on_key != "" ? rate_limit_options.value.enforce_on_key : null
          enforce_on_key_name = rate_limit_options.value.enforce_on_key_name != "" ? rate_limit_options.value.enforce_on_key_name : null
          ban_duration_sec    = rate_limit_options.value.ban_duration_sec > 0 ? rate_limit_options.value.ban_duration_sec : null

          dynamic "enforce_on_key_configs" {
            for_each = rate_limit_options.value.enforce_on_key_configs
            content {
              enforce_on_key_type = enforce_on_key_configs.value.enforce_on_key_type
              enforce_on_key_name = enforce_on_key_configs.value.enforce_on_key_name != "" ? enforce_on_key_configs.value.enforce_on_key_name : null
            }
          }

          rate_limit_threshold {
            count        = rate_limit_options.value.rate_limit_threshold.count
            interval_sec = rate_limit_options.value.rate_limit_threshold.interval_sec
          }

          dynamic "ban_threshold" {
            for_each = rate_limit_options.value.ban_threshold != null ? [rate_limit_options.value.ban_threshold] : []
            content {
              count        = ban_threshold.value.count
              interval_sec = ban_threshold.value.interval_sec
            }
          }
        }
      }

      # Preconfigured WAF rule exclusions.
      dynamic "preconfigured_waf_config" {
        for_each = rules.value.preconfigured_waf_config != null ? [rules.value.preconfigured_waf_config] : []
        content {
          dynamic "exclusion" {
            for_each = preconfigured_waf_config.value.exclusions
            content {
              target_rule_set = exclusion.value.target_rule_set
              target_rule_ids = length(exclusion.value.target_rule_ids) > 0 ? exclusion.value.target_rule_ids : null

              dynamic "request_header" {
                for_each = exclusion.value.request_headers
                content {
                  operator = request_header.value.operator
                  value    = request_header.value.value != "" ? request_header.value.value : null
                }
              }

              dynamic "request_cookie" {
                for_each = exclusion.value.request_cookies
                content {
                  operator = request_cookie.value.operator
                  value    = request_cookie.value.value != "" ? request_cookie.value.value : null
                }
              }

              dynamic "request_uri" {
                for_each = exclusion.value.request_uris
                content {
                  operator = request_uri.value.operator
                  value    = request_uri.value.value != "" ? request_uri.value.value : null
                }
              }

              dynamic "request_query_param" {
                for_each = exclusion.value.request_query_params
                content {
                  operator = request_query_param.value.operator
                  value    = request_query_param.value.value != "" ? request_query_param.value.value : null
                }
              }
            }
          }
        }
      }
    }
  }
}

# The region's enrollment in advanced network DDoS protection: Google's
# per-region, per-project network edge security service with this policy
# attached. Created only when the spec declares the block; shares the
# policy's deletion_policy so the enrollment and the policy live and die
# together.
resource "google_compute_network_edge_security_service" "this" {
  count = local.create_edge_service ? 1 : 0

  name            = local.edge_service_name
  project         = local.project_id
  region          = var.spec.region
  description     = var.spec.network_edge_security_service.description != "" ? var.spec.network_edge_security_service.description : null
  security_policy = google_compute_region_security_policy.this[0].self_link

  deletion_policy = local.deletion_policy
}
