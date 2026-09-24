# Local values for the Auth0User module. Every StringValueOrRef in the spec is
# already flattened to a plain string by the tfvars generator (the platform
# resolves references before the module runs), so the locals read strings.

locals {
  # The provider takes the two metadata documents as JSON strings. An absent
  # document stays null so the argument is left unset rather than sending "{}"
  # and having Auth0 record an authored-looking empty object. Twin: the Pulumi
  # module's renderMetadata.
  user_metadata_json = var.spec.user_metadata != null && length(var.spec.user_metadata) > 0 ? jsonencode(var.spec.user_metadata) : null
  app_metadata_json  = var.spec.app_metadata != null && length(var.spec.app_metadata) > 0 ? jsonencode(var.spec.app_metadata) : null

  # A declared password wins; a passwordless connection takes none; a database
  # user with neither gets one minted below. Same predicate as the Pulumi
  # module's GeneratePassword.
  generate_password = !var.spec.passwordless && var.spec.password == ""

  # The password the user is created with: null on a passwordless connection
  # (Auth0 refuses one), the minted value, or the declared one.
  password = var.spec.passwordless ? null : (local.generate_password ? random_password.this[0].result : var.spec.password)

  # Authoritative sets; an empty list means the companion resource is not
  # created at all rather than created empty.
  role_ids    = coalesce(var.spec.roles, [])
  permissions = coalesce(var.spec.permissions, [])
}
