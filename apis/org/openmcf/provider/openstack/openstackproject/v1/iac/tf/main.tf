# main.tf

# Create the OpenStack Identity (Keystone) project.
# Projects are the fundamental organizational unit -- all cloud resources
# belong to a project, which provides resource isolation and quota boundaries.
resource "openstack_identity_project_v3" "main" {
  name        = var.metadata.name
  description = var.spec.description != "" ? var.spec.description : null
  enabled     = var.spec.enabled

  # Domain (optional, ForceNew).
  domain_id = var.spec.domain_id != "" ? var.spec.domain_id : null

  # Parent project (optional, ForceNew).
  parent_id = var.spec.parent_id != "" ? var.spec.parent_id : null

  # Tags (optional).
  tags = length(var.spec.tags) > 0 ? toset(var.spec.tags) : null

  # Region override (optional).
  region = var.spec.region != "" ? var.spec.region : null
}
