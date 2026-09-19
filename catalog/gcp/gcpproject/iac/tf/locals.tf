locals {
  # display_name falls back to metadata.name — explicit conditional, so
  # both engines derive the identical display name.
  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  base_labels = {
    "planton-ai_resource" = "true"
    "planton-ai_name"     = var.metadata.name
    "planton-ai_kind"     = "gcpproject"
  }

  org_label = (
    var.metadata.org != null && var.metadata.org != ""
  ) ? { "planton-ai_organization" = var.metadata.org } : {}

  env_label = (
    var.metadata.env != null && var.metadata.env != ""
  ) ? { "planton-ai_environment" = var.metadata.env } : {}

  id_label = (
    var.metadata.id != null && var.metadata.id != ""
  ) ? { "planton-ai_id" = var.metadata.id } : {}

  # User labels first so platform attribution labels win on key conflicts —
  # identical merge order to the Pulumi module.
  final_labels = merge(
    var.spec.labels,
    local.base_labels,
    local.org_label,
    local.env_label,
    local.id_label,
  )

  # Exactly one of org_id / folder_id is sent. A folder_id reference (a
  # GcpFolder's folder_id output, or a literal folder id) IS the parent and
  # wins; otherwise parent_type selects which argument parent_id fills. The
  # spec's CEL keeps the two forms from being combined. The same rule lives in
  # the Pulumi module's project.go.
  folder_ref       = trimprefix(var.spec.folder_id, "folders/")
  parent_org_id    = local.folder_ref == "" && var.spec.parent_type == "organization" ? var.spec.parent_id : null
  parent_folder_id = local.folder_ref != "" ? local.folder_ref : (var.spec.parent_type == "folder" ? var.spec.parent_id : null)

  billing_account_id = var.spec.billing_account_id != "" ? var.spec.billing_account_id : null

  # DELETE is the provider default; passing it explicitly keeps destroy
  # semantics identical on both engines (the bridged provider would
  # otherwise apply its own client-side default).
  deletion_policy = var.spec.deletion_policy != "" ? var.spec.deletion_policy : "DELETE"
}
