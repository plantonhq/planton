locals {
  # The console display name defaults to metadata.name when the spec leaves
  # display_name empty -- the same naming basis every kind uses.
  display_name = var.spec.display_name != "" ? var.spec.display_name : var.metadata.name

  # Google's parent argument is one string rendered from whichever arm the
  # spec set (exactly one, proto-CEL-enforced): organizations/{id} for a
  # top-level folder, folders/{id} for a nested one. A folder_id that already
  # carries its prefix passes through so a hand-written full name still works.
  parent = (
    var.spec.parent.organization_id != ""
    ? "organizations/${var.spec.parent.organization_id}"
    : startswith(var.spec.parent.folder_id, "folders/")
    ? var.spec.parent.folder_id
    : "folders/${var.spec.parent.folder_id}"
  )

  # The destroy guard: GCP's default of true when the spec is silent. Always
  # sent explicitly so the spec is the single source of truth on both engines
  # (the Pulumi module checks presence the same way).
  deletion_protection = coalesce(var.spec.deletion_protection, true)
}
