locals {
  resource_name = coalesce(try(var.metadata.name, null), "cloudflare-origin-ca-certificate")

  # Defaults coalesced here so a standalone tofu apply matches the control-plane
  # middleware (and the Pulumi module) byte-for-byte.
  request_type       = try(var.spec.request_type, "") != "" ? var.spec.request_type : "origin-rsa"
  requested_validity = try(var.spec.requested_validity, 0) != 0 ? var.spec.requested_validity : 5475

  # Generate a key + CSR only when the user did not supply their own CSR.
  generate_key  = try(var.spec.csr, "") == ""
  key_algorithm = local.request_type == "origin-ecc" ? "ECDSA" : "RSA"
}
