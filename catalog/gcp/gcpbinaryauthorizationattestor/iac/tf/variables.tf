variable "metadata" {
  description = "Catalog object metadata"
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
  description = "GcpBinaryAuthorizationAttestor specification"
  type = object({
    # The project the attestor (and a created note) lives in: a literal
    # project ID or a GcpProject reference. Empty means the provider's
    # default project. The module enables binaryauthorization.googleapis.com
    # and containeranalysis.googleapis.com there.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The attestor's ID, unique in its project, e.g. "built-by-ci" or
    # "qa-approved". Defaults to metadata.name. Immutable: a new ID replaces
    # the attestor, and policies naming the old one must be re-applied.
    attestor_name = optional(string, "")

    # What the attestor vouches for; shown when choosing attestors.
    description = optional(string, "")

    # The note to create for this attestor. Mutually exclusive with
    # attestation_authority_note.note_reference.
    note = optional(object({
      # The note's ID in the project. Defaults to "{attestor_name}-note".
      # Immutable.
      note_name = optional(string, "")

      # The signer's name as people read it, e.g. "Production build pipeline"
      # (Google's attestation_authority.hint). Required.
      human_readable_name = string

      # A one-sentence description of the note.
      short_description = optional(string, "")

      # A longer description: what signing by this attestor means.
      long_description = optional(string, "")

      # When the note expires, as an RFC 3339 timestamp, e.g.
      # "2027-01-01T00:00:00Z". Empty: never.
      expiration_time = optional(string, "")

      # Other notes this one relates to: projects/{project}/notes/{note}.
      related_note_names = optional(list(string), [])

      # Links to more information, e.g. the signing pipeline's runbook.
      related_url = optional(list(object({
        # The URL. Required.
        url = string

        # The link's label.
        label = optional(string, "")
      })), [])
    }))

    # The attestor's note reference and public keys.
    attestation_authority_note = optional(object({
      # An existing ATTESTATION_AUTHORITY note: projects/{project}/notes/{note},
      # or a bare note ID in the attestor's project. Leave empty when note
      # creates one. Immutable: a new reference replaces the attestor.
      note_reference = optional(string, "")

      # The public keys that verify attestations; one verifying key is enough.
      # Rotate by adding the new key, re-signing, then removing the old one.
      # Empty: the attestor never verifies anything, so every rule requiring
      # it denies.
      public_keys = optional(list(object({
        # The key's ID, which every signature must name exactly. For a PKIX key
        # it must be an RFC 3986 URI; empty takes Google's default (a digest of
        # the key), or for a Cloud KMS key the module sets
        # //cloudkms.googleapis.com/v1/{key version}, the ID gcloud and Google's
        # signing tools use. Leave empty for a PGP key.
        id = optional(string, "")

        # A note about the key, e.g. who holds the private half.
        comment = optional(string, "")

        # An ASCII-armored PGP public key, the whole output of
        # `gpg --export --armor signer@example.com`.
        ascii_armored_pgp_public_key = optional(string, "")

        # A PKIX public key: a PEM, or a Cloud KMS signing key version.
        pkix_public_key = optional(object({
          # A PEM-encoded public key (RFC 7468 SubjectPublicKeyInfo), for a key
          # pair held outside Cloud KMS.
          public_key_pem = optional(string, "")

          # The algorithm signatures use with public_key_pem; it must match the
          # key. Google's names: EC_SIGN_P256_SHA256, EC_SIGN_P384_SHA384,
          # EC_SIGN_P521_SHA512 (the ECDSA_* names are the same algorithms),
          # RSA_SIGN_PKCS1_{2048,3072,4096}_SHA256, RSA_SIGN_PKCS1_4096_SHA512,
          # RSA_SIGN_PSS_{2048,3072,4096}_SHA256, RSA_SIGN_PSS_4096_SHA512 (and
          # their RSA_PSS_* aliases), and the post-quantum ML_DSA_65.
          signature_algorithm = optional(string, "")

          # A Cloud KMS asymmetric-sign key version whose public key verifies the
          # signatures: a GcpKmsKey reference (the version Google creates with the
          # key) or a literal
          # projects/*/locations/*/keyRings/*/cryptoKeys/*/cryptoKeyVersions/*.
          # Both modules read the version's public key and algorithm from Cloud
          # KMS, so the pipeline signs with the private half that never leaves
          # KMS. The key's purpose must be ASYMMETRIC_SIGN.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          kms_key_version = optional(string, "")
        }))
      })), [])
    }))

    # What destroying this block does, for the attestor and a created note:
    #   "" / "DELETE" -- both are deleted; policies naming the attestor stop
    #                    admitting images that only it signed
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- both leave management and stay in Google
    deletion_policy = optional(string, "")
  })
}
