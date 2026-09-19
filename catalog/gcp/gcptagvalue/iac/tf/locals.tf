locals {
  # The value's short name defaults to metadata.name when the spec leaves it
  # empty -- the same naming basis every kind uses.
  short_name = var.spec.short_name != "" ? var.spec.short_name : var.metadata.name
}
