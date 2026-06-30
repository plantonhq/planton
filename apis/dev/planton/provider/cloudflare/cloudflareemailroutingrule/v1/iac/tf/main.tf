# A single Email Routing rule: match messages and drop / forward / hand to a
# Worker. Requires Email Routing to be enabled on the zone.
resource "cloudflare_email_routing_rule" "main" {
  zone_id  = var.spec.zone_id
  name     = try(var.spec.name, "") != "" ? var.spec.name : null
  enabled  = var.spec.enabled
  priority = var.spec.priority

  matchers = [for m in var.spec.matchers : {
    type  = m.type
    field = try(m.field, "") != "" ? m.field : null
    value = try(m.value, "") != "" ? m.value : null
  }]

  actions = [{
    type  = var.spec.action.type
    value = length(local.action_values) > 0 ? local.action_values : null
  }]
}
