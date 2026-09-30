locals {
  # Empty strings mean "not set" for plain proto3 string fields -- drop them
  # rather than sending empty values. The private key is optional: CA uploads
  # used only to validate clients carry no key.
  name        = var.spec.name != "" ? var.spec.name : null
  private_key = var.spec.private_key != "" ? var.spec.private_key : null
}
