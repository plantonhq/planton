# main.tf

# Create the OpenStack Identity application credential.
# Application credentials allow passwordless authentication for automation.
# All fields are ForceNew -- any change destroys and recreates the credential.
resource "openstack_identity_application_credential_v3" "main" {
  name         = var.metadata.name
  description  = var.spec.description != "" ? var.spec.description : null
  unrestricted = var.spec.unrestricted

  # User-provided secret (optional; auto-generated if null).
  secret = var.spec.secret != "" ? var.spec.secret : null

  # Roles (optional; inherits all roles if empty).
  roles = length(var.spec.roles) > 0 ? toset(var.spec.roles) : null

  # Access rules (optional; unrestricted API access if empty).
  dynamic "access_rules" {
    for_each = var.spec.access_rules
    content {
      path    = access_rules.value.path
      method  = access_rules.value.method
      service = access_rules.value.service
    }
  }

  # Expiration (optional).
  expires_at = var.spec.expires_at != "" ? var.spec.expires_at : null

  # Region override (optional).
  region = var.spec.region != "" ? var.spec.region : null
}
