# The generated key material (pools.go), self-contained so a test can apply
# this file alone, with no cluster, and parse its pools with upstream's own
# code: the five certificate authorities and the actor token signing key, in
# upstream's exact pool wire format, kept in state (a re-apply never replaces
# them; a destroy does, exactly like an upstream reinstall).
#
# A PEM block's body is the base64 of its DER, so a pool field is the PEM
# with its armour and line breaks removed. The key ID is base64url (no
# padding) of the SHA-256 of the public key's PEM text: upstream names a key
# by hashing its DER, which Terraform cannot (it hashes only text); the pool
# accepts any ID, the published key set carries whatever ID the pool names,
# and both engines derive this one from the text they share.
#
# The key set (jwksJSON) is upstream's JWKS of the pool, byte for byte, in
# its field order (so a format string, not jsonencode, which sorts keys).
# Terraform has no bytes, so the public key's DER is decoded as ISO-8859-1,
# one character per byte, and each coordinate is cut out by regex, which
# counts characters (substr and length count grapheme clusters, and read a
# CR LF byte pair as one). The tls provider's keys have fixed layouts: a
# P-256 SubjectPublicKeyInfo is 91 bytes with x at 27 and y at 59, 32 bytes
# each; an RSA-4096 one is 550 bytes with the modulus at 33 (after its 0x00
# sign byte), 512 bytes, and the exponent at 547, 3 bytes.
#
# Validity is ten years, not upstream's 365 days: every connection that
# chains to an expired root fails, and this provider would replace an
# expired root on the next apply, rotating trust out from under every issued
# certificate.

locals {
  ca_pools = {
    "service-dns-ca-pool"  = { namespace = "podcertificate-controller-system", algorithm = "ED25519", common_name = "Agent Substrate service DNS CA" }
    "pod-identity-ca-pool" = { namespace = "podcertificate-controller-system", algorithm = "ED25519", common_name = "Agent Substrate pod identity CA" }
    "postgres-ca-pool"     = { namespace = "podcertificate-controller-system", algorithm = "ED25519", common_name = "Agent Substrate PostgreSQL client CA" }
    "actor-id-ca-pool"     = { namespace = "ate-system", algorithm = "ED25519", common_name = "Agent Substrate actor identity CA" }
    "egress-mitm-ca-pool"  = { namespace = "ate-system", algorithm = "ECDSA", common_name = "Agent Substrate egress interception CA" }
  }
  actor_jwt_algorithm = upper(coalesce(try(var.spec.api_server.actor_jwt.algorithm, ""), "es256"))
}

resource "tls_private_key" "ca" {
  for_each = local.ca_pools

  algorithm   = each.value.algorithm
  ecdsa_curve = each.value.algorithm == "ECDSA" ? "P256" : null
}

resource "tls_self_signed_cert" "ca" {
  for_each = local.ca_pools

  private_key_pem       = tls_private_key.ca[each.key].private_key_pem
  validity_period_hours = 87600
  is_ca_certificate     = true
  allowed_uses          = ["digital_signature", "cert_signing"]

  subject {
    common_name = each.value.common_name
  }
}

# A key's algorithm cannot change in place, so changing it replaces the key.
resource "tls_private_key" "actor_jwt" {
  algorithm   = local.actor_jwt_algorithm == "RS256" ? "RSA" : "ECDSA"
  ecdsa_curve = local.actor_jwt_algorithm == "RS256" ? null : "P256"
  rsa_bits    = local.actor_jwt_algorithm == "RS256" ? 4096 : null
}

locals {
  ca_pool_data = {
    for name, pool in local.ca_pools : name => {
      pool = jsonencode({
        CAs = [{
          ID                 = "1"
          SigningKeyPKCS8    = replace(replace(tls_private_key.ca[name].private_key_pem_pkcs8, "/-----(BEGIN|END) [A-Z ]+-----/", ""), "/\\s+/", "")
          RootCertificateDER = replace(replace(tls_self_signed_cert.ca[name].cert_pem, "/-----(BEGIN|END) [A-Z ]+-----/", ""), "/\\s+/", "")
        }]
        ActiveForSigning = "1"
      })
      "tls.crt" = tls_self_signed_cert.ca[name].cert_pem
      "tls.key" = tls_private_key.ca[name].private_key_pem_pkcs8
    }
  }

  actor_jwt_key_id = replace(replace(replace(base64sha256(tls_private_key.actor_jwt.public_key_pem), "+", "-"), "/[/]/", "_"), "/=+$/", "")
  actor_jwt_pool = jsonencode({
    Authorities = [{
      ID              = local.actor_jwt_key_id
      Algorithm       = local.actor_jwt_algorithm
      SigningKeyPKCS8 = replace(replace(tls_private_key.actor_jwt.private_key_pem_pkcs8, "/-----(BEGIN|END) [A-Z ]+-----/", ""), "/\\s+/", "")
    }]
    ActiveForSigning = local.actor_jwt_key_id
  })

  actor_jwt_spki = textdecodebase64(replace(replace(tls_private_key.actor_jwt.public_key_pem, "/-----(BEGIN|END) [A-Z ]+-----/", ""), "/\\s+/", ""), "ISO-8859-1")
  actor_jwt_parts = {
    for name, at in local.actor_jwt_algorithm == "RS256" ? tomap({ n = [33, 512], e = [547, 3] }) : tomap({ x = [27, 32], y = [59, 32] }) :
    name => replace(replace(replace(textencodebase64(regex("^(?s).{${at[0]}}(.{${at[1]}})", local.actor_jwt_spki)[0], "ISO-8859-1"), "+", "-"), "/[/]/", "_"), "/=+$/", "")
  }
  actor_jwt_jwks = local.actor_jwt_algorithm == "RS256" ? format(
    "{\"keys\":[{\"kty\":\"RSA\",\"kid\":\"%s\",\"use\":\"sig\",\"alg\":\"%s\",\"n\":\"%s\",\"e\":\"%s\"}]}",
    local.actor_jwt_key_id, local.actor_jwt_algorithm, local.actor_jwt_parts.n, local.actor_jwt_parts.e,
    ) : format(
    "{\"keys\":[{\"kty\":\"EC\",\"kid\":\"%s\",\"use\":\"sig\",\"alg\":\"%s\",\"crv\":\"P-256\",\"x\":\"%s\",\"y\":\"%s\"}]}",
    local.actor_jwt_key_id, local.actor_jwt_algorithm, local.actor_jwt_parts.x, local.actor_jwt_parts.y,
  )
}
