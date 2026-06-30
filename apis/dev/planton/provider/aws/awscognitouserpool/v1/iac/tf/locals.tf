locals {
  name = var.metadata.name
  spec = var.spec

  tags = {
    "planton.dev/resource"      = "true"
    "planton.dev/organization"  = var.metadata.org
    "planton.dev/environment"   = var.metadata.env
    "planton.dev/resource-kind" = "AwsCognitoUserPool"
    "planton.dev/resource-id"   = var.metadata.id
  }

  # Build client map keyed by name for for_each.
  client_map = { for c in try(local.spec.clients, []) : c.name => c }

  # Determine if domain is configured.
  has_domain    = try(local.spec.domain.domain, "") != ""
  is_custom_domain = local.has_domain && can(regex("\\.", local.spec.domain.domain))
}
