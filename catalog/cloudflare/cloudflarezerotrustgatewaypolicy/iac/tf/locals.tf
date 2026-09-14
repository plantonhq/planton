locals {
  # Resource naming
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-zt-gateway-policy")

  # Labels
  labels = merge({
    "name" = local.resource_name
  }, try(var.metadata.labels, {}))

  # Cloudflare models the filter as a list that can only contain a single value
  # (the constraint is API-enforced, not schema-enforced); the spec's singular
  # string keeps the escape hatch closed.
  filters = var.spec.filter != "" ? [var.spec.filter] : null

  # Expiration: settable only for dns rules; expires_at is required inside.
  expiration = var.spec.expiration == null ? null : {
    expires_at = var.spec.expiration.expires_at
    duration   = var.spec.expiration.duration
  }

  # Schedule: a day left empty deactivates the policy that day; empty strings
  # are dropped rather than sent.
  schedule = var.spec.schedule == null ? null : {
    mon       = var.spec.schedule.mon != "" ? var.spec.schedule.mon : null
    tue       = var.spec.schedule.tue != "" ? var.spec.schedule.tue : null
    wed       = var.spec.schedule.wed != "" ? var.spec.schedule.wed : null
    thu       = var.spec.schedule.thu != "" ? var.spec.schedule.thu : null
    fri       = var.spec.schedule.fri != "" ? var.spec.schedule.fri : null
    sat       = var.spec.schedule.sat != "" ? var.spec.schedule.sat : null
    sun       = var.spec.schedule.sun != "" ? var.spec.schedule.sun : null
    time_zone = var.spec.schedule.time_zone != "" ? var.spec.schedule.time_zone : null
  }

  # ALWAYS an object, never null: the provider's own test fixtures set
  # rule_settings = {} explicitly "to prevent API drift" -- omitting the block
  # makes every apply show settings drift. Every field is built through try()
  # so a wholly-absent spec rule_settings collapses each attribute to null
  # (HCL's {}) while keeping ONE object type (a `{} : {...}` conditional would
  # make Terraform reject the mismatched arms). The spec's wrapper map values
  # ({values = [...]}) unwrap to the provider's plain map-of-lists here.
  rule_settings = {
    add_headers        = try(length(var.spec.rule_settings.add_headers) > 0 ? { for header, wrapped in var.spec.rule_settings.add_headers : header => wrapped.values } : null, null)
    allow_child_bypass = try(var.spec.rule_settings.allow_child_bypass, null)
    audit_ssh          = try(var.spec.rule_settings.audit_ssh, null)
    biso_admin_controls = try(var.spec.rule_settings.biso_admin_controls == null ? null : {
      version  = var.spec.rule_settings.biso_admin_controls.version != "" ? var.spec.rule_settings.biso_admin_controls.version : null
      copy     = var.spec.rule_settings.biso_admin_controls.copy != "" ? var.spec.rule_settings.biso_admin_controls.copy : null
      download = var.spec.rule_settings.biso_admin_controls.download != "" ? var.spec.rule_settings.biso_admin_controls.download : null
      paste    = var.spec.rule_settings.biso_admin_controls.paste != "" ? var.spec.rule_settings.biso_admin_controls.paste : null
      keyboard = var.spec.rule_settings.biso_admin_controls.keyboard != "" ? var.spec.rule_settings.biso_admin_controls.keyboard : null
      printing = var.spec.rule_settings.biso_admin_controls.printing != "" ? var.spec.rule_settings.biso_admin_controls.printing : null
      upload   = var.spec.rule_settings.biso_admin_controls.upload != "" ? var.spec.rule_settings.biso_admin_controls.upload : null
      dcp      = var.spec.rule_settings.biso_admin_controls.dcp
      dd       = var.spec.rule_settings.biso_admin_controls.dd
      dk       = var.spec.rule_settings.biso_admin_controls.dk
      dp       = var.spec.rule_settings.biso_admin_controls.dp
      du       = var.spec.rule_settings.biso_admin_controls.du
      wm_id    = var.spec.rule_settings.biso_admin_controls.wm_id != "" ? var.spec.rule_settings.biso_admin_controls.wm_id : null
    }, null)
    block_page         = try(var.spec.rule_settings.block_page, null)
    block_page_enabled = try(var.spec.rule_settings.block_page_enabled, null)
    block_reason       = try(var.spec.rule_settings.block_reason != "" ? var.spec.rule_settings.block_reason : null, null)
    bypass_parent_rule = try(var.spec.rule_settings.bypass_parent_rule, null)
    check_session = try(var.spec.rule_settings.check_session == null ? null : {
      duration = var.spec.rule_settings.check_session.duration != "" ? var.spec.rule_settings.check_session.duration : null
      enforce  = var.spec.rule_settings.check_session.enforce
    }, null)
    delete_headers = try(length(var.spec.rule_settings.delete_headers) > 0 ? var.spec.rule_settings.delete_headers : null, null)
    # Custom upstream resolvers. vnet_id arrives as a plain string (the spec's
    # reference is resolved before the module runs).
    dns_resolvers = try(var.spec.rule_settings.dns_resolvers == null ? null : {
      ipv4 = length(var.spec.rule_settings.dns_resolvers.ipv4) > 0 ? [
        for resolver in var.spec.rule_settings.dns_resolvers.ipv4 : {
          ip                            = resolver.ip
          port                          = resolver.port
          route_through_private_network = resolver.route_through_private_network ? true : null
          vnet_id                       = resolver.vnet_id != "" ? resolver.vnet_id : null
        }
      ] : null
      ipv6 = length(var.spec.rule_settings.dns_resolvers.ipv6) > 0 ? [
        for resolver in var.spec.rule_settings.dns_resolvers.ipv6 : {
          ip                            = resolver.ip
          port                          = resolver.port
          route_through_private_network = resolver.route_through_private_network ? true : null
          vnet_id                       = resolver.vnet_id != "" ? resolver.vnet_id : null
        }
      ] : null
    }, null)
    egress = try(var.spec.rule_settings.egress == null ? null : {
      ipv4          = var.spec.rule_settings.egress.ipv4 != "" ? var.spec.rule_settings.egress.ipv4 : null
      ipv4_fallback = var.spec.rule_settings.egress.ipv4_fallback != "" ? var.spec.rule_settings.egress.ipv4_fallback : null
      ipv6          = var.spec.rule_settings.egress.ipv6 != "" ? var.spec.rule_settings.egress.ipv6 : null
    }, null)
    forensic_copy                      = try(var.spec.rule_settings.forensic_copy, null)
    ignore_cname_category_matches      = try(var.spec.rule_settings.ignore_cname_category_matches, null)
    insecure_disable_dnssec_validation = try(var.spec.rule_settings.insecure_disable_dnssec_validation, null)
    ip_categories                      = try(var.spec.rule_settings.ip_categories, null)
    ip_indicator_feeds                 = try(var.spec.rule_settings.ip_indicator_feeds, null)
    l4override = try(var.spec.rule_settings.l4override == null ? null : {
      ip   = var.spec.rule_settings.l4override.ip != "" ? var.spec.rule_settings.l4override.ip : null
      port = var.spec.rule_settings.l4override.port
    }, null)
    notification_settings = try(var.spec.rule_settings.notification_settings == null ? null : {
      enabled         = var.spec.rule_settings.notification_settings.enabled
      include_context = var.spec.rule_settings.notification_settings.include_context
      msg             = var.spec.rule_settings.notification_settings.msg != "" ? var.spec.rule_settings.notification_settings.msg : null
      support_url     = var.spec.rule_settings.notification_settings.support_url != "" ? var.spec.rule_settings.notification_settings.support_url : null
    }, null)
    override_host = try(var.spec.rule_settings.override_host != "" ? var.spec.rule_settings.override_host : null, null)
    override_ips  = try(length(var.spec.rule_settings.override_ips) > 0 ? var.spec.rule_settings.override_ips : null, null)
    payload_log   = try(var.spec.rule_settings.payload_log, null)
    quarantine    = try(var.spec.rule_settings.quarantine, null)
    redirect      = try(var.spec.rule_settings.redirect, null)
    resolve_dns_internally = try(var.spec.rule_settings.resolve_dns_internally == null ? null : {
      fallback = var.spec.rule_settings.resolve_dns_internally.fallback != "" ? var.spec.rule_settings.resolve_dns_internally.fallback : null
      view_id  = var.spec.rule_settings.resolve_dns_internally.view_id != "" ? var.spec.rule_settings.resolve_dns_internally.view_id : null
    }, null)
    resolve_dns_through_cloudflare = try(var.spec.rule_settings.resolve_dns_through_cloudflare, null)
    set_headers                    = try(length(var.spec.rule_settings.set_headers) > 0 ? { for header, wrapped in var.spec.rule_settings.set_headers : header => wrapped.values } : null, null)
    untrusted_cert = try(var.spec.rule_settings.untrusted_cert == null ? null : {
      action = var.spec.rule_settings.untrusted_cert.action != "" ? var.spec.rule_settings.untrusted_cert.action : null
    }, null)
  }
}
