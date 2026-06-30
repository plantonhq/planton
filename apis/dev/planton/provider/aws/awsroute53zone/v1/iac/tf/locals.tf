locals {
  # Derive a stable resource ID
  resource_id = (
    var.metadata.id != null && var.metadata.id != ""
    ? var.metadata.id
    : var.metadata.name
  )

  # Base tags
  base_tags = {
    "resource"      = "true"
    "resource_id"   = local.resource_id
    "resource_kind" = "aws_route53_zone"
  }

  # Organization tag only if var.metadata.org is non-empty
  org_tag = (
  var.metadata.org != null && var.metadata.org != ""
  ) ? {
    "organization" = var.metadata.org
  } : {}

  # Environment tag only if var.metadata.env is non-empty
  env_tag = (
  var.metadata.env != null &&
  try(var.metadata.env, "") != ""
  ) ? {
    "environment" = var.metadata.env
  } : {}

  # Merge base, org, and environment tags
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag)

  # replace '.' with '-' in hosted zone name

  normalized_r53_zone_name = replace(var.metadata.name, ".", "-")

  # Normalize record names to produce stable keys for for_each.
  # 1. trim starting and trailing dot
  # 2. replace '.' with '_'
  # 3. replace '*' with 'wildcard'

  normalized_records = {
    for rec in var.spec.records :
    format(
      "%s_%s",
      replace(
        replace(
          trim(rec.name, "."),
          ".", "_"
        ),
        "*", "wildcard"
      ),
      rec.record_type
    ) => rec
  }
}
