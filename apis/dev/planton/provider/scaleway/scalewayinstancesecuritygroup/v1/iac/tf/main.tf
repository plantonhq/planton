# Scaleway Instance Security Group
#
# A stateful (by default) firewall that controls inbound and outbound traffic
# to Scaleway Instances. Assigned to individual Instances via the Instance's
# security_group_id field -- not attached to networks.
resource "scaleway_instance_security_group" "sg" {
  name        = local.sg_name
  description = local.description != "" ? local.description : null
  zone        = local.zone
  tags        = local.standard_tags

  # Stateful mode: when true, return traffic for accepted connections is
  # automatically permitted.
  stateful = local.stateful

  # Default policies: what happens to traffic that matches no rule.
  # "accept" = allow-all by default (denylist model)
  # "drop"   = deny-all by default (allowlist model)
  inbound_default_policy  = local.inbound_default_policy
  outbound_default_policy = local.outbound_default_policy

  # SMTP security: blocks ports 25, 465, 587 when true.
  enable_default_security = local.enable_default_security

  # ── Inbound rules ──────────────────────────────────────────────────────
  dynamic "inbound_rule" {
    for_each = local.inbound_rules
    content {
      action     = inbound_rule.value.action
      protocol   = inbound_rule.value.protocol
      port_range = inbound_rule.value.port_range != "" ? inbound_rule.value.port_range : null
      ip_range   = inbound_rule.value.ip_range != "" ? inbound_rule.value.ip_range : null
    }
  }

  # ── Outbound rules ─────────────────────────────────────────────────────
  dynamic "outbound_rule" {
    for_each = local.outbound_rules
    content {
      action     = outbound_rule.value.action
      protocol   = outbound_rule.value.protocol
      port_range = outbound_rule.value.port_range != "" ? outbound_rule.value.port_range : null
      ip_range   = outbound_rule.value.ip_range != "" ? outbound_rule.value.ip_range : null
    }
  }
}
