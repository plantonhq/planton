locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CatalogKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azurewebapplicationfirewallpolicy"
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

  # Metadata-derived tags first, then the user's spec tags merged over
  # them: user tags deliberately win so an org's governance conventions
  # can override the derived values where they collide.
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # The spec's enums arrive as FULL proto value names (the tfvars wire
  # format never strips prefixes); each map below carries the complete
  # verbatim vocabulary for its enum, mapped to ARM's values. A missing
  # entry would silently drop the setting, so the maps are exhaustive by
  # construction.
  rule_type_map = {
    "MATCH_RULE"      = "MatchRule"
    "RATE_LIMIT_RULE" = "RateLimitRule"
  }

  custom_rule_action_map = {
    "ALLOW"        = "Allow"
    "BLOCK"        = "Block"
    "LOG"          = "Log"
    "JS_CHALLENGE" = "JSChallenge"
  }

  rate_limit_duration_map = {
    "ONE_MIN"   = "OneMin"
    "FIVE_MINS" = "FiveMins"
  }

  group_rate_limit_by_map = {
    "CLIENT_ADDR"             = "ClientAddr"
    "CLIENT_ADDR_XFF_HEADER"  = "ClientAddrXFFHeader"
    "GEO_LOCATION"            = "GeoLocation"
    "GEO_LOCATION_XFF_HEADER" = "GeoLocationXFFHeader"
    "NONE"                    = "None"
  }

  match_variable_name_map = {
    "REMOTE_ADDR"     = "RemoteAddr"
    "REQUEST_METHOD"  = "RequestMethod"
    "QUERY_STRING"    = "QueryString"
    "POST_ARGS"       = "PostArgs"
    "REQUEST_URI"     = "RequestUri"
    "REQUEST_HEADERS" = "RequestHeaders"
    "REQUEST_BODY"    = "RequestBody"
    "REQUEST_COOKIES" = "RequestCookies"
  }

  match_operator_map = {
    "ANY"                   = "Any"
    "IP_MATCH"              = "IPMatch"
    "GEO_MATCH"             = "GeoMatch"
    "EQUAL"                 = "Equal"
    "CONTAINS"              = "Contains"
    "LESS_THAN"             = "LessThan"
    "GREATER_THAN"          = "GreaterThan"
    "LESS_THAN_OR_EQUAL"    = "LessThanOrEqual"
    "GREATER_THAN_OR_EQUAL" = "GreaterThanOrEqual"
    "BEGINS_WITH"           = "BeginsWith"
    "ENDS_WITH"             = "EndsWith"
    "REGEX"                 = "Regex"
  }

  transform_map = {
    "HTML_ENTITY_DECODE" = "HtmlEntityDecode"
    "LOWERCASE"          = "Lowercase"
    "REMOVE_NULLS"       = "RemoveNulls"
    "TRIM"               = "Trim"
    "URL_DECODE"         = "UrlDecode"
    "URL_ENCODE"         = "UrlEncode"
    "UPPERCASE"          = "Uppercase"
  }

  # Unspecified rule-set type means OWASP (azurerm's own default,
  # materialized identically on both engines).
  managed_rule_set_type_map = {
    "OWASP"                          = "OWASP"
    "MICROSOFT_BOT_MANAGER_RULE_SET" = "Microsoft_BotManagerRuleSet"
    "MICROSOFT_DEFAULT_RULE_SET"     = "Microsoft_DefaultRuleSet"
  }

  rule_override_action_map = {
    "OVERRIDE_ALLOW"           = "Allow"
    "OVERRIDE_ANOMALY_SCORING" = "AnomalyScoring"
    "OVERRIDE_BLOCK"           = "Block"
    "OVERRIDE_JS_CHALLENGE"    = "JSChallenge"
    "OVERRIDE_LOG"             = "Log"
  }

  exclusion_match_variable_map = {
    "REQUEST_ARG_KEYS"      = "RequestArgKeys"
    "REQUEST_ARG_NAMES"     = "RequestArgNames"
    "REQUEST_ARG_VALUES"    = "RequestArgValues"
    "REQUEST_COOKIE_KEYS"   = "RequestCookieKeys"
    "REQUEST_COOKIE_NAMES"  = "RequestCookieNames"
    "REQUEST_COOKIE_VALUES" = "RequestCookieValues"
    "REQUEST_HEADER_KEYS"   = "RequestHeaderKeys"
    "REQUEST_HEADER_NAMES"  = "RequestHeaderNames"
    "REQUEST_HEADER_VALUES" = "RequestHeaderValues"
  }

  selector_match_operator_map = {
    "SELECTOR_EQUALS"      = "Equals"
    "SELECTOR_CONTAINS"    = "Contains"
    "SELECTOR_STARTS_WITH" = "StartsWith"
    "SELECTOR_ENDS_WITH"   = "EndsWith"
    "SELECTOR_EQUALS_ANY"  = "EqualsAny"
  }

  scrubbing_match_variable_map = {
    "SCRUB_REQUEST_ARG_NAMES"      = "RequestArgNames"
    "SCRUB_REQUEST_COOKIE_NAMES"   = "RequestCookieNames"
    "SCRUB_REQUEST_HEADER_NAMES"   = "RequestHeaderNames"
    "SCRUB_REQUEST_IP_ADDRESS"     = "RequestIPAddress"
    "SCRUB_REQUEST_JSON_ARG_NAMES" = "RequestJSONArgNames"
    "SCRUB_REQUEST_POST_ARG_NAMES" = "RequestPostArgNames"
  }

  mode_map = {
    "PREVENTION" = "Prevention"
    "DETECTION"  = "Detection"
  }
}
