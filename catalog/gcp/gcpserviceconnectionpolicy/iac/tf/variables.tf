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
  description = "GcpServiceConnectionPolicy specification"
  type = object({
    # GCP project that owns the network and the policy. Can be a literal
    # project ID or a reference to a GcpProject resource. If omitted, the
    # provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name for the policy resource. If not specified, defaults to
    # metadata.name. Must start with a lowercase letter, contain only
    # lowercase letters, numbers, and hyphens, and end with a letter or
    # number. Immutable after creation.
    policy_name = optional(string, "")

    # GCP region the policy applies to (e.g. us-central1). Service
    # connectivity automation is regional: a producer instance in another
    # region needs its own policy there. Immutable after creation.
    location = string

    # The consumer VPC network this policy authorizes connections into.
    # A reference resolves to the GcpVpcNetwork's resource path
    # (projects/{project}/global/networks/{name}) — the format the Service
    # Connectivity API requires; both engines also normalize full self-link
    # URLs to that path. Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    network = string

    # The service class this policy authorizes — the producer's published
    # identifier. Google services use a gcp- prefix (gcp-memorystore for
    # Memorystore for Valkey, gcp-memorystore-redis for Redis Cluster);
    # third-party producers publish their own class names. One policy
    # exists per (network, service class, region). Immutable after
    # creation.
    service_class = string

    # Free-text description of what this policy is for — which workloads
    # or teams the authorized service class serves.
    description = optional(string, "")

    # User-defined labels to organize and track the policy. Merged beneath
    # Planton's platform attribution labels (platform keys win on
    # conflict).
    labels = optional(map(string), {})

    # Private Service Connect configuration: the subnets endpoint IPs come
    # from, the connection limit, and the optional producer-location
    # allowlist. Required in practice — the policy authorizes nothing
    # usable without address space — but optional in the API, so presets
    # always set it.
    psc_config = optional(object({
      # Subnets the connectivity automation allocates PSC endpoint IPs from.
      # References resolve to the GcpSubnetwork's self-link; both engines
      # normalize self-link URLs to the relative resource path
      # (projects/{project}/regions/{region}/subnetworks/{name}) the Service
      # Connectivity API expects. The subnets must live in the same region as
      # the policy and inside the policy's network. At least one is required.
      # Regular-purpose subnets work — no special PSC purpose is needed for
      # service connection policies (unlike PSC NAT subnets for published
      # services).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnetworks = list(string)

      # Maximum number of PSC connections the automation may create under this
      # policy. 0 (unset) leaves the limit to GCP's default. Useful as a
      # guardrail in shared networks: it caps how many managed-service
      # instances can attach through this policy before an operator has to
      # deliberately raise it.
      limit = optional(number, 0)

      # Authorization mechanism deciding which producer projects the
      # automation will connect to. Leave empty for GCP's default behavior
      # (any producer instance the consumer project provisions).
      # CUSTOM_RESOURCE_HIERARCHY_LEVELS restricts producers to the resource
      # hierarchy entries listed in
      # allowed_google_producers_resource_hierarchy_levels.
      producer_instance_location = optional(string, "")

      # Projects, folders, or organizations producer instances may live in,
      # each as 'projects/{id-or-number}', 'folders/{number}', or
      # 'organizations/{number}' (e.g. projects/my-project-id, folders/891,
      # organizations/123). Only consulted when producer_instance_location is
      # CUSTOM_RESOURCE_HIERARCHY_LEVELS.
      allowed_google_producers_resource_hierarchy_levels = optional(list(string), [])
    }))

    # Deletion policy for the policy — what happens when this resource is
    # destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the policy is deleted (existing PSC endpoints are
    #                stranded and new instances of the service class can no
    #                longer connect in this region)
    #   "PREVENT" -- destroy FAILS; protects the connectivity every managed
    #                instance under this policy depends on
    #   "ABANDON" -- the policy is removed from management but keeps
    #                authorizing connections in GCP
    deletion_policy = optional(string, "")
  })
}
