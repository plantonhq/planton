locals {
  # Empty strings mean "not set" for plain proto3 string fields -- drop them
  # so Cloudflare applies its own defaults.
  rate_limiting_technique = var.spec.rate_limiting_technique != "" ? var.spec.rate_limiting_technique : null
  logpush_public_key      = var.spec.logpush_public_key != "" ? var.spec.logpush_public_key : null
  workers_ai_billing_mode = var.spec.workers_ai_billing_mode != "" ? var.spec.workers_ai_billing_mode : null
  store_id                = var.spec.store_id != "" ? var.spec.store_id : null

  # The spec groups Cloudflare's flat retry_* arguments into retry{} for
  # authoring clarity -- fan them back out here.
  retry_backoff      = try(var.spec.retry.backoff, "") != "" ? var.spec.retry.backoff : null
  retry_delay        = try(var.spec.retry.delay, null)
  retry_max_attempts = try(var.spec.retry.max_attempts, null)

  # Same for log_management{} -> log_management + log_management_strategy.
  # Cloudflare echoes SERVER DEFAULTS (100000 records, DELETE_OLDEST) on
  # every read when the pair was never set (live-measured 2026-08-27), so an
  # omitted send drifts forever against the echo -- send the defaults
  # explicitly instead; semantics are unchanged because they are what
  # Cloudflare applies anyway.
  log_management          = coalesce(try(var.spec.log_management.max_records, null), 100000)
  log_management_strategy = coalesce(try(var.spec.log_management.strategy, "") != "" ? var.spec.log_management.strategy : null, "DELETE_OLDEST")

  # Cloudflare echoes these three toggles as booleans on every read (false
  # when never set), so an omitted send drifts forever against the echo --
  # always send them; proto3's unset false IS Cloudflare's default
  # (live-measured 2026-08-27).
  authentication = coalesce(try(var.spec.authentication, null), false)
  logpush        = coalesce(try(var.spec.logpush, null), false)
  zdr            = coalesce(try(var.spec.zdr, null), false)

  # DLP: drop the empty-string default action and empty collections so the
  # API sees only what the manifest states.
  dlp = var.spec.dlp == null ? null : {
    enabled  = var.spec.dlp.enabled
    action   = var.spec.dlp.action != "" ? var.spec.dlp.action : null
    profiles = length(coalesce(var.spec.dlp.profiles, [])) > 0 ? var.spec.dlp.profiles : null
    policies = length(coalesce(var.spec.dlp.policies, [])) > 0 ? [
      for policy in var.spec.dlp.policies : {
        id       = policy.id
        enabled  = policy.enabled
        action   = policy.action
        check    = policy.check
        profiles = policy.profiles
      }
    ] : null
  }

  # Guardrails: an unset control (empty string) means "do not evaluate this
  # hazard category" -- strip it rather than sending "". Objects iterate as
  # maps of their (all-string) attributes here, so the subset survives.
  guardrails = var.spec.guardrails == null ? null : {
    prompt   = { for code, control in var.spec.guardrails.prompt : code => control if control != "" }
    response = { for code, control in var.spec.guardrails.response : code => control if control != "" }
  }

  # OTel destinations: headers is Required at the API (empty map when the
  # manifest declares none); the authorization credential drops when unset.
  otel = length(coalesce(var.spec.otel, [])) > 0 ? [
    for destination in var.spec.otel : {
      url           = destination.url
      headers       = coalesce(destination.headers, {})
      authorization = destination.authorization != "" ? destination.authorization : null
      content_type  = destination.content_type != "" ? destination.content_type : null
    }
  ] : null

  stripe = var.spec.stripe == null ? null : {
    authorization = var.spec.stripe.authorization
    usage_events  = var.spec.stripe.usage_events
  }

  # Spend limits: the rule id is always sent explicitly -- the provider
  # schema's default is a leaked example value shared by every omitted id,
  # which would collapse multiple rules into one (the spec requires explicit
  # unique ids). The provider filter's wire name is "provider"; the
  # Terraform argument is ai_gateway_provider.
  spend_limits = var.spec.spend_limits == null ? null : {
    enabled = var.spec.spend_limits.enabled
    rules = length(coalesce(var.spec.spend_limits.rules, [])) > 0 ? [
      for rule in var.spec.spend_limits.rules : {
        id                 = rule.id
        enabled            = rule.enabled
        limit              = rule.limit
        limit_type         = rule.limit_type
        window             = rule.window
        technique          = rule.technique != "" ? rule.technique : null
        metadata           = length(coalesce(rule.metadata, {})) > 0 ? rule.metadata : null
        model              = rule.model
        ai_gateway_provider = rule.provider
      }
    ] : null
  }

  # Dynamic routes: one provider object per route, keyed by route name. The
  # spec's on_true/on_false edges are the wire's true/false (renamed in the
  # spec because proto cannot use boolean literals as field names); the
  # spec's properties.provider is the argument ai_gateway_dynamic_routing_provider.
  dynamic_routes = {
    for route in coalesce(var.spec.dynamic_routes, []) : route.name => [
      for element in route.elements : {
        id   = element.id
        type = element.type
        outputs = {
          next       = element.outputs.next
          "true"     = element.outputs.on_true
          "false"    = element.outputs.on_false
          success    = element.outputs.success
          fallback   = element.outputs.fallback
          element_id = element.outputs.element_id != "" ? element.outputs.element_id : null
        }
        properties = element.properties == null ? null : {
          conditions                          = element.properties.conditions != "" ? element.properties.conditions : null
          key                                 = element.properties.key != "" ? element.properties.key : null
          limit                               = element.properties.limit
          limit_type                          = element.properties.limit_type != "" ? element.properties.limit_type : null
          window                              = element.properties.window
          model                               = element.properties.model != "" ? element.properties.model : null
          ai_gateway_dynamic_routing_provider = element.properties.provider != "" ? element.properties.provider : null
          retries                             = element.properties.retries
          timeout                             = element.properties.timeout
        }
      }
    ]
  }
}
