# DigitalOcean Certificate Resource
#
# This module implements the discriminated union pattern for DigitalOcean certificates:
# - type = "lets_encrypt": Fully automated, auto-renewing certificates (requires DigitalOcean DNS)
# - type = "custom": User-provided certificates (bring your own cert)
#
# The resource conditionally sets fields based on the certificate type to match
# the DigitalOcean API requirements.

resource "digitalocean_certificate" "certificate" {
  name = var.spec.certificate_name
  type = local.cert_type

  # Let's Encrypt configuration (only used when type = "lets_encrypt")
  # The domains field is required for Let's Encrypt certificates
  domains = local.is_lets_encrypt ? local.le_domains : null

  # Custom certificate configuration (only used when type = "custom")
  # All three fields are required for custom certificates to work correctly
  leaf_certificate  = local.is_custom ? local.custom_leaf_cert : null
  private_key       = local.is_custom ? local.custom_private_key : null
  certificate_chain = local.is_custom && local.custom_cert_chain != "" ? local.custom_cert_chain : null

  # Lifecycle management for zero-downtime certificate rotation
  # When a custom certificate is replaced (e.g., expiring cert), create the new one
  # before destroying the old one to prevent service disruption
  lifecycle {
    create_before_destroy = true
  }
}

