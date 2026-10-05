# Computed values for the KubernetesFlagdFlagFile module. Every resolution has
# an exact twin in the Pulumi module - the rendered definitions are
# byte-identical (Go json.Marshal and OpenTofu jsonencode sort keys and escape
# alike). Flags and evaluators arrive untyped (they hold free-form JSONLogic),
# so every field is read with try().

locals {
  name      = var.metadata.name
  namespace = var.spec.namespace
  key       = try(var.spec.key, "") != "" && try(var.spec.key, null) != null ? var.spec.key : "flags.flagd.json"

  labels = merge(
    {
      "planton.ai/resource" = "true"
      "planton.ai/name"     = var.metadata.name
      "planton.ai/kind"     = "KubernetesFlagdFlagFile"
    },
    try(var.metadata.id, "") != "" ? { "planton.ai/id" = var.metadata.id } : {},
    try(var.metadata.org, "") != "" ? { "planton.ai/organization" = var.metadata.org } : {},
    try(var.metadata.env, "") != "" ? { "planton.ai/environment" = var.metadata.env } : {}
  )

  definitions = { for k, v in {
    "$schema" = "https://flagd.dev/schema/v0/flags.json"
    flags = { for name, f in try(var.spec.flags, {}) : name => { for fk, fv in {
      state          = f.state
      variants       = { for vn, vv in try(f.variants, {}) : vn => try(vv.bool_value, vv.string_value, vv.number_value, vv.object_value, null) }
      defaultVariant = try(f.default_variant, "") != "" ? f.default_variant : null
      targeting      = length(try(keys(f.targeting), [])) > 0 ? f.targeting : null
      metadata       = length(try(keys(f.metadata), [])) > 0 ? f.metadata : null
    } : fk => fv if fv != null } }
    "$evaluators" = length(try(var.spec.evaluators, {})) > 0 ? var.spec.evaluators : null
    metadata      = length(try(keys(var.spec.metadata), [])) > 0 ? var.spec.metadata : null
  } : k => v if v != null }

  flag_file = jsonencode(local.definitions)
}
