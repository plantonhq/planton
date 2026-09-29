locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azuremonitormetricalert"
    "resource_name" = var.metadata.name
  }

  # Organization tag only if var.metadata.org is non-empty
  org_tag = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "organization" = var.metadata.org } : {}

  # Environment tag only if var.metadata.env is non-empty
  env_tag = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "environment" = var.metadata.env } : {}

  id_tag = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "resource_id" = var.metadata.id } : {}

  # Merge base, org, environment, id, and user tags -- user tags win on key
  # conflicts (the governance surface belongs to the user).
  final_tags = merge(local.base_tags, local.org_tag, local.env_tag, local.id_tag, var.spec.tags)

  # Enum wire maps. The tfvars wire format carries the FULL proto enum
  # value names -- these maps must match them verbatim.
  aggregation_map = {
    "AVERAGE" = "Average"
    "COUNT"   = "Count"
    "MINIMUM" = "Minimum"
    "MAXIMUM" = "Maximum"
    "TOTAL"   = "Total"
  }

  # One shared operator vocabulary speaks both criteria families' wire
  # dialects (the spec CELs keep each family to its legal subset).
  operator_map = {
    "EQUALS"                = "Equals"
    "GREATER_THAN"          = "GreaterThan"
    "GREATER_THAN_OR_EQUAL" = "GreaterThanOrEqual"
    "LESS_THAN"             = "LessThan"
    "LESS_THAN_OR_EQUAL"    = "LessThanOrEqual"
    "GREATER_OR_LESS_THAN"  = "GreaterOrLessThan"
  }

  dimension_operator_map = {
    "INCLUDE"     = "Include"
    "EXCLUDE"     = "Exclude"
    "STARTS_WITH" = "StartsWith"
  }

  sensitivity_map = {
    "LOW"    = "Low"
    "MEDIUM" = "Medium"
    "HIGH"   = "High"
  }
}
