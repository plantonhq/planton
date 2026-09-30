locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azureapplicationgateway"
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

  # On the v2 platform (and Basic), SKU name and tier carry the same
  # value.
  sku_map = {
    "BASIC"       = "Basic"
    "STANDARD_V2" = "Standard_v2"
    "WAF_V2"      = "WAF_v2"
  }

  protocol_map = {
    "HTTP"  = "Http"
    "HTTPS" = "Https"
    "TCP"   = "Tcp"
    "TLS"   = "Tls"
  }

  rule_type_map = {
    "BASIC_ROUTING"      = "Basic"
    "PATH_BASED_ROUTING" = "PathBasedRouting"
  }

  ip_allocation_map = {
    "DYNAMIC" = "Dynamic"
    "STATIC"  = "Static"
  }

  identity_type_map = {
    "SYSTEM_ASSIGNED"          = "SystemAssigned"
    "USER_ASSIGNED"            = "UserAssigned"
    "SYSTEM_AND_USER_ASSIGNED" = "SystemAssigned, UserAssigned"
  }

  ssl_policy_type_map = {
    "PREDEFINED" = "Predefined"
    "CUSTOM"     = "Custom"
    "CUSTOM_V2"  = "CustomV2"
  }

  tls_protocol_map = {
    "TLS_V1_0" = "TLSv1_0"
    "TLS_V1_1" = "TLSv1_1"
    "TLS_V1_2" = "TLSv1_2"
    "TLS_V1_3" = "TLSv1_3"
  }

  revocation_check_map = {
    "OCSP" = "OCSP"
  }

  redirect_type_map = {
    "PERMANENT" = "Permanent"
    "FOUND"     = "Found"
    "SEE_OTHER" = "SeeOther"
    "TEMPORARY" = "Temporary"
  }

  url_component_map = {
    "PATH_ONLY"         = "path_only"
    "QUERY_STRING_ONLY" = "query_string_only"
  }

  status_code_map = {
    "HTTP_STATUS_400" = "HttpStatus400"
    "HTTP_STATUS_403" = "HttpStatus403"
    "HTTP_STATUS_404" = "HttpStatus404"
    "HTTP_STATUS_405" = "HttpStatus405"
    "HTTP_STATUS_408" = "HttpStatus408"
    "HTTP_STATUS_500" = "HttpStatus500"
    "HTTP_STATUS_502" = "HttpStatus502"
    "HTTP_STATUS_503" = "HttpStatus503"
    "HTTP_STATUS_504" = "HttpStatus504"
  }

  # The gateway's single IP configuration, derived from the dedicated
  # subnet -- pure ARM plumbing users never name.
  gateway_ip_configuration_name = "${var.spec.name}-gateway-ip"
}
