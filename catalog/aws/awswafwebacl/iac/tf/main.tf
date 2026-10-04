resource "aws_wafv2_web_acl" "this" {
  # The ACL's AWS name is the Planton resource name -- the stable identity
  # operators see. Name and scope are create-time immutable (ForceNew).
  name  = var.metadata.name
  scope = var.spec.scope

  description = var.spec.description != "" ? var.spec.description : null

  # Baseline posture when no rule matches: exactly one of allow/block, with
  # the action's own customization (custom block response, custom allow
  # request headers) -- the spec's CEL couples each to its action type.
  default_action {
    dynamic "allow" {
      for_each = var.spec.default_action.type == "allow" ? [var.spec.default_action] : []
      content {
        dynamic "custom_request_handling" {
          for_each = length(allow.value.custom_request_headers) > 0 ? [1] : []
          content {
            dynamic "insert_header" {
              for_each = allow.value.custom_request_headers
              content {
                name  = insert_header.value.name
                value = insert_header.value.value
              }
            }
          }
        }
      }
    }

    dynamic "block" {
      for_each = var.spec.default_action.type == "block" ? [var.spec.default_action] : []
      content {
        dynamic "custom_response" {
          for_each = block.value.custom_response != null ? [block.value.custom_response] : []
          content {
            response_code            = custom_response.value.response_code
            custom_response_body_key = custom_response.value.custom_response_body_key != "" ? custom_response.value.custom_response_body_key : null

            dynamic "response_header" {
              for_each = custom_response.value.response_headers
              content {
                name  = response_header.value.name
                value = response_header.value.value
              }
            }
          }
        }
      }
    }
  }

  # ACL-level visibility defaults: metrics on, sampling on, metric name =
  # resource name (identical defaults in the Pulumi module).
  visibility_config {
    cloudwatch_metrics_enabled = var.spec.visibility_config != null ? var.spec.visibility_config.cloudwatch_metrics_enabled : true
    sampled_requests_enabled   = var.spec.visibility_config != null ? var.spec.visibility_config.sampled_requests_enabled : true
    metric_name                = var.spec.visibility_config != null && try(var.spec.visibility_config.metric_name, "") != "" ? var.spec.visibility_config.metric_name : var.metadata.name
  }

  # Reusable branded error bodies, referenced from block actions by key.
  dynamic "custom_response_body" {
    for_each = var.spec.custom_response_bodies
    content {
      key          = custom_response_body.value.key
      content      = custom_response_body.value.content
      content_type = custom_response_body.value.content_type
    }
  }

  token_domains = length(var.spec.token_domains) > 0 ? var.spec.token_domains : null

  # Web-ACL-wide immunity windows for CAPTCHA solves and silent-challenge
  # responses (rules can override per rule inside the rule JSON).
  dynamic "captcha_config" {
    for_each = var.spec.captcha_config != null ? [var.spec.captcha_config] : []
    content {
      immunity_time_property {
        immunity_time = captcha_config.value.immunity_time_sec
      }
    }
  }

  dynamic "challenge_config" {
    for_each = var.spec.challenge_config != null ? [var.spec.challenge_config] : []
    content {
      immunity_time_property {
        immunity_time = challenge_config.value.immunity_time_sec
      }
    }
  }

  # Per-resource-type request-body inspection limits (default 16 KB).
  dynamic "association_config" {
    for_each = var.spec.association_config != null ? [var.spec.association_config] : []
    content {
      request_body {
        dynamic "cloudfront" {
          for_each = association_config.value.cloudfront_request_body_limit != "" ? [1] : []
          content {
            default_size_inspection_limit = association_config.value.cloudfront_request_body_limit
          }
        }
        dynamic "api_gateway" {
          for_each = association_config.value.api_gateway_request_body_limit != "" ? [1] : []
          content {
            default_size_inspection_limit = association_config.value.api_gateway_request_body_limit
          }
        }
        dynamic "cognito_user_pool" {
          for_each = association_config.value.cognito_user_pool_request_body_limit != "" ? [1] : []
          content {
            default_size_inspection_limit = association_config.value.cognito_user_pool_request_body_limit
          }
        }
        dynamic "app_runner_service" {
          for_each = association_config.value.app_runner_service_request_body_limit != "" ? [1] : []
          content {
            default_size_inspection_limit = association_config.value.app_runner_service_request_body_limit
          }
        }
        dynamic "verified_access_instance" {
          for_each = association_config.value.verified_access_instance_request_body_limit != "" ? [1] : []
          content {
            default_size_inspection_limit = association_config.value.verified_access_instance_request_body_limit
          }
        }
      }
    }
  }

  # Field-level masking in ALL WAF outputs (logs, sampled requests, rule
  # match details) -- stronger than the logging block's redaction, which
  # only affects the log destination.
  dynamic "data_protection_config" {
    for_each = var.spec.data_protection_config != null ? [var.spec.data_protection_config] : []
    content {
      dynamic "data_protection" {
        for_each = data_protection_config.value.data_protections
        content {
          action = data_protection.value.action
          field {
            field_type = data_protection.value.field_type
            field_keys = length(data_protection.value.field_keys) > 0 ? data_protection.value.field_keys : null
          }
          exclude_rule_match_details = data_protection.value.exclude_rule_match_details
          exclude_rate_based_details = data_protection.value.exclude_rate_based_details
        }
      }
    }
  }

  # Rules as AWS API JSON -- the single rule surface both engines share; see
  # locals.tf for the statement-tree serialization and the depth model.
  rule_json = length(var.spec.rules) > 0 ? jsonencode(local.rules_waf) : null

  tags = local.aws_tags

  lifecycle {
    # The statement serialization unrolls three nesting levels below a
    # rule's root (the same depth the provider's own structured schema
    # supports). A deeper tree must fail the plan loudly -- silently
    # dropping a nested statement would deploy a WEAKER firewall than the
    # manifest declares.
    precondition {
      condition     = length(local.depth_overflow_nodes) == 0
      error_message = "A rule statement is nested more than three levels deep (via and/or/not or scope-down statements). Restructure the rule, or express the deep subtree with the custom_statement escape hatch (raw AWS WAF JSON), which supports any depth."
    }
  }
}

