# Auth0User Outputs
# Maps to the Auth0UserStackOutputs protobuf message

output "user_id" {
  description = "The user's full identity-provider subject, connection prefix included (e.g. auth0|66f1c2d3...) -- the sub claim in every token issued for the user"
  # The resource ID is the full subject for every user, whether Auth0 assigned
  # the id or the spec declared its unprefixed half; the user_id attribute
  # reads back the same value, but the ID is the one the provider guarantees.
  value = auth0_user.this.id
}

output "email" {
  description = "The user's email address as Auth0 stored it (lowercase)"
  value       = auth0_user.this.email
}

output "username" {
  description = "The user's login name, when the connection requires one"
  value       = auth0_user.this.username
}

output "name" {
  description = "The user's full display name as stored, including the default Auth0 derived when none was declared"
  value       = auth0_user.this.name
}

output "nickname" {
  description = "The user's short name as stored, including the default Auth0 derived when none was declared"
  value       = auth0_user.this.nickname
}

output "picture" {
  description = "The URL of the user's avatar as stored, including the placeholder Auth0 assigned when none was declared"
  value       = auth0_user.this.picture
}

output "connection_name" {
  description = "The name of the connection the user belongs to"
  value       = auth0_user.this.connection_name
}

output "password" {
  description = "The initial password the module generated -- set only when spec.password was left empty on a database connection; a declared password is never echoed back and a passwordless user has none"
  # Reads the random_password leaf directly (never a conditional over the
  # whole resource) so the output carries only the leaf's own sensitivity.
  value     = local.generate_password ? random_password.this[0].result : null
  sensitive = true
}
