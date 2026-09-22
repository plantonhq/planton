# Auth0User Main Resources
# Creates the Auth0 user (minting its initial password when none is declared on
# a database connection) and sets its authoritative role and permission lists.

# The initial password, minted only when the spec declares none on a database
# connection: letters and digits only, no symbols, so it satisfies Auth0's every
# built-in password policy and never needs quoting wherever a person pastes
# it. Twin: the Pulumi module's random.RandomPassword with the same arguments.
resource "random_password" "this" {
  count = local.generate_password ? 1 : 0

  length      = 24
  special     = false
  min_upper   = 2
  min_lower   = 2
  min_numeric = 2

  # The generation-shape arguments are ignored after creation so an IMPORTED
  # credential never silently regenerates: rotation stays an explicit act in
  # Auth0, never plan fallout. Twin: the Pulumi module's IgnoreChanges on the
  # same argument set.
  lifecycle {
    ignore_changes = [
      length, special, upper, lower, numeric,
      min_lower, min_numeric, min_special, min_upper, override_special,
    ]
  }
}

# The user. Optional strings are sent only when declared (an empty value would
# reach the provider as a change from "unset" and, for the fields Auth0
# derives -- name, nickname, picture -- overwrite the derived value). The
# plain bools are sent as declared: their false equals the provider's omitted
# behavior. verify_email is sent only when stated, because its unset state
# means "Auth0 decides", a third state a plain bool cannot carry.
resource "auth0_user" "this" {
  connection_name = var.spec.connection_name

  email          = var.spec.email != "" ? var.spec.email : null
  email_verified = var.spec.email_verified
  verify_email   = var.spec.verify_email

  username    = var.spec.username != "" ? var.spec.username : null
  name        = var.spec.name != "" ? var.spec.name : null
  given_name  = var.spec.given_name != "" ? var.spec.given_name : null
  family_name = var.spec.family_name != "" ? var.spec.family_name : null
  nickname    = var.spec.nickname != "" ? var.spec.nickname : null
  picture     = var.spec.picture != "" ? var.spec.picture : null

  phone_number   = var.spec.phone_number != "" ? var.spec.phone_number : null
  phone_verified = var.spec.phone_verified

  blocked = var.spec.blocked

  user_id  = var.spec.user_id != "" ? var.spec.user_id : null
  password = local.password

  user_metadata = local.user_metadata_json
  app_metadata  = local.app_metadata_json

  custom_domain_header = var.spec.custom_domain_header != "" ? var.spec.custom_domain_header : null
}

# The authoritative role set. auth0_user_roles manages the complete list of
# roles assigned to the user -- a role omitted here is removed from the user on
# apply. Not created when the spec declares no roles.
resource "auth0_user_roles" "this" {
  count = length(local.role_ids) > 0 ? 1 : 0

  user_id = auth0_user.this.id
  roles   = local.role_ids
}

# The authoritative set of direct API permissions. auth0_user_permissions
# manages the complete list -- a permission omitted here is removed from the
# user on apply. Not created when the spec declares no permissions.
resource "auth0_user_permissions" "this" {
  count = length(local.permissions) > 0 ? 1 : 0

  user_id = auth0_user.this.id

  dynamic "permissions" {
    for_each = local.permissions
    content {
      name                       = permissions.value.name
      resource_server_identifier = permissions.value.resource_server_identifier
    }
  }
}
