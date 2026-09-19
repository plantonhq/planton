# The hierarchical firewall policy container: the organization- or
# folder-owned object the rules and associations hang off. Google names it
# with a server-assigned numeric ID (its `name`), which the rules and
# associations reference and the policy_id output carries.
#
# parent and short_name are immutable: changing either recreates the policy,
# and with it every rule and association. description changes in place.
# deletion_policy is sent only when set so the provider's default (DELETE)
# stays the provider's.
resource "google_compute_firewall_policy" "this" {
  parent     = local.parent
  short_name = local.short_name

  description     = var.spec.description != "" ? var.spec.description : null
  deletion_policy = local.deletion_policy
}

# One rule resource per spec.rules entry, keyed by priority (see locals.tf).
#
# Optional inputs are sent only when set so the provider's defaults stay the
# provider's: the two Optional+Computed network-context enums would otherwise
# fight the API's read-back on re-plan. Booleans are sent as the spec states
# them (false is the API's own default and reads back as such). The security
# profile group and tls_inspect travel together: both are legal only with
# action apply_security_profile_group (spec CEL).
resource "google_compute_firewall_policy_rule" "this" {
  for_each = local.rules_by_priority

  firewall_policy = google_compute_firewall_policy.this.name
  priority        = each.value.priority
  action          = each.value.action
  direction       = each.value.direction

  description    = each.value.description != "" ? each.value.description : null
  disabled       = each.value.disabled
  enable_logging = each.value.enable_logging

  # Empty means every VM in every network beneath the association's node.
  target_resources        = length(each.value.target_resources) > 0 ? each.value.target_resources : null
  target_service_accounts = length(each.value.target_service_accounts) > 0 ? each.value.target_service_accounts : null

  # Each secure tag is a tag value's tagValues/{id} name; the block's `state`
  # is read-only and left to the API.
  dynamic "target_secure_tags" {
    for_each = each.value.target_secure_tags
    content {
      name = target_secure_tags.value
    }
  }

  security_profile_group = each.value.security_profile_group != "" ? each.value.security_profile_group : null
  tls_inspect            = each.value.security_profile_group != "" ? each.value.tls_inspect : null

  match {
    dynamic "layer4_configs" {
      for_each = each.value.match.layer4_configs
      content {
        ip_protocol = layer4_configs.value.ip_protocol
        ports       = length(layer4_configs.value.ports) > 0 ? layer4_configs.value.ports : null
      }
    }

    src_ip_ranges             = length(each.value.match.src_ip_ranges) > 0 ? each.value.match.src_ip_ranges : null
    dest_ip_ranges            = length(each.value.match.dest_ip_ranges) > 0 ? each.value.match.dest_ip_ranges : null
    src_address_groups        = length(each.value.match.src_address_groups) > 0 ? each.value.match.src_address_groups : null
    dest_address_groups       = length(each.value.match.dest_address_groups) > 0 ? each.value.match.dest_address_groups : null
    src_fqdns                 = length(each.value.match.src_fqdns) > 0 ? each.value.match.src_fqdns : null
    dest_fqdns                = length(each.value.match.dest_fqdns) > 0 ? each.value.match.dest_fqdns : null
    src_region_codes          = length(each.value.match.src_region_codes) > 0 ? each.value.match.src_region_codes : null
    dest_region_codes         = length(each.value.match.dest_region_codes) > 0 ? each.value.match.dest_region_codes : null
    src_threat_intelligences  = length(each.value.match.src_threat_intelligences) > 0 ? each.value.match.src_threat_intelligences : null
    dest_threat_intelligences = length(each.value.match.dest_threat_intelligences) > 0 ? each.value.match.dest_threat_intelligences : null
    src_networks              = length(each.value.match.src_networks) > 0 ? each.value.match.src_networks : null

    dynamic "src_secure_tags" {
      for_each = each.value.match.src_secure_tags
      content {
        name = src_secure_tags.value
      }
    }

    # Optional+Computed: null stays out of the payload so Google's unset
    # value never shows as a diff on re-plan.
    src_network_context  = each.value.match.src_network_context
    dest_network_context = each.value.match.dest_network_context
  }

  deletion_policy = local.deletion_policy
}

# One association per spec.associations entry, attaching the policy to the
# organization or a folder so every network beneath that node enforces it.
# name and attachment_target are immutable in Google: a change to either
# detaches the old association and creates the new one.
resource "google_compute_firewall_policy_association" "this" {
  for_each = local.associations_by_name

  firewall_policy   = google_compute_firewall_policy.this.name
  name              = each.key
  attachment_target = each.value.target

  deletion_policy = local.deletion_policy
}
