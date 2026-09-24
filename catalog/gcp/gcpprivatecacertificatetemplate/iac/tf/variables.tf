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
  description = "GcpPrivateCaCertificateTemplate specification"
  type = object({
    # The GCP project the template lives in: a literal project ID or a GcpProject reference.
    # If omitted, the provider's default project is used. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The CA Service region, e.g. "us-central1"; certificates using the template must be in the same region. Immutable.
    location = string

    # The template's ID, unique in its parent: letters, digits, "-" and "_",
    # at most 63 characters. Defaults to metadata.name. Immutable.
    template_id = optional(string, "")

    # What the template is for, shown to people choosing one.
    description = optional(string, "")

    # The longest lifetime a certificate issued with the template may have,
    # e.g. 2592000s (30 days). The pool's own maximum still applies; the
    # shorter wins.
    maximum_lifetime = optional(string, "")

    # X.509 values stamped onto every certificate issued with the template.
    predefined_values = optional(object({
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

    # Limits on the identities certificates issued with it may carry. Omit
    # for no limits.
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
        expression = optional(string, "")

        # A short title for the expression.
        title = optional(string, "")

        # What the expression checks.
        description = optional(string, "")

        # Where the expression came from, for error messages (a file and line).
        location = optional(string, "")
      }))
    }))

    # Extensions a request may pass through the template.
    passthrough_extensions = optional(object({
      # Named extensions: BASE_KEY_USAGE, EXTENDED_KEY_USAGE, CA_OPTIONS,
      # POLICY_IDS, AIA_OCSP_SERVERS, NAME_CONSTRAINTS.
      known_extensions = optional(list(string), [])

      # Custom extensions, by OID.
      additional_extensions = optional(list(object({
        # The OID's arcs, most significant first.
        object_id_path = list(number)
      })), [])
    }))

    # Labels on the template. The platform attribution labels are added on
    # top and win on a key conflict.
    labels = optional(map(string), {})

    # What happens to the template when this resource is destroyed:
    #   "" / "DELETE" -- deleted (the provider's default)
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- it leaves management and stays in GCP
    deletion_policy = optional(string, "")
  })
}
