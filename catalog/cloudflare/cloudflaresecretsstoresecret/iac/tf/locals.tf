locals {
  # Empty string means "not set" for the optional comment -- drop it rather
  # than sending an empty value.
  comment = var.spec.comment != "" ? var.spec.comment : null
}