# WAF request logging is a separate PUT-style AWS resource keyed by the web
# ACL's ARN (at most one per ACL), so it shares this module rather than being
# its own kind. The destination's name must start with "aws-waf-logs-"
# (AWS-enforced).
resource "aws_wafv2_web_acl_logging_configuration" "this" {
  count = var.spec.logging != null ? 1 : 0

  resource_arn = aws_wafv2_web_acl.this.arn
  # The generator flattens StringValueOrRef to its resolved string (the
  # orchestrator resolves any value_from before the module runs).
  log_destination_configs = [var.spec.logging.destination_arn]

  dynamic "redacted_fields" {
    for_each = var.spec.logging.redacted_header_names
    content {
      single_header {
        name = redacted_fields.value
      }
    }
  }

  dynamic "redacted_fields" {
    for_each = var.spec.logging.redact_uri_path ? [1] : []
    content {
      uri_path {}
    }
  }

  dynamic "redacted_fields" {
    for_each = var.spec.logging.redact_query_string ? [1] : []
    content {
      query_string {}
    }
  }

  dynamic "redacted_fields" {
    for_each = var.spec.logging.redact_method ? [1] : []
    content {
      method {}
    }
  }

  # Log filtering: keep or drop records by the action WAF applied or by
  # labels on the request. Each spec condition carries exactly one of
  # action / label_name (spec CEL), so each renders as a one-armed
  # condition block.
  dynamic "logging_filter" {
    for_each = var.spec.logging.filter != null ? [var.spec.logging.filter] : []
    content {
      default_behavior = logging_filter.value.default_behavior

      dynamic "filter" {
        for_each = logging_filter.value.filters
        content {
          behavior    = filter.value.behavior
          requirement = filter.value.requirement

          dynamic "condition" {
            for_each = filter.value.conditions
            content {
              dynamic "action_condition" {
                for_each = condition.value.action != "" ? [condition.value.action] : []
                content {
                  action = action_condition.value
                }
              }

              dynamic "label_name_condition" {
                for_each = condition.value.label_name != "" ? [condition.value.label_name] : []
                content {
                  label_name = label_name_condition.value
                }
              }
            }
          }
        }
      }
    }
  }
}
