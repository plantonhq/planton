variable "metadata" {
  description = "Cloud resource metadata"
  type = object({
    name        = string
    id          = optional(string, "")
    org         = optional(string, "")
    env         = optional(string, "")
    labels      = optional(map(string), {})
    annotations = optional(map(string), {})
    tags        = optional(list(string), [])
  })
}

variable "spec" {
  description = "GcpPrivateCaPool specification"
  type = object({
    # The GCP project the pool lives in: a literal project ID or a GcpProject reference.
    # If omitted, the provider's default project is used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The CA Service region, e.g. "us-central1"; the pool's authorities and certificates live in the same region. Immutable.
    location = string

    # The pool's ID, unique in its parent: letters, digits, "-" and "_",
    # at most 63 characters. Defaults to metadata.name. Immutable.
    ca_pool_id = optional(string, "")

    # ENTERPRISE or DEVOPS (see the message comment). Immutable.
    tier = string

    # The policy every certificate the pool issues follows. Omit for no
    # policy: any key, both request forms, any identity, no baseline values.
    issuance_policy = optional(object({
      # The key types a request's public key must match. Empty admits any key.
      allowed_key_types = optional(list(object({
        # An RSA key; an empty message admits any RSA key the service accepts.
        rsa = optional(object({
          # Smallest allowed modulus (inclusive), e.g. 2048.
          min_modulus_size = optional(number, 0)

          # Largest allowed modulus (inclusive), e.g. 4096.
          max_modulus_size = optional(number, 0)
        }))

        # An elliptic-curve key for this signature algorithm:
        #   ECDSA_P256 -- the common choice for TLS
        #   ECDSA_P384
        #   EDDSA_25519
        elliptic_curve_signature_algorithm = optional(string, "")
      })), [])

      # The longest lifetime an issued certificate may have, e.g. 7776000s
      # (90 days). A certificate is also cut short when its issuing authority
      # expires first.
      maximum_lifetime = optional(string, "")

      # Backdate every issued certificate's not-before time by this much (the
      # lifetime is preserved), to absorb clock skew on clients, e.g. 3600s.
      # Google allows at most 48 hours.
      backdate_duration = optional(string, "")

      # Which request forms are allowed. Omit to allow both.
      allowed_issuance_modes = optional(object({
        # Requests that carry a PEM certificate signing request (pem_csr).
        allow_csr_based_issuance = optional(bool, false)

        # Requests that describe the certificate as structured config (subject,
        # SANs, public key) instead of a CSR.
        allow_config_based_issuance = optional(bool, false)
      }))

      # Limits on the identities certificates may carry. Omit for no limits.
      identity_constraints = optional(object({
        # Copy the Subject from the certificate request into the certificate;
        # false discards the requested Subject.
        allow_subject_passthrough = optional(bool, false)

        # Copy the subject alternative names from the request; false discards
        # them.
        allow_subject_alt_names_passthrough = optional(bool, false)

        # A CEL expression over the resolved subject and subject alternative
        # names that must hold before a certificate is signed, e.g.
        # subject_alt_names.all(san, san.type == DNS && san.value.endsWith(".internal.example.com")).
        cel_expression = optional(object({
          # The expression.
          expression = string

          # A short title for the expression.
          title = optional(string, "")

          # What the expression checks.
          description = optional(string, "")

          # Where the expression came from, for error messages (a file and line).
          location = optional(string, "")
        }))
      }))

      # X.509 values stamped onto every certificate the pool issues.
      baseline_values = optional(object({
        # What the certificate's key may be used for (the key usage and extended
        # key usage extensions). Omit for no key-usage statement; both usage
        # groups are optional and an empty group states nothing.
        key_usage = optional(object({
          # The key usage extension's bits.
          base_key_usage = optional(object({
            # The key may make digital signatures other than certificate and CRL
            # signatures (TLS handshakes, signed tokens).
            digital_signature = optional(bool, false)

            # The key may protect content against the signer later denying it
            # (non-repudiation).
            content_commitment = optional(bool, false)

            # The key may encipher other keys (RSA key transport in TLS).
            key_encipherment = optional(bool, false)

            # The key may encipher raw data directly.
            data_encipherment = optional(bool, false)

            # The key may be used in key agreement (ECDH).
            key_agreement = optional(bool, false)

            # The key may sign certificates -- set on every CA certificate.
            cert_sign = optional(bool, false)

            # The key may sign certificate revocation lists -- set on every CA
            # certificate that publishes CRLs.
            crl_sign = optional(bool, false)

            # With key_agreement, the key may only encipher.
            encipher_only = optional(bool, false)

            # With key_agreement, the key may only decipher.
            decipher_only = optional(bool, false)
          }))

          # The extended key usage extension's well-known purposes.
          extended_key_usage = optional(object({
            # TLS server authentication -- what an HTTPS or gRPC server certificate
            # needs.
            server_auth = optional(bool, false)

            # TLS client authentication -- what an mTLS client certificate needs.
            client_auth = optional(bool, false)

            # Code signing.
            code_signing = optional(bool, false)

            # S/MIME email protection.
            email_protection = optional(bool, false)

            # Trusted timestamping.
            time_stamping = optional(bool, false)

            # Signing OCSP responses.
            ocsp_signing = optional(bool, false)
          }))

          # Extended key usages with no named field here, each an OID path.
          unknown_extended_key_usages = optional(list(object({
            # The OID's arcs, most significant first.
            object_id_path = list(number)
          })), [])
        }))

        # The basic constraints extension: whether the certificate is a CA and
        # how many CA levels may sit below it. Omit to leave the extension to
        # Google's default (for a leaf certificate, is_ca false).
        ca_options = optional(object({
          # true makes a CA certificate (CA:TRUE), false states CA:FALSE; unset
          # leaves the CA flag out of the extension.
          is_ca = optional(bool)

          # The path length constraint: how many CA levels may sit below this
          # certificate. 0 means it may sign only leaf certificates -- a
          # subordinate that issues workload certificates; unset means no limit
          # is stated.
          max_issuer_path_length = optional(number)
        }))

        # Certificate policy object identifiers (RFC 5280 section 4.2.1.4), each
        # an OID path such as [2, 23, 140, 1, 2, 1] (CA/Browser Forum
        # domain-validated).
        policy_ids = optional(list(object({
          # The OID's arcs, most significant first.
          object_id_path = list(number)
        })), [])

        # OCSP responder URLs placed in the Authority Information Access
        # extension, e.g. "http://ocsp.example.com". Google runs no OCSP
        # responder; list your own.
        aia_ocsp_servers = optional(list(string), [])

        # Custom X.509 extensions, each an OID, a base64 DER value, and whether
        # a relying party that does not understand it must reject the
        # certificate (critical).
        additional_extensions = optional(list(object({
          # The extension's OID.
          object_id = object({
            # The OID's arcs, most significant first.
            object_id_path = list(number)
          })

          # A relying party that does not understand a critical extension must
          # reject the certificate. Mark an extension critical only when every
          # consumer knows it.
          critical = optional(bool, false)

          # The extension's DER value, base64-encoded.
          value = string
        })), [])

        # The name constraints extension (RFC 5280 section 4.2.1.10): the names
        # certificates below a CA may and may not carry. Meaningful on CA
        # certificates.
        name_constraints = optional(object({
          # Whether the constraints are marked critical. RFC 5280 requires true on
          # a CA certificate that carries them.
          critical = optional(bool, false)

          # DNS names certificates below this CA may carry.
          permitted_dns_names = optional(list(string), [])

          # DNS names certificates below this CA must not carry.
          excluded_dns_names = optional(list(string), [])

          # IP ranges (CIDR) certificates below this CA may carry.
          permitted_ip_ranges = optional(list(string), [])

          # IP ranges (CIDR) certificates below this CA must not carry.
          excluded_ip_ranges = optional(list(string), [])

          # Email addresses, hosts, or ".domain" suffixes that may appear.
          permitted_email_addresses = optional(list(string), [])

          # Email addresses, hosts, or ".domain" suffixes that must not appear.
          excluded_email_addresses = optional(list(string), [])

          # URI hosts or ".domain" suffixes that may appear.
          permitted_uris = optional(list(string), [])

          # URI hosts or ".domain" suffixes that must not appear.
          excluded_uris = optional(list(string), [])
        }))
      }))
    }))

    # What the pool's authorities publish. Omit to publish nothing.
    publishing_options = optional(object({
      # Publish each authority's CA certificate and put its URL in the
      # Authority Information Access extension of issued certificates.
      publish_ca_cert = optional(bool, false)

      # Publish each authority's CRL (rebuilt daily and shortly after a
      # revocation; each expires after 7 days) and put its URL in the CRL
      # Distribution Points extension.
      publish_crl = optional(bool, false)

      # How published CA certificates and CRLs are encoded: "PEM" (the
      # default) or "DER".
      encoding_format = optional(string, "")
    }))

    # The Cloud KMS key that encrypts the subject, SANs, and PEM certificate
    # of stored certificates at rest -- a GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
    # in the pool's region. CA Service's service agent needs
    # cryptoKeyEncrypterDecrypter on it. Empty keeps Google's encryption.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Labels on the pool. The platform attribution labels are added on
    # top and win on a key conflict.
    labels = optional(map(string), {})

    # What happens to the pool when this resource is destroyed:
    #   "" / "DELETE" -- deleted (the provider's default); Google refuses while the
    #                    pool still holds an authority, even one in its 30-day
    #                    soft delete (destroy authorities with skip_grace_period
    #                    first)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- it leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
