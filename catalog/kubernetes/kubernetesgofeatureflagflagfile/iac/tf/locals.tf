# Computed values for the KubernetesGoFeatureFlagFlagFile module. Every
# resolution has an exact twin in the Pulumi module - the rendered flag file
# is byte-identical (Go json.Marshal and OpenTofu jsonencode sort keys and
# escape alike). The flags arrive untyped (variations hold free-form values),
# so every field is read with try().

locals {
  name      = var.metadata.name
  namespace = var.spec.namespace
  key       = try(var.spec.key, "") != "" && try(var.spec.key, null) != null ? var.spec.key : "flags.goff.yaml"

  labels = merge(
    {
      "planton.ai/resource" = "true"
      "planton.ai/name"     = var.metadata.name
      "planton.ai/kind"     = "KubernetesGoFeatureFlagFlagFile"
    },
    try(var.metadata.id, "") != "" ? { "planton.ai/id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  # A rule renders the same way wherever it appears (targeting, defaultRule,
  # a scheduled step's overlays). HCL has no functions, so the three call
  # sites below repeat this shape - keep them identical.
  flags = { for name, f in try(var.spec.flags, {}) : name => { for k, v in {
    variations = { for vn, vv in try(f.variations, {}) : vn => try(vv.bool_value, vv.string_value, vv.number_value, vv.object_value, vv.list_value, null) }
    defaultRule = { for rk, rv in {
      name       = try(f.default_rule.name, "") != "" ? f.default_rule.name : null
      query      = try(f.default_rule.query, "") != "" ? f.default_rule.query : null
      variation  = try(f.default_rule.variation, "") != "" ? f.default_rule.variation : null
      percentage = length(try(f.default_rule.percentage, {})) > 0 ? f.default_rule.percentage : null
      progressiveRollout = try(f.default_rule.progressive_rollout, null) == null ? null : {
        initial = { for sk, sv in { variation = f.default_rule.progressive_rollout.initial.variation, date = f.default_rule.progressive_rollout.initial.date, percentage = try(f.default_rule.progressive_rollout.initial.percentage, null) } : sk => sv if sv != null }
        end     = { for sk, sv in { variation = f.default_rule.progressive_rollout.end.variation, date = f.default_rule.progressive_rollout.end.date, percentage = try(f.default_rule.progressive_rollout.end.percentage, null) } : sk => sv if sv != null }
      }
      disable = try(f.default_rule.disable, false) == true ? true : null
    } : rk => rv if rv != null }
    targeting = length(try(f.targeting, [])) == 0 ? null : [for r in f.targeting : { for rk, rv in {
      name       = try(r.name, "") != "" ? r.name : null
      query      = try(r.query, "") != "" ? r.query : null
      variation  = try(r.variation, "") != "" ? r.variation : null
      percentage = length(try(r.percentage, {})) > 0 ? r.percentage : null
      progressiveRollout = try(r.progressive_rollout, null) == null ? null : {
        initial = { for sk, sv in { variation = r.progressive_rollout.initial.variation, date = r.progressive_rollout.initial.date, percentage = try(r.progressive_rollout.initial.percentage, null) } : sk => sv if sv != null }
        end     = { for sk, sv in { variation = r.progressive_rollout.end.variation, date = r.progressive_rollout.end.date, percentage = try(r.progressive_rollout.end.percentage, null) } : sk => sv if sv != null }
      }
      disable = try(r.disable, false) == true ? true : null
    } : rk => rv if rv != null }]
    bucketingKey    = try(f.bucketing_key, "") != "" ? f.bucketing_key : null
    trackEvents     = try(f.track_events, null)
    disable         = try(f.disable, false) == true ? true : null
    version         = try(f.version, "") != "" ? f.version : null
    metadata        = length(try(keys(f.metadata), [])) > 0 ? f.metadata : null
    experimentation = try(f.experimentation, null) == null ? null : { start = f.experimentation.start, end = f.experimentation.end }
    scheduledRollout = length(try(f.scheduled_rollout, [])) == 0 ? null : [for s in f.scheduled_rollout : { for sk, sv in {
      date       = s.date
      variations = length(try(s.variations, {})) > 0 ? { for vn, vv in s.variations : vn => try(vv.bool_value, vv.string_value, vv.number_value, vv.object_value, vv.list_value, null) } : null
      targeting = length(try(s.targeting, [])) == 0 ? null : [for r in s.targeting : { for rk, rv in {
        name       = try(r.name, "") != "" ? r.name : null
        query      = try(r.query, "") != "" ? r.query : null
        variation  = try(r.variation, "") != "" ? r.variation : null
        percentage = length(try(r.percentage, {})) > 0 ? r.percentage : null
        progressiveRollout = try(r.progressive_rollout, null) == null ? null : {
          initial = { for pk, pv in { variation = r.progressive_rollout.initial.variation, date = r.progressive_rollout.initial.date, percentage = try(r.progressive_rollout.initial.percentage, null) } : pk => pv if pv != null }
          end     = { for pk, pv in { variation = r.progressive_rollout.end.variation, date = r.progressive_rollout.end.date, percentage = try(r.progressive_rollout.end.percentage, null) } : pk => pv if pv != null }
        }
        disable = try(r.disable, false) == true ? true : null
      } : rk => rv if rv != null }]
      defaultRule = try(s.default_rule, null) == null ? null : { for rk, rv in {
        name       = try(s.default_rule.name, "") != "" ? s.default_rule.name : null
        query      = try(s.default_rule.query, "") != "" ? s.default_rule.query : null
        variation  = try(s.default_rule.variation, "") != "" ? s.default_rule.variation : null
        percentage = length(try(s.default_rule.percentage, {})) > 0 ? s.default_rule.percentage : null
        progressiveRollout = try(s.default_rule.progressive_rollout, null) == null ? null : {
          initial = { for pk, pv in { variation = s.default_rule.progressive_rollout.initial.variation, date = s.default_rule.progressive_rollout.initial.date, percentage = try(s.default_rule.progressive_rollout.initial.percentage, null) } : pk => pv if pv != null }
          end     = { for pk, pv in { variation = s.default_rule.progressive_rollout.end.variation, date = s.default_rule.progressive_rollout.end.date, percentage = try(s.default_rule.progressive_rollout.end.percentage, null) } : pk => pv if pv != null }
        }
        disable = try(s.default_rule.disable, false) == true ? true : null
      } : rk => rv if rv != null }
      trackEvents     = try(s.track_events, null)
      disable         = try(s.disable, null)
      version         = try(s.version, "") != "" ? s.version : null
      experimentation = try(s.experimentation, null) == null ? null : { start = s.experimentation.start, end = s.experimentation.end }
    } : sk => sv if sv != null }]
  } : k => v if v != null } }

  flag_file = jsonencode(local.flags)
}
