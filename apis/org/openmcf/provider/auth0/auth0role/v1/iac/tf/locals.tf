# Local values for Auth0Role module

locals {
  # role_name defaults to metadata.name when spec.name is not provided.
  role_name   = coalesce(var.spec.name, var.metadata.name)
  description = var.spec.description
  permissions = coalesce(var.spec.permissions, [])
}
