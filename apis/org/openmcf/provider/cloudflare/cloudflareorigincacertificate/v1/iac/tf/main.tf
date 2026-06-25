# Origin CA certificate. When no CSR is supplied, generate a private key (keyed to
# request_type) and a CSR for the requested hostnames; the generated key is exported
# as a sensitive output. When a CSR is supplied, the user's key never leaves their
# control and the private_key output is empty. Mirrors the Pulumi module exactly.

resource "tls_private_key" "origin" {
  count = local.generate_key ? 1 : 0

  algorithm   = local.key_algorithm
  rsa_bits    = local.key_algorithm == "RSA" ? 2048 : null
  ecdsa_curve = local.key_algorithm == "ECDSA" ? "P256" : null
}

resource "tls_cert_request" "origin" {
  count = local.generate_key ? 1 : 0

  private_key_pem = tls_private_key.origin[0].private_key_pem

  subject {
    common_name = var.spec.hostnames[0]
  }

  dns_names = var.spec.hostnames
}

resource "cloudflare_origin_ca_certificate" "main" {
  csr                = local.generate_key ? tls_cert_request.origin[0].cert_request_pem : var.spec.csr
  hostnames          = var.spec.hostnames
  request_type       = local.request_type
  requested_validity = local.requested_validity
}
