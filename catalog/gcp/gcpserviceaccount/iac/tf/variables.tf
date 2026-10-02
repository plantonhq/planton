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
  description = "GcpServiceAccount specification"
  type = object({
    # The short account ID that forms the service account email:
    # <service_account_id>@<project>.iam.gserviceaccount.com.
    # Must be 6-30 characters, start with a lowercase letter, and contain only
    # lowercase letters, digits, and hyphens (cannot end with a hyphen).
    # Immutable: changing it destroys and recreates the service account, which
    # invalidates every IAM binding and workload identity that references the old email.
    service_account_id = string

    # The GCP project in which the service account is created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used — set it explicitly in
    # multi-project layouts so the identity's home project is unambiguous.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Human-readable display name shown in the GCP console (max 100 characters
    # in GCP; not validated here because the API truncates rather than rejects).
    # If omitted, the resource's metadata name is used.
    # Mutable: can be changed without recreating the service account.
    display_name = optional(string, "")

    # Human-readable description of what this service account is for (max 256 bytes).
    # Surfaces in the GCP console and `gcloud iam service-accounts describe` — write it
    # for the operator who finds this identity two years from now.
    # Mutable: can be changed without recreating the service account.
    description = optional(string, "")

    # Whether the service account is disabled. A disabled service account keeps its
    # IAM bindings but cannot authenticate — tokens are rejected until it is re-enabled.
    # Useful as a kill switch during incident response or for staged decommissioning
    # (disable first, observe breakage, then delete).
    disabled = optional(bool)

    # Create a user-managed key for this service account. Omit for keyless
    # (the recommended default — prefer Workload Identity, impersonation, or
    # federation wherever the workload supports it). When present, a key is
    # created with the configured algorithm and formats, and the private key
    # (unless public_key_data supplies your own public key) is exported in
    # stack outputs as `key_base64` — treat that output as a live credential.
    user_managed_key = optional(object({
      # Algorithm used to generate the key:
      #   ""                  -- GCP default ("KEY_ALG_RSA_2048")
      #   "KEY_ALG_RSA_2048"  -- 2048-bit RSA (the standard choice)
      #   "KEY_ALG_RSA_1024"  -- 1024-bit RSA (legacy; weaker — avoid for
      #                          new keys)
      # Create-time only: changing it replaces the key.
      algorithm = optional(string, "")

      # Output format of the generated private key:
      #   ""                             -- GCP default
      #                                     ("TYPE_GOOGLE_CREDENTIALS_FILE")
      #   "TYPE_GOOGLE_CREDENTIALS_FILE" -- the standard JSON credentials file
      #   "TYPE_PKCS12_FILE"             -- PKCS#12 bundle (password "notasecret";
      #                                     for legacy tooling that requires p12)
      private_key_type = optional(string, "")

      # Output format of the public key:
      #   ""                    -- GCP default ("TYPE_X509_PEM_FILE")
      #   "TYPE_X509_PEM_FILE"  -- X.509 certificate PEM
      #   "TYPE_RAW_PUBLIC_KEY" -- raw public key bytes
      #   "TYPE_NONE"           -- do not return the public key
      public_key_type = optional(string, "")

      # Your own public key (base64-encoded X.509 PEM) — the UPLOAD flow: the
      # matching private key never leaves your custody and GCP returns no
      # private key material (the key_base64 stack output stays empty).
      # The strongest key posture when a user-managed key is unavoidable.
      public_key_data = optional(string, "")

      # Arbitrary key/value pairs whose CHANGE forces a new key to be
      # generated — the idiomatic rotation trigger. Set e.g.
      # {"rotation": "2026-08"} and bump the value on your rotation cadence;
      # the old key is destroyed and a fresh one exported.
      keepers = optional(map(string), {})

      # Deletion policy for the key itself:
      #   ""        -- same as "DELETE" (provider default)
      #   "DELETE"  -- the key is deleted on destroy
      #   "PREVENT" -- destroy FAILS while this key exists
      deletion_policy = optional(string, "")

      # Whether the key exists. Unset means yes: declaring the block has always
      # meant creating a key, and this switch lets a manifest say the opposite
      # out loud (the key is destroyed when it flips to false).
      enabled = optional(bool)
    }))

    # IAM roles granted to this service account at the PROJECT scope, e.g.
    # ["roles/logging.logWriter", "roles/storage.admin"]. Grants are additive
    # (member-level): they never clobber other members' bindings on the same role.
    # For grants that should be first-class, referenceable nodes in the resource
    # graph (visible dependencies, independent lifecycle), use GcpProjectIamMember
    # instead of this list.
    project_iam_roles = optional(list(string), [])

    # The numeric organization ID (e.g. "123456789012") — required only when
    # org_iam_roles is set, to identify which organization receives the grants.
    org_id = optional(string, "")

    # IAM roles granted to this service account at the ORGANIZATION scope, e.g.
    # ["roles/resourcemanager.organizationViewer"]. Requires org_id. Grants are
    # additive (member-level). Org-scope grants affect every project under the
    # organization — grant sparingly.
    org_iam_roles = optional(list(string), [])

    # If true, creating the service account succeeds (as a no-op adoption)
    # when an account with the same email already exists, instead of
    # failing. Useful for idempotent bootstrap flows that may race other
    # provisioning paths onto well-known identity names.
    create_ignore_already_exists = optional(bool, false)

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the service account is deleted
    #   "PREVENT" -- destroy FAILS; a guard rail for identities whose
    #                deletion would invalidate IAM bindings fleet-wide
    # Mutable in place.
    deletion_policy = optional(string, "")
  })
}
