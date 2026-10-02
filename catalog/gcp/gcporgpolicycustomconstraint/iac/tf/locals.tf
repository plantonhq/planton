locals {
  # The constraint's name as Google knows it: custom.<constraint_name>, with
  # constraint_name defaulting to metadata.name when the spec leaves it empty.
  # The spec takes the bare name; the module owns the prefix so it can never
  # be doubled or forgotten (the Pulumi module does the same).
  constraint_name = "custom.${var.spec.constraint_name != "" ? var.spec.constraint_name : var.metadata.name}"

  parent = "organizations/${var.spec.organization_id}"
}
