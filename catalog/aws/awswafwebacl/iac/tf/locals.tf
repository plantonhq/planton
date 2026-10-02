# This module serializes the spec's typed, recursive statement tree into the
# AWS WAFv2 API JSON (PascalCase keys) that aws_wafv2_web_acl's rule_json
# argument expects. Both engines pass rules to the provider as this JSON --
# the provider's typed rule schema adds nothing here because AWS validates
# the JSON directly -- so THIS mapping (and its Pulumi twin in
# iac/pulumi/module/rules.go) is the single behavioral surface that must stay
# in lockstep with the spec.
#
# HCL cannot recurse, so the tree is transformed bottom-up through a bounded
# set of unrolled levels: level 3 is a rule's root statement, and each
# AND/OR/NOT or scope-down step descends one level, giving three levels of
# statement nesting below the root -- the same depth the AWS provider's own
# structured `rule` schema supports. Deeper trees fail the plan loudly (see
# the depth-guard precondition in main.tf); the custom_statement escape
# hatch carries arbitrary depth because it passes through verbatim.
#
# PARITY-EXCEPTION: the Pulumi module serializes the same tree with natural
# Go recursion and therefore has no depth ceiling. For any tree within three
# nesting levels (every structured configuration the AWS provider itself can
# express) the two engines emit identical JSON; a deeper tree fails Terraform
# at plan time with a clear message while Pulumi deploys it. Stack outputs
# are unaffected.
#
# Two HCL disciplines used throughout, both forced by the `rules` attribute
# being `any`-typed (a recursive tree cannot be expressed as a Terraform
# type):
#   - an unset field is ABSENT rather than null, so optional attributes are
#     read with try() defaults;
#   - statement shapes are heterogeneous, and HCL's conditional operator
#     must type-unify its two arms -- which differently-shaped objects fail.
#     Every cross-shape conditional therefore routes through
#     jsonencode()/jsondecode(): both arms become strings (always
#     unifiable), and the decoded result is the plain value. The tree is
#     jsonencode()d into rule_json at the end regardless, so this costs
#     nothing semantically.
#
# Defaults applied while serializing (identical in both engines):
#   - rule visibility_config: metrics on, sampling on, metric_name = rule name
#   - rate_based aggregate_key_type: "IP"
#   - ip_set_reference forwarded-IP position: "FIRST"
locals {
  # Resource-identity tags follow the catalog convention. The Name tag is
  # what the WAF console displays alongside the ACL.
  aws_tags = {
    "Name"                     = var.metadata.name
    "planton.ai/resource"      = "true"
    "planton.ai/organization"  = var.metadata.org
    "planton.ai/environment"   = var.metadata.env
    "planton.ai/resource-kind" = "AwsWafWebAcl"
    "planton.ai/resource-id"   = var.metadata.id
  }

  # ---------------------------------------------------------------------------
  # Statement-tree enumeration (bottom-up levels)
  #
  # Nodes are keyed by path: "<ruleIndex>" at the root, then "/and.N",
  # "/or.N", "/not", or "/sd" (scope-down) per descent. Children are
  # gathered with filter comprehensions (never conditionals) so absent arms
  # simply contribute nothing: try() turns a missing attribute chain into
  # an empty collection.
  # ---------------------------------------------------------------------------

  l3_nodes = { for rule_index, rule in var.spec.rules : tostring(rule_index) => rule.statement }

  l2_nodes = merge([
    for k, s in local.l3_nodes : merge(
      { for i, child in try(s.and_statement.statements, []) : "${k}/and.${i}" => child },
      { for i, child in try(s.or_statement.statements, []) : "${k}/or.${i}" => child },
      { for _, child in try([s.not_statement.statement], []) : "${k}/not" => child },
      { for _, child in try([s.managed_rule_group.scope_down_statement], []) : "${k}/sd" => child },
      { for _, child in try([s.rate_based.scope_down_statement], []) : "${k}/sd" => child }
    )
  ]...)

  l1_nodes = merge([
    for k, s in local.l2_nodes : merge(
      { for i, child in try(s.and_statement.statements, []) : "${k}/and.${i}" => child },
      { for i, child in try(s.or_statement.statements, []) : "${k}/or.${i}" => child },
      { for _, child in try([s.not_statement.statement], []) : "${k}/not" => child },
      { for _, child in try([s.managed_rule_group.scope_down_statement], []) : "${k}/sd" => child },
      { for _, child in try([s.rate_based.scope_down_statement], []) : "${k}/sd" => child }
    )
  ]...)

  l0_nodes = merge([
    for k, s in local.l1_nodes : merge(
      { for i, child in try(s.and_statement.statements, []) : "${k}/and.${i}" => child },
      { for i, child in try(s.or_statement.statements, []) : "${k}/or.${i}" => child },
      { for _, child in try([s.not_statement.statement], []) : "${k}/not" => child },
      { for _, child in try([s.managed_rule_group.scope_down_statement], []) : "${k}/sd" => child },
      { for _, child in try([s.rate_based.scope_down_statement], []) : "${k}/sd" => child }
    )
  ]...)

  # Level-0 nodes must be pure leaves. Any child found below them means the
  # tree is deeper than the supported three nesting levels -- the depth
  # guard in main.tf turns this into a clear plan-time error.
  depth_overflow_nodes = merge([
    for k, s in local.l0_nodes : merge(
      { for i, child in try(s.and_statement.statements, []) : "${k}/and.${i}" => true },
      { for i, child in try(s.or_statement.statements, []) : "${k}/or.${i}" => true },
      { for _, child in try([s.not_statement.statement], []) : "${k}/not" => true },
      { for _, child in try([s.managed_rule_group.scope_down_statement], []) : "${k}/sd" => true },
      { for _, child in try([s.rate_based.scope_down_statement], []) : "${k}/sd" => true }
    )
  ]...)

  all_nodes = merge(local.l3_nodes, local.l2_nodes, local.l1_nodes, local.l0_nodes)

  # ---------------------------------------------------------------------------
  # Shared per-node helpers (written once, used by every level)
  # ---------------------------------------------------------------------------

  # The one field_to_match a component-inspecting statement carries,
  # whichever arm it lives under (CEL guarantees at most one arm exists, so
  # the concat holds at most one element).
  node_ftm_source = {
    for k, s in local.all_nodes : k => concat(
      try([s.byte_match.field_to_match], []),
      try([s.sqli_match.field_to_match], []),
      try([s.xss_match.field_to_match], []),
      try([s.size_constraint.field_to_match], []),
      try([s.regex_match.field_to_match], []),
      try([s.regex_pattern_set_reference.field_to_match], [])
    )
  }

  # FieldToMatch -> AWS shape. Empty-config components are empty objects;
  # configured components carry their settings. Bool arms are CEL-pinned to
  # `true` in the spec, so a plain presence check suffices.
  node_ftm = {
    for k, sources in local.node_ftm_source : k => length(sources) == 0 ? null : jsondecode(
      try(sources[0].uri_path, false) ? jsonencode({ UriPath = {} }) :
      try(sources[0].query_string, false) ? jsonencode({ QueryString = {} }) :
      try(sources[0].method, false) ? jsonencode({ Method = {} }) :
      try(sources[0].all_query_arguments, false) ? jsonencode({ AllQueryArguments = {} }) :
      can(sources[0].single_header.name) ? jsonencode({ SingleHeader = { Name = sources[0].single_header.name } }) :
      can(sources[0].single_query_argument.name) ? jsonencode({ SingleQueryArgument = { Name = sources[0].single_query_argument.name } }) :
      try(sources[0].body, null) != null ? jsonencode({ Body = merge(
        {},
        try(sources[0].body.oversize_handling, "") != "" ? { OversizeHandling = sources[0].body.oversize_handling } : {}
      ) }) :
      try(sources[0].json_body, null) != null ? jsonencode({ JsonBody = merge(
        {
          MatchScope   = sources[0].json_body.match_scope
          MatchPattern = length(try(sources[0].json_body.included_paths, [])) > 0 ? { IncludedPaths = sources[0].json_body.included_paths } : { All = {} }
        },
        try(sources[0].json_body.invalid_fallback_behavior, "") != "" ? { InvalidFallbackBehavior = sources[0].json_body.invalid_fallback_behavior } : {},
        try(sources[0].json_body.oversize_handling, "") != "" ? { OversizeHandling = sources[0].json_body.oversize_handling } : {}
      ) }) :
      # The include/exclude key names differ between headers and cookies in
      # the AWS API (IncludedHeaders vs IncludedCookies).
      try(sources[0].headers, null) != null ? jsonencode({ Headers = {
        MatchPattern = (
          length(try(sources[0].headers.match_pattern.included_names, [])) > 0 ? { IncludedHeaders = sources[0].headers.match_pattern.included_names } :
          length(try(sources[0].headers.match_pattern.excluded_names, [])) > 0 ? { ExcludedHeaders = sources[0].headers.match_pattern.excluded_names } :
          { All = {} }
        )
        MatchScope       = sources[0].headers.match_scope
        OversizeHandling = sources[0].headers.oversize_handling
      } }) :
      try(sources[0].cookies, null) != null ? jsonencode({ Cookies = {
        MatchPattern = (
          length(try(sources[0].cookies.match_pattern.included_names, [])) > 0 ? { IncludedCookies = sources[0].cookies.match_pattern.included_names } :
          length(try(sources[0].cookies.match_pattern.excluded_names, [])) > 0 ? { ExcludedCookies = sources[0].cookies.match_pattern.excluded_names } :
          { All = {} }
        )
        MatchScope       = sources[0].cookies.match_scope
        OversizeHandling = sources[0].cookies.oversize_handling
      } }) :
      can(sources[0].header_order.oversize_handling) ? jsonencode({ HeaderOrder = { OversizeHandling = sources[0].header_order.oversize_handling } }) :
      can(sources[0].ja3_fingerprint.fallback_behavior) ? jsonencode({ JA3Fingerprint = { FallbackBehavior = sources[0].ja3_fingerprint.fallback_behavior } }) :
      can(sources[0].ja4_fingerprint.fallback_behavior) ? jsonencode({ JA4Fingerprint = { FallbackBehavior = sources[0].ja4_fingerprint.fallback_behavior } }) :
      try(sources[0].uri_fragment, null) != null ? jsonencode({ UriFragment = merge(
        {},
        try(sources[0].uri_fragment.fallback_behavior, "") != "" ? { FallbackBehavior = sources[0].uri_fragment.fallback_behavior } : {}
      ) }) :
      "null"
    )
  }

  # The one text_transformations list a component-inspecting statement
  # carries, in AWS shape (again, at most one source arm exists).
  node_text_transformations = {
    for k, s in local.all_nodes : k => [
      for transformation in concat(
        try(s.byte_match.text_transformations, []),
        try(s.sqli_match.text_transformations, []),
        try(s.xss_match.text_transformations, []),
        try(s.size_constraint.text_transformations, []),
        try(s.regex_match.text_transformations, []),
        try(s.regex_pattern_set_reference.text_transformations, [])
        ) : {
        Priority = try(transformation.priority, 0)
        Type     = transformation.type
      }
    ]
  }

  # Per-rule action overrides (managed and referenced rule groups share the
  # shape and the tuning workflow).
  node_rule_action_overrides = {
    for k, s in local.all_nodes : k => [
      for override in concat(
        try(s.managed_rule_group.rule_action_overrides, []),
        try(s.rule_group_reference.rule_action_overrides, [])
        ) : {
        Name = override.name
        ActionToUse = (
          override.action == "allow" ? { Allow = {} } :
          override.action == "block" ? { Block = {} } :
          override.action == "captcha" ? { Captcha = {} } :
          override.action == "challenge" ? { Challenge = {} } :
          { Count = {} }
        )
      }
    ]
  }

  # ATP/ACFP response inspection -- how the rule set recognizes success vs
  # failure in the application's own responses. Both rule sets share one
  # shape; the map holds the AWS-shaped block per node/config pair.
  node_response_inspection = {
    for k, s in local.all_nodes : k => {
      atp  = try(s.managed_rule_group.managed_rule_group_configs.account_takeover_prevention.response_inspection, null)
      acfp = try(s.managed_rule_group.managed_rule_group_configs.account_creation_fraud_prevention.response_inspection, null)
    }
  }

  node_response_inspection_waf = {
    for k, pair in local.node_response_inspection : k => {
      for which, ri in pair : which => merge(
        jsondecode(can(ri.status_code.success_codes) ? jsonencode({ StatusCode = {
          SuccessCodes = ri.status_code.success_codes
          FailureCodes = ri.status_code.failure_codes
        } }) : "{}"),
        jsondecode(can(ri.header.name) ? jsonencode({ Header = {
          Name          = ri.header.name
          SuccessValues = ri.header.success_values
          FailureValues = ri.header.failure_values
        } }) : "{}"),
        jsondecode(can(ri.body_json.identifier) ? jsonencode({ Json = {
          Identifier    = ri.body_json.identifier
          SuccessValues = ri.body_json.success_values
          FailureValues = ri.body_json.failure_values
        } }) : "{}"),
        jsondecode(can(ri.body_contains.success_strings) ? jsonencode({ BodyContains = {
          SuccessStrings = ri.body_contains.success_strings
          FailureStrings = ri.body_contains.failure_strings
        } }) : "{}")
      ) if ri != null
    }
  }

  # Intelligent-threat managed rule group configs (Bot Control, ATP, ACFP,
  # anti-DDoS) -- AWS models them as a LIST with one entry per rule set.
  node_managed_configs = {
    for k, s in local.all_nodes : k => concat(
      jsondecode(try(s.managed_rule_group.managed_rule_group_configs.bot_control, null) != null ? jsonencode([{
        AWSManagedRulesBotControlRuleSet = merge(
          { InspectionLevel = s.managed_rule_group.managed_rule_group_configs.bot_control.inspection_level },
          # optional bool: absent keeps the AWS default (true).
          try([{ EnableMachineLearning = s.managed_rule_group.managed_rule_group_configs.bot_control.enable_machine_learning }], [{}])[0]
        )
      }]) : "[]"),
      jsondecode(try(s.managed_rule_group.managed_rule_group_configs.account_takeover_prevention, null) != null ? jsonencode([{
        AWSManagedRulesATPRuleSet = merge(
          { LoginPath = s.managed_rule_group.managed_rule_group_configs.account_takeover_prevention.login_path },
          try(s.managed_rule_group.managed_rule_group_configs.account_takeover_prevention.enable_regex_in_path, false) ? { EnableRegexInPath = true } : {},
          try([{
            RequestInspection = {
              PayloadType   = s.managed_rule_group.managed_rule_group_configs.account_takeover_prevention.request_inspection.payload_type
              UsernameField = { Identifier = s.managed_rule_group.managed_rule_group_configs.account_takeover_prevention.request_inspection.username_field.identifier }
              PasswordField = { Identifier = s.managed_rule_group.managed_rule_group_configs.account_takeover_prevention.request_inspection.password_field.identifier }
            }
          }], [{}])[0],
          try([{ ResponseInspection = local.node_response_inspection_waf[k].atp }], [{}])[0]
        )
      }]) : "[]"),
      jsondecode(try(s.managed_rule_group.managed_rule_group_configs.account_creation_fraud_prevention, null) != null ? jsonencode([{
        AWSManagedRulesACFPRuleSet = merge(
          {
            CreationPath         = s.managed_rule_group.managed_rule_group_configs.account_creation_fraud_prevention.creation_path
            RegistrationPagePath = s.managed_rule_group.managed_rule_group_configs.account_creation_fraud_prevention.registration_page_path
          },
          try(s.managed_rule_group.managed_rule_group_configs.account_creation_fraud_prevention.enable_regex_in_path, false) ? { EnableRegexInPath = true } : {},
          {
            RequestInspection = merge(
              { PayloadType = s.managed_rule_group.managed_rule_group_configs.account_creation_fraud_prevention.request_inspection.payload_type },
              try([{ UsernameField = { Identifier = s.managed_rule_group.managed_rule_group_configs.account_creation_fraud_prevention.request_inspection.username_field.identifier } }], [{}])[0],
              try([{ PasswordField = { Identifier = s.managed_rule_group.managed_rule_group_configs.account_creation_fraud_prevention.request_inspection.password_field.identifier } }], [{}])[0],
              try([{ EmailField = { Identifier = s.managed_rule_group.managed_rule_group_configs.account_creation_fraud_prevention.request_inspection.email_field.identifier } }], [{}])[0],
              # AWS models the multi-field groups as lists of {Identifier}.
              try([{ PhoneNumberFields = [for identifier in s.managed_rule_group.managed_rule_group_configs.account_creation_fraud_prevention.request_inspection.phone_number_fields.identifiers : { Identifier = identifier }] }], [{}])[0],
              try([{ AddressFields = [for identifier in s.managed_rule_group.managed_rule_group_configs.account_creation_fraud_prevention.request_inspection.address_fields.identifiers : { Identifier = identifier }] }], [{}])[0]
            )
          },
          try([{ ResponseInspection = local.node_response_inspection_waf[k].acfp }], [{}])[0]
        )
      }]) : "[]"),
      jsondecode(try(s.managed_rule_group.managed_rule_group_configs.anti_ddos, null) != null ? jsonencode([{
        AWSManagedRulesAntiDDoSRuleSet = merge(
          {
            ClientSideActionConfig = {
              Challenge = merge(
                { UsageOfAction = s.managed_rule_group.managed_rule_group_configs.anti_ddos.client_side_action.usage_of_action },
                try(s.managed_rule_group.managed_rule_group_configs.anti_ddos.client_side_action.sensitivity, "") != "" ? { Sensitivity = s.managed_rule_group.managed_rule_group_configs.anti_ddos.client_side_action.sensitivity } : {},
                length(try(s.managed_rule_group.managed_rule_group_configs.anti_ddos.client_side_action.exempt_uri_regular_expressions, [])) > 0 ? { ExemptUriRegularExpressions = [for expression in s.managed_rule_group.managed_rule_group_configs.anti_ddos.client_side_action.exempt_uri_regular_expressions : { RegexString = expression }] } : {}
              )
            }
          },
          try(s.managed_rule_group.managed_rule_group_configs.anti_ddos.sensitivity_to_block, "") != "" ? { SensitivityToBlock = s.managed_rule_group.managed_rule_group_configs.anti_ddos.sensitivity_to_block } : {}
        )
      }]) : "[]")
    )
  }

  # Rate-based composite aggregation keys. Bool arms are CEL-pinned to
  # `true`, so a plain presence check suffices; each key contributes exactly
  # one property object.
  node_custom_keys = {
    for k, s in local.all_nodes : k => [
      for custom_key in try(s.rate_based.custom_keys, []) : jsondecode(
        can(custom_key.header.name) ? jsonencode({ Header = {
          Name                = custom_key.header.name
          TextTransformations = [for transformation in try(custom_key.header.text_transformations, []) : { Priority = try(transformation.priority, 0), Type = transformation.type }]
        } }) :
        can(custom_key.cookie.name) ? jsonencode({ Cookie = {
          Name                = custom_key.cookie.name
          TextTransformations = [for transformation in try(custom_key.cookie.text_transformations, []) : { Priority = try(transformation.priority, 0), Type = transformation.type }]
        } }) :
        can(custom_key.query_argument.name) ? jsonencode({ QueryArgument = {
          Name                = custom_key.query_argument.name
          TextTransformations = [for transformation in try(custom_key.query_argument.text_transformations, []) : { Priority = try(transformation.priority, 0), Type = transformation.type }]
        } }) :
        can(custom_key.query_string.text_transformations) ? jsonencode({ QueryString = {
          TextTransformations = [for transformation in try(custom_key.query_string.text_transformations, []) : { Priority = try(transformation.priority, 0), Type = transformation.type }]
        } }) :
        can(custom_key.uri_path.text_transformations) ? jsonencode({ UriPath = {
          TextTransformations = [for transformation in try(custom_key.uri_path.text_transformations, []) : { Priority = try(transformation.priority, 0), Type = transformation.type }]
        } }) :
        try(custom_key.http_method, false) ? jsonencode({ HTTPMethod = {} }) :
        try(custom_key.ip, false) ? jsonencode({ IP = {} }) :
        try(custom_key.forwarded_ip, false) ? jsonencode({ ForwardedIP = {} }) :
        try(custom_key.asn, false) ? jsonencode({ ASN = {} }) :
        can(custom_key.label_namespace.namespace) ? jsonencode({ LabelNamespace = { Namespace = custom_key.label_namespace.namespace } }) :
        can(custom_key.ja3_fingerprint.fallback_behavior) ? jsonencode({ JA3Fingerprint = { FallbackBehavior = custom_key.ja3_fingerprint.fallback_behavior } }) :
        can(custom_key.ja4_fingerprint.fallback_behavior) ? jsonencode({ JA4Fingerprint = { FallbackBehavior = custom_key.ja4_fingerprint.fallback_behavior } }) :
        "{}"
      )
    ]
  }

  # ---------------------------------------------------------------------------
  # Leaf transform (one pass over every node; logical nodes map to null and
  # are assembled per level below). Scope-down statements are injected during
  # level assembly, because they live one level deeper. Reference ARNs arrive
  # as plain strings: the generator flattens StringValueOrRef (the
  # orchestrator resolves any value_from before the module runs), including
  # inside the any-typed rules subtree.
  # ---------------------------------------------------------------------------

  leaf_waf = {
    for k, s in local.all_nodes : k => jsondecode(
      try(s.managed_rule_group, null) != null ? jsonencode({ ManagedRuleGroupStatement = merge(
        {
          Name       = s.managed_rule_group.name
          VendorName = s.managed_rule_group.vendor_name
        },
        try(s.managed_rule_group.version, "") != "" ? { Version = s.managed_rule_group.version } : {},
        length(local.node_rule_action_overrides[k]) > 0 ? { RuleActionOverrides = local.node_rule_action_overrides[k] } : {},
        length(local.node_managed_configs[k]) > 0 ? { ManagedRuleGroupConfigs = local.node_managed_configs[k] } : {}
      ) }) :
      try(s.rate_based, null) != null ? jsonencode({ RateBasedStatement = merge(
        {
          Limit            = s.rate_based.limit
          AggregateKeyType = try(s.rate_based.aggregate_key_type, "") != "" ? s.rate_based.aggregate_key_type : "IP"
        },
        try(s.rate_based.evaluation_window_sec, 0) > 0 ? { EvaluationWindowSec = s.rate_based.evaluation_window_sec } : {},
        length(local.node_custom_keys[k]) > 0 ? { CustomKeys = local.node_custom_keys[k] } : {},
        can(s.rate_based.forwarded_ip_config.header_name) ? { ForwardedIPConfig = {
          HeaderName       = s.rate_based.forwarded_ip_config.header_name
          FallbackBehavior = s.rate_based.forwarded_ip_config.fallback_behavior
        } } : {}
      ) }) :
      try(s.rule_group_reference, null) != null ? jsonencode({ RuleGroupReferenceStatement = merge(
        { ARN = s.rule_group_reference.arn },
        length(local.node_rule_action_overrides[k]) > 0 ? { RuleActionOverrides = local.node_rule_action_overrides[k] } : {}
      ) }) :
      try(s.ip_set_reference, null) != null ? jsonencode({ IPSetReferenceStatement = merge(
        { ARN = s.ip_set_reference.arn },
        can(s.ip_set_reference.forwarded_ip_config.header_name) ? { IPSetForwardedIPConfig = {
          HeaderName       = s.ip_set_reference.forwarded_ip_config.header_name
          FallbackBehavior = s.ip_set_reference.forwarded_ip_config.fallback_behavior
          Position         = try(s.ip_set_reference.forwarded_ip_config.position, "") != "" ? s.ip_set_reference.forwarded_ip_config.position : "FIRST"
        } } : {}
      ) }) :
      try(s.regex_pattern_set_reference, null) != null ? jsonencode({ RegexPatternSetReferenceStatement = {
        ARN                 = s.regex_pattern_set_reference.arn
        FieldToMatch        = local.node_ftm[k]
        TextTransformations = local.node_text_transformations[k]
      } }) :
      try(s.geo_match, null) != null ? jsonencode({ GeoMatchStatement = merge(
        { CountryCodes = s.geo_match.country_codes },
        can(s.geo_match.forwarded_ip_config.header_name) ? { ForwardedIPConfig = {
          HeaderName       = s.geo_match.forwarded_ip_config.header_name
          FallbackBehavior = s.geo_match.forwarded_ip_config.fallback_behavior
        } } : {}
      ) }) :
      try(s.byte_match, null) != null ? jsonencode({ ByteMatchStatement = {
        SearchString         = s.byte_match.search_string
        PositionalConstraint = s.byte_match.positional_constraint
        FieldToMatch         = local.node_ftm[k]
        TextTransformations  = local.node_text_transformations[k]
      } }) :
      try(s.sqli_match, null) != null ? jsonencode({ SqliMatchStatement = merge(
        {
          FieldToMatch        = local.node_ftm[k]
          TextTransformations = local.node_text_transformations[k]
        },
        try(s.sqli_match.sensitivity_level, "") != "" ? { SensitivityLevel = s.sqli_match.sensitivity_level } : {}
      ) }) :
      try(s.xss_match, null) != null ? jsonencode({ XssMatchStatement = {
        FieldToMatch        = local.node_ftm[k]
        TextTransformations = local.node_text_transformations[k]
      } }) :
      try(s.size_constraint, null) != null ? jsonencode({ SizeConstraintStatement = {
        ComparisonOperator  = s.size_constraint.comparison_operator
        Size                = try(s.size_constraint.size, 0)
        FieldToMatch        = local.node_ftm[k]
        TextTransformations = local.node_text_transformations[k]
      } }) :
      try(s.regex_match, null) != null ? jsonencode({ RegexMatchStatement = {
        RegexString         = s.regex_match.regex_string
        FieldToMatch        = local.node_ftm[k]
        TextTransformations = local.node_text_transformations[k]
      } }) :
      try(s.label_match, null) != null ? jsonencode({ LabelMatchStatement = {
        Scope = s.label_match.scope
        Key   = s.label_match.key
      } }) :
      try(s.asn_match, null) != null ? jsonencode({ AsnMatchStatement = merge(
        { AsnList = s.asn_match.asn_list },
        can(s.asn_match.forwarded_ip_config.header_name) ? { ForwardedIPConfig = {
          HeaderName       = s.asn_match.forwarded_ip_config.header_name
          FallbackBehavior = s.asn_match.forwarded_ip_config.fallback_behavior
        } } : {}
      ) }) :
      # Escape hatch: the user writes PascalCase keys matching the AWS API,
      # so the payload passes through verbatim.
      try(s.custom_statement, null) != null ? jsonencode(s.custom_statement) :
      "null"
    )
  }

  # ---------------------------------------------------------------------------
  # Level assembly (bottom-up). A node is either logical (AND/OR/NOT over
  # children one level down), a leaf whose scope-down child sits one level
  # down (injected generically via the leaf's single wrapper key), or a
  # plain leaf.
  # ---------------------------------------------------------------------------

  waf_l0 = { for k, s in local.l0_nodes : k => local.leaf_waf[k] }

  waf_l1 = {
    for k, s in local.l1_nodes : k => jsondecode(
      try(s.and_statement, null) != null ? jsonencode({ AndStatement = { Statements = [for i in range(length(s.and_statement.statements)) : local.waf_l0["${k}/and.${i}"]] } }) :
      try(s.or_statement, null) != null ? jsonencode({ OrStatement = { Statements = [for i in range(length(s.or_statement.statements)) : local.waf_l0["${k}/or.${i}"]] } }) :
      try(s.not_statement, null) != null ? jsonencode({ NotStatement = { Statement = local.waf_l0["${k}/not"] } }) :
      contains(keys(local.waf_l0), "${k}/sd") ? jsonencode({ (keys(local.leaf_waf[k])[0]) = merge(values(local.leaf_waf[k])[0], { ScopeDownStatement = local.waf_l0["${k}/sd"] }) }) :
      jsonencode(local.leaf_waf[k])
    )
  }

  waf_l2 = {
    for k, s in local.l2_nodes : k => jsondecode(
      try(s.and_statement, null) != null ? jsonencode({ AndStatement = { Statements = [for i in range(length(s.and_statement.statements)) : local.waf_l1["${k}/and.${i}"]] } }) :
      try(s.or_statement, null) != null ? jsonencode({ OrStatement = { Statements = [for i in range(length(s.or_statement.statements)) : local.waf_l1["${k}/or.${i}"]] } }) :
      try(s.not_statement, null) != null ? jsonencode({ NotStatement = { Statement = local.waf_l1["${k}/not"] } }) :
      contains(keys(local.waf_l1), "${k}/sd") ? jsonencode({ (keys(local.leaf_waf[k])[0]) = merge(values(local.leaf_waf[k])[0], { ScopeDownStatement = local.waf_l1["${k}/sd"] }) }) :
      jsonencode(local.leaf_waf[k])
    )
  }

  waf_l3 = {
    for k, s in local.l3_nodes : k => jsondecode(
      try(s.and_statement, null) != null ? jsonencode({ AndStatement = { Statements = [for i in range(length(s.and_statement.statements)) : local.waf_l2["${k}/and.${i}"]] } }) :
      try(s.or_statement, null) != null ? jsonencode({ OrStatement = { Statements = [for i in range(length(s.or_statement.statements)) : local.waf_l2["${k}/or.${i}"]] } }) :
      try(s.not_statement, null) != null ? jsonencode({ NotStatement = { Statement = local.waf_l2["${k}/not"] } }) :
      contains(keys(local.waf_l2), "${k}/sd") ? jsonencode({ (keys(local.leaf_waf[k])[0]) = merge(values(local.leaf_waf[k])[0], { ScopeDownStatement = local.waf_l2["${k}/sd"] }) }) :
      jsonencode(local.leaf_waf[k])
    )
  }

  # ---------------------------------------------------------------------------
  # Rule assembly. Rules are any-typed (see the header), so every optional
  # attribute is read through try(); jsonencode arms keep the heterogeneous
  # merge fragments type-consistent.
  # ---------------------------------------------------------------------------

  rules_waf = [
    for rule_index, rule in var.spec.rules : merge(
      {
        Name      = rule.name
        Priority  = rule.priority
        Statement = local.waf_l3[tostring(rule_index)]
        # Rule visibility defaults: metrics on, sampling on, metric name =
        # rule name (identical defaults in the Pulumi module).
        VisibilityConfig = {
          CloudWatchMetricsEnabled = try(rule.visibility_config, null) == null ? true : try(rule.visibility_config.cloudwatch_metrics_enabled, false)
          SampledRequestsEnabled   = try(rule.visibility_config, null) == null ? true : try(rule.visibility_config.sampled_requests_enabled, false)
          MetricName               = try(rule.visibility_config.metric_name, "") != "" ? rule.visibility_config.metric_name : rule.name
        }
      },
      # Exactly one of action / override_action is present (CEL-enforced):
      # match rules carry an action, group rules an override action.
      jsondecode(try(rule.action, "") != "" ? jsonencode({ Action = merge(
        rule.action == "block" ? { Block = try([{ CustomResponse = merge(
          { ResponseCode = rule.custom_response.response_code },
          try(rule.custom_response.custom_response_body_key, "") != "" ? { CustomResponseBodyKey = rule.custom_response.custom_response_body_key } : {},
          length(try(rule.custom_response.response_headers, [])) > 0 ? { ResponseHeaders = [for header in rule.custom_response.response_headers : { Name = header.name, Value = header.value }] } : {}
        ) }], [{}])[0] } : {},
        rule.action == "allow" ? { Allow = length(try(rule.custom_request_headers, [])) > 0 ? { CustomRequestHandling = { InsertHeaders = [for header in rule.custom_request_headers : { Name = header.name, Value = header.value }] } } : {} } : {},
        rule.action == "count" ? { Count = length(try(rule.custom_request_headers, [])) > 0 ? { CustomRequestHandling = { InsertHeaders = [for header in rule.custom_request_headers : { Name = header.name, Value = header.value }] } } : {} } : {},
        rule.action == "captcha" ? { Captcha = length(try(rule.custom_request_headers, [])) > 0 ? { CustomRequestHandling = { InsertHeaders = [for header in rule.custom_request_headers : { Name = header.name, Value = header.value }] } } : {} } : {},
        rule.action == "challenge" ? { Challenge = length(try(rule.custom_request_headers, [])) > 0 ? { CustomRequestHandling = { InsertHeaders = [for header in rule.custom_request_headers : { Name = header.name, Value = header.value }] } } : {} } : {}
      ) }) : "{}"),
      jsondecode(try(rule.override_action, "") != "" ? jsonencode({ OverrideAction = rule.override_action == "count" ? { Count = {} } : { None = {} } }) : "{}"),
      jsondecode(length(try(rule.rule_labels, [])) > 0 ? jsonencode({ RuleLabels = [for label in rule.rule_labels : { Name = label }] }) : "{}"),
      jsondecode(can(rule.captcha_config.immunity_time_sec) ? jsonencode({ CaptchaConfig = { ImmunityTimeProperty = { ImmunityTime = rule.captcha_config.immunity_time_sec } } }) : "{}"),
      jsondecode(can(rule.challenge_config.immunity_time_sec) ? jsonencode({ ChallengeConfig = { ImmunityTimeProperty = { ImmunityTime = rule.challenge_config.immunity_time_sec } } }) : "{}")
    )
  ]
}
