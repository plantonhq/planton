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
  description = "GcpKmsKeyRing specification"
  type = object({
    # The GCP project in which to create this key ring.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Example: "my-prod-project-123"
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the key ring in GCP. Immutable after creation.
    # Must be 1-63 characters: letters (upper or lower), digits, hyphens, or underscores.
    # This is the GCP resource name, distinct from the Planton metadata.name.
    # Because key rings are permanent, a name can never be reused within its
    # project and location — pick names that will not need recycling.
    # Example: "prod-encryption", "data-keys-us-central1"
    key_ring_name = string

    # GCP location (region, multi-region, or "global") where the key ring
    # resides. Immutable after creation. Keys must live in the same location
    # as the resources they protect for most CMEK integrations, so choose
    # based on where the encrypted data lives — plus data-residency and
    # redundancy requirements.
    #
    # Common values:
    #   Region:       "us-central1", "europe-west1", "asia-east1"
    #   Multi-region: "us", "europe", "asia"
    #   Global:       "global"
    #
    # Run `gcloud kms locations list` for a full list of valid locations.
    location = string
  })
}
