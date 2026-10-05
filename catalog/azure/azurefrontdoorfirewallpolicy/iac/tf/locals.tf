locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurefrontdoorfirewallpolicy"
    "resource_name" = var.metadata.name
  }

  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Metadata-derived tags first, then the user's spec tags merged over them:
  # user tags deliberately win so an org's governance conventions (cost
  # center, owner) can override the derived values where they collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # The spec's sku enum arrives as the FULL proto value name; absent means
  # STANDARD (the tfvars wire format drops zero-valued proto fields, so the
  # module materializes the spec's documented default).
  sku_map = {
    "STANDARD" = "Standard_AzureFrontDoor"
    "PREMIUM"  = "Premium_AzureFrontDoor"
  }
  sku_name   = local.sku_map[coalesce(var.spec.sku, "STANDARD")]
  is_premium = local.sku_name == "Premium_AzureFrontDoor"

  mode_map = {
    "DETECTION"  = "Detection"
    "PREVENTION" = "Prevention"
  }

  custom_rule_type_map = {
    "MATCH_RULE"      = "MatchRule"
    "RATE_LIMIT_RULE" = "RateLimitRule"
  }

  # JSChallenge/CAPTCHA are Premium-only (spec-enforced).
  custom_rule_action_map = {
    "ALLOW"        = "Allow"
    "BLOCK"        = "Block"
    "LOG"          = "Log"
    "REDIRECT"     = "Redirect"
    "JS_CHALLENGE" = "JSChallenge"
    "CAPTCHA"      = "CAPTCHA"
  }

  match_variable_map = {
    "COOKIES"        = "Cookies"
    "POST_ARGS"      = "PostArgs"
    "QUERY_STRING"   = "QueryString"
    "REMOTE_ADDR"    = "RemoteAddr"
    "REQUEST_BODY"   = "RequestBody"
    "REQUEST_HEADER" = "RequestHeader"
    "REQUEST_METHOD" = "RequestMethod"
    "REQUEST_URI"    = "RequestUri"
    "SOCKET_ADDR"    = "SocketAddr"
  }

  operator_map = {
    "ANY"                   = "Any"
    "BEGINS_WITH"           = "BeginsWith"
    "CONTAINS"              = "Contains"
    "ENDS_WITH"             = "EndsWith"
    "EQUAL"                 = "Equal"
    "GEO_MATCH"             = "GeoMatch"
    "GREATER_THAN"          = "GreaterThan"
    "GREATER_THAN_OR_EQUAL" = "GreaterThanOrEqual"
    "IP_MATCH"              = "IPMatch"
    "LESS_THAN"             = "LessThan"
    "LESS_THAN_OR_EQUAL"    = "LessThanOrEqual"
    "REG_EX"                = "RegEx"
  }

  # Note the canonical casing is "UrlDecode"/"UrlEncode" (the SDK
  # constants' STRING VALUES) -- the provider validates case-sensitively,
  # so the SDK's URLDecode/URLEncode Go identifiers are NOT the wire
  # values.
  transform_map = {
    "LOWERCASE"    = "Lowercase"
    "REMOVE_NULLS" = "RemoveNulls"
    "TRIM"         = "Trim"
    "UPPERCASE"    = "Uppercase"
    "URL_DECODE"   = "UrlDecode"
    "URL_ENCODE"   = "UrlEncode"
  }

  # The RULE_SET_ / OVERRIDE_ / SELECTOR_ / EXCLUDE_ / SCRUB_ prefixes
  # exist only to keep the proto enum names collision-free within the
  # kind -- ARM's wire values are the bare names.
  managed_rule_set_action_map = {
    "RULE_SET_BLOCK"    = "Block"
    "RULE_SET_LOG"      = "Log"
    "RULE_SET_REDIRECT" = "Redirect"
  }

  exclusion_match_variable_map = {
    "EXCLUDE_QUERY_STRING_ARG_NAMES"      = "QueryStringArgNames"
    "EXCLUDE_REQUEST_BODY_JSON_ARG_NAMES" = "RequestBodyJsonArgNames"
    "EXCLUDE_REQUEST_BODY_POST_ARG_NAMES" = "RequestBodyPostArgNames"
    "EXCLUDE_REQUEST_COOKIE_NAMES"        = "RequestCookieNames"
    "EXCLUDE_REQUEST_HEADER_NAMES"        = "RequestHeaderNames"
  }

  selector_operator_map = {
    "SELECTOR_CONTAINS"    = "Contains"
    "SELECTOR_ENDS_WITH"   = "EndsWith"
    "SELECTOR_EQUALS"      = "Equals"
    "SELECTOR_EQUALS_ANY"  = "EqualsAny"
    "SELECTOR_STARTS_WITH" = "StartsWith"
  }

  managed_rule_override_action_map = {
    "OVERRIDE_ALLOW"           = "Allow"
    "OVERRIDE_ANOMALY_SCORING" = "AnomalyScoring"
    "OVERRIDE_BLOCK"           = "Block"
    "OVERRIDE_CAPTCHA"         = "CAPTCHA"
    "OVERRIDE_JS_CHALLENGE"    = "JSChallenge"
    "OVERRIDE_LOG"             = "Log"
    "OVERRIDE_REDIRECT"        = "Redirect"
  }

  scrubbing_match_variable_map = {
    "SCRUB_QUERY_STRING_ARG_NAMES"      = "QueryStringArgNames"
    "SCRUB_REQUEST_BODY_JSON_ARG_NAMES" = "RequestBodyJsonArgNames"
    "SCRUB_REQUEST_BODY_POST_ARG_NAMES" = "RequestBodyPostArgNames"
    "SCRUB_REQUEST_COOKIE_NAMES"        = "RequestCookieNames"
    "SCRUB_REQUEST_HEADER_NAMES"        = "RequestHeaderNames"
    "SCRUB_REQUEST_IP_ADDRESS"          = "RequestIPAddress"
    "SCRUB_REQUEST_URI"                 = "RequestUri"
  }

  custom_rules  = coalesce(var.spec.custom_rules, [])
  managed_rules = coalesce(var.spec.managed_rules, [])
}
