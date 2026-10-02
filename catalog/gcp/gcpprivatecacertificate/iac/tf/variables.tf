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
  description = "GcpPrivateCaCertificate specification"
  type = object({
    # The GCP project the certificate is issued in (its pool's project): a literal project ID or a GcpProject reference.
    # If omitted, the provider's default project is used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The CA Service region, e.g. "us-central1" -- its pool's region. Immutable.
    location = string

    # The CA pool that issues the certificate. A GcpPrivateCaPool reference resolves to its full
    # resource path (name output); a literal takes the full path
    # projects/{project}/locations/{location}/caPools/{id} or the bare pool
    # ID. The modules derive the bare ID Google's resource expects; the pool
    # must be in this resource's project and location. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    pool = string

    # The certificate's ID, unique in its parent: letters, digits, "-" and "_",
    # at most 63 characters. Defaults to metadata.name. Immutable.
    certificate_id = optional(string, "")

    # The authority in the pool that signs the certificate -- a
    # GcpPrivateCaCertificateAuthority reference (its full name) or a
    # literal full name or bare ID; the modules derive the ID. Empty lets
    # the pool choose an enabled authority. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    certificate_authority = optional(string, "")

    # The template to issue with -- a GcpPrivateCaCertificateTemplate
    # reference (its full name) or a literal
    # projects/*/locations/*/certificateTemplates/*, in the certificate's
    # location. The caller needs privateca.templateUser on it. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    certificate_template = optional(string, "")

    # How long the certificate is valid. Defaults to 315360000s (10 years);
    # cut short by the pool's or template's maximum and by the issuing
    # authority's own expiry. Immutable.
    lifetime = optional(string, "")

    # A PEM certificate signing request. Exactly one of pem_csr or config.
    # Immutable.
    pem_csr = optional(string, "")

    # The certificate as structured fields. Exactly one of pem_csr or
    # config. Immutable.
    config = optional(object({
      # The certificate's subject.
      subject_config = object({
        # The distinguished name. common_name is required.
        subject = object({
          # The common name (CN), e.g. "Example Root CA" or a host name.
          common_name = string

          # Two-letter country code (C), e.g. "US".
          country_code = optional(string, "")

          # Organization (O).
          organization = optional(string, "")

          # Organizational unit (OU).
          organizational_unit = optional(string, "")

          # Locality or city (L).
          locality = optional(string, "")

          # Province, territory, or state (ST).
          province = optional(string, "")

          # Street address.
          street_address = optional(string, "")

          # Postal code.
          postal_code = optional(string, "")
        })

        # Subject alternative names -- what TLS clients actually check against
        # the host they dialed. At least one entry when set.
        subject_alt_name = optional(object({
          # DNS names, e.g. "api.internal.example.com".
          dns_names = optional(list(string), [])

          # URIs, e.g. a SPIFFE ID "spiffe://example.org/ns/prod/sa/api".
          uris = optional(list(string), [])

          # Email addresses.
          email_addresses = optional(list(string), [])

          # IPv4 or IPv6 addresses.
          ip_addresses = optional(list(string), [])
        }))
      })

      # A custom Subject Key Identifier, lowercase hex. Omit to let Google
      # derive it.
      subject_key_id = optional(string, "")

      # The certificate's X.509 fields. A TLS server leaf sets
      # key_usage.base_key_usage.digital_signature and key_encipherment and
      # key_usage.extended_key_usage.server_auth.
      x509_config = object({
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
      })

      # The public key to certify.
      public_key = object({
        # The PEM public key (a SubjectPublicKeyInfo block), base64-encoded as a
        # whole -- Terraform's filebase64("key.pub.pem").
        key = string

        # The key's format. PEM is the only value Google accepts; empty sends
        # PEM.
        format = optional(string, "")
      })
    }))

    # Labels on the certificate. The platform attribution labels are added on
    # top and win on a key conflict.
    labels = optional(map(string), {})

    # What happens to the certificate when this resource is destroyed:
    #   "" / "DELETE" -- deleted (the provider's default) -- which revokes it
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- it leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
