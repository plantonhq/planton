locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-zero-trust-access-application")

  # The type enum flattens to its string name; unspecified maps to self_hosted.
  app_type = (
    try(var.spec.type, "") == "" || var.spec.type == "application_type_unspecified"
  ) ? "self_hosted" : var.spec.type

  # Policies are referenced by ID; the proto's `policy` value carries the policy ID.
  policies = [
    for p in try(var.spec.policies, []) : {
      id         = p.policy
      precedence = try(p.precedence, 0) > 0 ? p.precedence : null
    }
  ]

  # Rebuild the provider's target_attributes map ({name => values}) from the
  # proto's [{name, values}] list (proto maps can't carry repeated values).
  target_criteria = [
    for tc in try(var.spec.target_criteria, []) : {
      port              = tc.port
      protocol          = tc.protocol
      target_attributes = { for a in tc.target_attributes : a.name => a.values }
    }
  ]
}
