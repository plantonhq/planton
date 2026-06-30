# main.tf

# Create the OpenStack Identity role assignment.
# Binds a role to a principal (user or group) on a scope (project or domain).
# All fields are ForceNew -- any change recreates the assignment.
resource "openstack_identity_role_assignment_v3" "main" {
  role_id = var.spec.role_id

  # Scope: exactly one of project_id or domain_id.
  project_id = var.spec.project_id != "" ? var.spec.project_id : null
  domain_id  = var.spec.domain_id != "" ? var.spec.domain_id : null

  # Principal: exactly one of user_id or group_id.
  user_id  = var.spec.user_id != "" ? var.spec.user_id : null
  group_id = var.spec.group_id != "" ? var.spec.group_id : null

  # Region override (optional).
  region = var.spec.region != "" ? var.spec.region : null
}
