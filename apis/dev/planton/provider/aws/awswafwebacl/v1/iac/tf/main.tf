# ──────────────────────────────────────────────────────────────────────────────
# WAFv2 Web ACL
# ──────────────────────────────────────────────────────────────────────────────

resource "aws_wafv2_web_acl" "this" {
  name        = var.metadata.name
  scope       = local.scope
  description = try(var.spec.description, null)

  default_action {
    dynamic "allow" {
      for_each = local.default_action_type == "allow" ? [1] : []
      content {}
    }
    dynamic "block" {
      for_each = local.default_action_type == "block" ? [1] : []
      content {}
    }
  }

  # Rules are passed as JSON for uniform handling of typed and custom statements.
  rule_json = length(local.rules) > 0 ? jsonencode([
    for rule in local.rules : {
      Name     = rule.name
      Priority = rule.priority

      Statement = try(rule.managed_rule_group, null) != null ? {
        ManagedRuleGroupStatement = merge(
          {
            Name       = rule.managed_rule_group.name
            VendorName = rule.managed_rule_group.vendor_name
          },
          try(rule.managed_rule_group.version, "") != "" ? { Version = rule.managed_rule_group.version } : {},
          length(try(rule.managed_rule_group.rule_action_overrides, [])) > 0 ? {
            RuleActionOverrides = [
              for o in rule.managed_rule_group.rule_action_overrides : {
                Name        = o.name
                ActionToUse = { for k, v in { (title(o.action)) = {} } : k => v }
              }
            ]
          } : {},
          try(rule.managed_rule_group.scope_down_statement, null) != null ? {
            ScopeDownStatement = rule.managed_rule_group.scope_down_statement
          } : {}
        )
      } : try(rule.rate_based, null) != null ? {
        RateBasedStatement = merge(
          {
            Limit            = rule.rate_based.limit
            AggregateKeyType = coalesce(try(rule.rate_based.aggregate_key_type, null), "IP")
          },
          try(rule.rate_based.evaluation_window_sec, 0) > 0 ? {
            EvaluationWindowSec = rule.rate_based.evaluation_window_sec
          } : {},
          try(rule.rate_based.forwarded_ip_config, null) != null ? {
            ForwardedIPConfig = {
              HeaderName       = rule.rate_based.forwarded_ip_config.header_name
              FallbackBehavior = rule.rate_based.forwarded_ip_config.fallback_behavior
            }
          } : {},
          try(rule.rate_based.scope_down_statement, null) != null ? {
            ScopeDownStatement = rule.rate_based.scope_down_statement
          } : {}
        )
      } : try(rule.geo_match, null) != null ? {
        GeoMatchStatement = merge(
          { CountryCodes = rule.geo_match.country_codes },
          try(rule.geo_match.forwarded_ip_config, null) != null ? {
            ForwardedIPConfig = {
              HeaderName       = rule.geo_match.forwarded_ip_config.header_name
              FallbackBehavior = rule.geo_match.forwarded_ip_config.fallback_behavior
            }
          } : {}
        )
      } : try(rule.ip_set_reference, null) != null ? {
        IPSetReferenceStatement = merge(
          { ARN = rule.ip_set_reference.arn },
          try(rule.ip_set_reference.forwarded_ip_config, null) != null ? {
            IPSetForwardedIPConfig = {
              HeaderName       = rule.ip_set_reference.forwarded_ip_config.header_name
              FallbackBehavior = rule.ip_set_reference.forwarded_ip_config.fallback_behavior
              Position         = coalesce(try(rule.ip_set_reference.forwarded_ip_config.position, null), "FIRST")
            }
          } : {}
        )
      } : try(rule.custom_statement, null) != null ? rule.custom_statement : {}

      Action = try(rule.action, "") != "" ? {
        for k, v in { (title(rule.action)) = merge(
          {},
          rule.action == "block" && try(rule.custom_response, null) != null ? {
            CustomResponse = merge(
              { ResponseCode = rule.custom_response.response_code },
              try(rule.custom_response.custom_response_body_key, "") != "" ? {
                CustomResponseBodyKey = rule.custom_response.custom_response_body_key
              } : {}
            )
          } : {}
        ) } : k => v
      } : null

      OverrideAction = try(rule.override_action, "") != "" ? (
        rule.override_action == "count" ? { Count = {} } : { None = {} }
      ) : null

      VisibilityConfig = {
        CloudWatchMetricsEnabled = try(rule.visibility_config.cloudwatch_metrics_enabled, true)
        SampledRequestsEnabled   = try(rule.visibility_config.sampled_requests_enabled, true)
        MetricName               = try(rule.visibility_config.metric_name, rule.name)
      }
    }
  ]) : null

  # Custom response bodies.
  dynamic "custom_response_body" {
    for_each = local.custom_response_bodies
    content {
      key          = custom_response_body.key
      content      = custom_response_body.value.content
      content_type = custom_response_body.value.content_type
    }
  }

  # Token domains.
  token_domains = try(var.spec.token_domains, null)

  visibility_config {
    cloudwatch_metrics_enabled = local.acl_metrics_enabled
    sampled_requests_enabled   = local.acl_sampled_enabled
    metric_name                = local.acl_metric_name
  }

  tags = local.tags
}

# ──────────────────────────────────────────────────────────────────────────────
# WAFv2 Web ACL Logging Configuration (optional)
# ──────────────────────────────────────────────────────────────────────────────

resource "aws_wafv2_web_acl_logging_configuration" "this" {
  count = local.logging_enabled ? 1 : 0

  resource_arn            = aws_wafv2_web_acl.this.arn
  log_destination_configs = [var.spec.logging.destination_arn]

  # Redacted fields: single headers.
  dynamic "redacted_fields" {
    for_each = try(var.spec.logging.redacted_header_names, [])
    content {
      single_header {
        name = lower(redacted_fields.value)
      }
    }
  }

  # Redacted fields: URI path.
  dynamic "redacted_fields" {
    for_each = try(var.spec.logging.redact_uri_path, false) ? [1] : []
    content {
      uri_path {}
    }
  }

  # Redacted fields: query string.
  dynamic "redacted_fields" {
    for_each = try(var.spec.logging.redact_query_string, false) ? [1] : []
    content {
      query_string {}
    }
  }
}
