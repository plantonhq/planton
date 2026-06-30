# Auth0Role Main Resources
# Creates the Auth0 role and sets its authoritative permission list.

resource "auth0_role" "this" {
  name        = local.role_name
  description = local.description
}

# Auth0 Role Permissions
# Sets the complete set of permissions for the role. auth0_role_permissions is
# authoritative — a permission removed here is removed from the role on apply.
resource "auth0_role_permissions" "this" {
  count = length(local.permissions) > 0 ? 1 : 0

  role_id = auth0_role.this.id

  dynamic "permissions" {
    for_each = local.permissions
    content {
      name                       = permissions.value.name
      resource_server_identifier = permissions.value.resource_server_identifier
    }
  }
}
