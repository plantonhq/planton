# Enable the Compute Engine API so a fresh project can host the policy.
# disable_on_destroy is false: tearing down one policy must never disable
# the API for everything else in the project.
resource "google_project_service" "compute_api" {
  project = local.project_id
  service = "compute.googleapis.com"

  disable_dependent_services = true
  disable_on_destroy         = false
}

# One kind, two provider resource families: Google models global and
# regional network firewall policies as separate resources with identical
# rule and association shapes. Exactly one family exists per instance,
# selected by var.spec.region (see locals.tf). The GLOBAL family follows;
# the REGIONAL family below it repeats the same shape plus `region`.
#
# On every family: name, policy_type, and project are immutable (a change
# recreates the policy with every rule and association); description
# changes in place. policy_type and project are sent only when set so the
# provider's defaults stay the provider's; deletion_policy fans to every
# resource.

# --- Global ------------------------------------------------------------------

resource "google_compute_network_firewall_policy" "this" {
  count = local.is_regional ? 0 : 1

  project = local.project_id
  name    = local.policy_name

  description     = var.spec.description != "" ? var.spec.description : null
  policy_type     = var.spec.policy_type
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]
}

# One rule resource per spec.rules entry, keyed by priority. Optional inputs
# are sent only when set so the provider's defaults stay the provider's (the
# Optional+Computed target_type and network-context enums would otherwise
# fight the API's read-back on re-plan). Booleans are sent as the spec states
# them. The security profile group and tls_inspect travel together: both are
# legal only with action apply_security_profile_group (spec CEL);
# target_forwarding_rules only with target_type INTERNAL_MANAGED_LB.
resource "google_compute_network_firewall_policy_rule" "this" {
  for_each = local.global_rules

  project         = local.project_id
  firewall_policy = one(google_compute_network_firewall_policy.this).name
  priority        = each.value.priority
  action          = each.value.action
  direction       = each.value.direction

  description    = each.value.description != "" ? each.value.description : null
  rule_name      = each.value.rule_name != "" ? each.value.rule_name : null
  disabled       = each.value.disabled
  enable_logging = each.value.enable_logging

  target_type             = each.value.target_type
  target_forwarding_rules = length(each.value.target_forwarding_rules) > 0 ? each.value.target_forwarding_rules : null
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

    src_network_context  = each.value.match.src_network_context
    dest_network_context = each.value.match.dest_network_context
  }

  deletion_policy = local.deletion_policy
}

# One association per spec.associations entry, attaching the policy to one
# VPC network in the project. name and attachment_target are immutable in
# Google: a change to either detaches the old association and creates the
# new one.
resource "google_compute_network_firewall_policy_association" "this" {
  for_each = local.global_associations

  project           = local.project_id
  firewall_policy   = one(google_compute_network_firewall_policy.this).name
  name              = each.key
  attachment_target = each.value

  deletion_policy = local.deletion_policy
}

# --- Regional ----------------------------------------------------------------

resource "google_compute_region_network_firewall_policy" "this" {
  count = local.is_regional ? 1 : 0

  project = local.project_id
  name    = local.policy_name
  region  = var.spec.region

  description     = var.spec.description != "" ? var.spec.description : null
  policy_type     = var.spec.policy_type
  deletion_policy = local.deletion_policy

  depends_on = [google_project_service.compute_api]
}

resource "google_compute_region_network_firewall_policy_rule" "this" {
  for_each = local.regional_rules

  project         = local.project_id
  region          = var.spec.region
  firewall_policy = one(google_compute_region_network_firewall_policy.this).name
  priority        = each.value.priority
  action          = each.value.action
  direction       = each.value.direction

  description    = each.value.description != "" ? each.value.description : null
  rule_name      = each.value.rule_name != "" ? each.value.rule_name : null
  disabled       = each.value.disabled
  enable_logging = each.value.enable_logging

  target_type             = each.value.target_type
  target_forwarding_rules = length(each.value.target_forwarding_rules) > 0 ? each.value.target_forwarding_rules : null
  target_service_accounts = length(each.value.target_service_accounts) > 0 ? each.value.target_service_accounts : null

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

    src_network_context  = each.value.match.src_network_context
    dest_network_context = each.value.match.dest_network_context
  }

  deletion_policy = local.deletion_policy
}

resource "google_compute_region_network_firewall_policy_association" "this" {
  for_each = local.regional_associations

  project           = local.project_id
  region            = var.spec.region
  firewall_policy   = one(google_compute_region_network_firewall_policy.this).name
  name              = each.key
  attachment_target = each.value

  deletion_policy = local.deletion_policy
}
