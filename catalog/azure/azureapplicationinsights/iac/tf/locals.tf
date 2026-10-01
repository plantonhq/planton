locals {
  # Identity tags -- the same keys and values the Pulumi module writes.
  # resource_kind is the CloudResourceKind enum name lowercased, spelled as
  # that exact literal; resource_id is added (id_tag below) only when the
  # resource has an id, never with the name as a stand-in.
  base_tags = {
    "resource"      = "true"
    "resource_kind" = "azureapplicationinsights"
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

  # Application-type wire values. Azure's API strings are CASE-SENSITIVE
  # and irregular ("Node.JS", "MobileCenter") -- an unmatched value would be
  # silently treated as ASP.NET by Azure, which is why the spec closes the
  # vocabulary and this map carries the exact wire strings.
  application_type_map = {
    "WEB"           = "web"
    "JAVA"          = "java"
    "NODE_JS"       = "Node.JS"
    "OTHER"         = "other"
    "IOS"           = "ios"
    "PHONE"         = "phone"
    "STORE"         = "store"
    "MOBILE_CENTER" = "MobileCenter"
  }
  application_type = (
    var.spec.application_type != null && var.spec.application_type != ""
    ? local.application_type_map[var.spec.application_type]
    : "web"
  )
}
