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
  description = "GcpKmsKey specification"
  type = object({
    # The key ring that this key belongs to.
    # Accepts the fully qualified key ring path
    #   projects/{project}/locations/{location}/keyRings/{name}
    # or a reference to a GcpKmsKeyRing resource. The key inherits the ring's
    # project and location. Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    key_ring_id = string

    # Name of the key in GCP. Immutable after creation.
    # Must be 1-63 characters: letters (upper or lower), digits, hyphens,
    # or underscores. This is the GCP resource name, distinct from the
    # Planton metadata.name. Because keys are permanent (see the message
    # comment), a name can never be reused within its key ring — pick
    # versioned names (e.g. "cmek-data-v2") if a key may ever be replaced.
    # Example: "cmek-encrypt-key", "artifact-signing-key"
    key_name = string

    # The immutable purpose of this key. Determines what cryptographic
    # operations the key supports and which algorithms its version template
    # may use. Cannot be changed after creation. If not set, GCP defaults
    # to ENCRYPT_DECRYPT — the purpose every CMEK integration expects.
    #
    # Known values:
    #   "ENCRYPT_DECRYPT"      -- symmetric encryption for CMEK (default)
    #   "ASYMMETRIC_SIGN"      -- digital signatures
    #   "ASYMMETRIC_DECRYPT"   -- asymmetric decryption
    #   "RAW_ENCRYPT_DECRYPT"  -- raw AES-GCM/CTR for interoperable encryption
    #   "MAC"                  -- keyed message authentication codes
    #
    # GCP adds purposes over time (for example key encapsulation for
    # post-quantum schemes), so this field deliberately accepts any string
    # and lets the API validate — see the CryptoKeyPurpose reference for
    # the authoritative list:
    # https://cloud.google.com/kms/docs/reference/rest/v1/projects.locations.keyRings.cryptoKeys#CryptoKeyPurpose
    purpose = optional(string, "")

    # How often to auto-generate a new CryptoKeyVersion and set it as primary.
    # Format: decimal seconds with suffix "s" (e.g., "7776000s" for 90 days).
    # Must be at least 86400s (24 hours) with at most 9 fractional digits.
    # Only allowed for ENCRYPT_DECRYPT keys (enforced pre-deploy); other
    # purposes require manual key version management. Mutable: shortening or
    # lengthening the period applies in place and reschedules the next
    # rotation. Old versions remain usable for decryption until destroyed —
    # rotation limits blast radius going forward; it does not re-encrypt
    # existing data.
    rotation_period = optional(string, "")

    # How long destroyed CryptoKeyVersions spend in DESTROY_SCHEDULED state
    # before being permanently destroyed — the recovery window during which
    # a destruction can still be undone. Immutable after creation.
    # Format: decimal seconds with suffix "s" (e.g., "2592000s" for 30 days).
    # Defaults to 30 days when not specified. Shorter windows reduce the
    # time compromised material lingers; longer windows protect against
    # accidental destruction of data-encrypting keys.
    destroy_scheduled_duration = optional(string, "")

    # Template describing settings for new CryptoKeyVersions.
    # Use this to specify the encryption algorithm and protection level.
    # If omitted, GCP defaults to GOOGLE_SYMMETRIC_ENCRYPTION with SOFTWARE
    # protection -- which is correct for standard CMEK use cases.
    version_template = optional(object({
      # The algorithm to use when creating a CryptoKeyVersion based on this
      # template. Required when version_template is specified. The algorithm
      # must be compatible with the key's purpose; GCP rejects mismatches at
      # create time. Mutable: changing the algorithm affects only versions
      # created afterward — existing versions keep the algorithm they were
      # generated with.
      #
      # Common values by purpose:
      #   ENCRYPT_DECRYPT:     "GOOGLE_SYMMETRIC_ENCRYPTION" (default if omitted)
      #   ASYMMETRIC_SIGN:     "EC_SIGN_P256_SHA256", "RSA_SIGN_PSS_2048_SHA256", ...
      #   ASYMMETRIC_DECRYPT:  "RSA_DECRYPT_OAEP_2048_SHA256", ...
      #   RAW_ENCRYPT_DECRYPT: "AES_128_GCM", "AES_256_GCM", ...
      #   MAC:                 "HMAC_SHA256"
      #
      # GCP adds algorithms over time (for example post-quantum signature
      # schemes), so this field deliberately accepts any string and lets the
      # API validate — see the CryptoKeyVersionAlgorithm reference for the
      # authoritative list:
      # https://cloud.google.com/kms/docs/reference/rest/v1/CryptoKeyVersionAlgorithm
      algorithm = string

      # The protection level for CryptoKeyVersions created with this template.
      # Immutable after creation.
      #
      # Valid values:
      #   "SOFTWARE" (default) -- keys protected in software
      #   "HSM"                -- keys protected by Cloud HSM (FIPS 140-2 Level 3)
      #   "EXTERNAL"           -- key material held in an external key manager,
      #                           linked per version via an external key URI
      #   "EXTERNAL_VPC"       -- key material held in an external key manager
      #                           reached over a VPC; requires the key-level
      #                           crypto_key_backend to name the EKM connection
      protection_level = optional(string, "")
    }))

    # If true, the key is created without an initial CryptoKeyVersion.
    # You must create versions manually afterward (or import material via a
    # key ring import job). Consumed only at create time. Required (and
    # enforced pre-deploy) when import_only is true.
    skip_initial_version_creation = optional(bool, false)

    # If true, this key may contain only imported key versions — GCP will
    # never generate material for it, making the key a bring-your-own-key
    # (BYOK) container whose material provenance is externally controlled.
    # Immutable after creation. Requires skip_initial_version_creation
    # (enforced pre-deploy), since an auto-generated initial version would
    # violate the import-only guarantee.
    import_only = optional(bool, false)

    # The EKM connection through which an external key manager backs this
    # key's versions. Applies only when version_template.protection_level is
    # EXTERNAL_VPC (enforced pre-deploy). Accepts the fully qualified
    # connection path
    #   projects/{project}/locations/{location}/ekmConnections/{name}
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    crypto_key_backend = optional(string, "")

    # User-defined labels attached to the key, for cost attribution and
    # fleet queries. Merged with Planton's platform labels (which win on
    # key conflicts). Mutable in place.
    labels = optional(map(string), {})

    # What destroying this resource does to the key. KMS keys can NEVER be
    # deleted from GCP — only their versions can — so the choice here is
    # about the versions, and it is the most consequential destroy switch
    # in this spec:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- EVERY key version is scheduled for destruction (after
    #                destroy_scheduled_duration) and rotation is disabled;
    #                data encrypted under this key becomes UNRECOVERABLE
    #                once the versions are destroyed. The key shell remains
    #                in the project.
    #   "PREVENT" -- destroy FAILS; the right default posture for any key
    #                protecting data you cannot afford to lose
    #   "ABANDON" -- the key leaves management with every version intact
    #                and enabled — encrypted data stays decryptable
    deletion_policy = optional(string, "")
  })
}
