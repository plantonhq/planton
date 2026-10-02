locals {
  # The full-resource-name authority for hierarchy nodes.
  crm_prefix = "//cloudresourcemanager.googleapis.com/"

  parent_project_id  = var.spec.parent != null ? trimprefix(var.spec.parent.project_id, "projects/") : ""
  parent_folder_id   = var.spec.parent != null ? trimprefix(var.spec.parent.folder_id, "folders/") : ""
  parent_org_id      = var.spec.parent != null ? var.spec.parent.organization_id : ""
  parent_resource    = var.spec.parent != null ? var.spec.parent.resource_name : ""
  is_project_parent  = local.parent_folder_id == "" && local.parent_org_id == "" && local.parent_resource == ""
  project_is_numeric = can(regex("^[0-9]+$", local.parent_project_id))

  # Google wants the project NUMBER in a binding's parent. A GcpProject
  # reference resolves to the number and a numeric literal is used as is; a
  # project ID literal, or an empty parent meaning the provider's default
  # project, is resolved through one read of the project. Count-gating the
  # lookup keeps every plan whose number is already known credential-free
  # (the Pulumi module gates its LookupProject the same way).
  needs_project_lookup = local.is_project_parent && !local.project_is_numeric

  parent = (
    local.parent_folder_id != ""
    ? "${local.crm_prefix}folders/${local.parent_folder_id}"
    : local.parent_org_id != ""
    ? "${local.crm_prefix}organizations/${local.parent_org_id}"
    : local.parent_resource != ""
    ? local.parent_resource
    : local.project_is_numeric
    ? "${local.crm_prefix}projects/${local.parent_project_id}"
    : "${local.crm_prefix}projects/${data.google_project.this[0].number}"
  )

  # spec.location selects the provider resource: the location-scoped binding
  # for a regional or zonal resource, the global binding for everything else.
  # Exactly one of the two resources below is created.
  is_location_scoped = var.spec.location != ""
}
