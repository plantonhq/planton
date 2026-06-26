# Cloudflare Zero Trust Access policy: an account-scoped decision (allow / deny /
# non_identity / bypass) plus the rules that determine who it applies to. Attach it
# to one or more Access applications by referencing this policy's ID.
resource "cloudflare_zero_trust_access_policy" "main" {
  account_id = var.spec.account_id
  name       = var.spec.name
  decision   = var.spec.decision

  include = local.include
  exclude = length(local.exclude) > 0 ? local.exclude : null
  require = length(local.require) > 0 ? local.require : null

  session_duration = var.spec.session_duration != "" ? var.spec.session_duration : null

  approval_required              = var.spec.approval_required
  approval_groups                = length(local.approval_groups) > 0 ? local.approval_groups : null
  isolation_required             = var.spec.isolation_required
  purpose_justification_required = var.spec.purpose_justification_required
  purpose_justification_prompt   = var.spec.purpose_justification_prompt != "" ? var.spec.purpose_justification_prompt : null

  connection_rules = local.connection_rules
  mfa_config       = local.mfa_config
}
