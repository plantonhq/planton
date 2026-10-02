locals {
  # The key's short name defaults to metadata.name when the spec leaves it
  # empty -- the same naming basis every kind uses.
  short_name = var.spec.short_name != "" ? var.spec.short_name : var.metadata.name

  # Google's parent argument is one string rendered from whichever owner arm
  # the spec set (exactly one, proto-CEL-enforced): organizations/{id} or
  # projects/{id}. A project_id that already carries its prefix passes
  # through so a hand-written full name still works.
  parent = (
    var.spec.parent.organization_id != ""
    ? "organizations/${var.spec.parent.organization_id}"
    : startswith(var.spec.parent.project_id, "projects/")
    ? var.spec.parent.project_id
    : "projects/${var.spec.parent.project_id}"
  )
}
