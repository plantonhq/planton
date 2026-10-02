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
  description = "GcpBigtableInstance specification"
  type = object({
    # GCP project where the Bigtable instance will be created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Immutable: changing the project destroys and recreates the instance.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Bigtable instance (also called Instance ID in GCP Console).
    # This becomes the GCP resource name and is used by Bigtable client
    # libraries to connect. Must be 6-33 characters: lowercase letters,
    # numbers, and hyphens only. Must start with a lowercase letter and
    # end with a letter or number.
    # Immutable after creation.
    instance_name = string

    # Human-readable display name for the instance.
    # If not specified, defaults to the instance_name value.
    display_name = optional(string, "")

    # Whether deletion protection is enabled. When true, the instance
    # cannot be destroyed without first setting this to false.
    # Strongly recommended for production instances.
    # Default: true.
    deletion_protection = optional(bool)

    # Whether to delete all backups in the instance when destroying it.
    # Bigtable blocks instance deletion if backups exist unless this is
    # set to true. Only relevant during destroy operations.
    force_destroy = optional(bool, false)

    # One or more clusters that serve as physical replicas for this instance.
    # Each cluster must be in a different zone within the same or different
    # regions. At least one cluster is required. Up to 8 clusters can be
    # configured across cloud regions for multi-region replication.
    clusters = list(object({
      # Unique identifier for this cluster within the instance.
      # Must be 6-30 characters: lowercase letters, numbers, and hyphens only.
      # Must start with a lowercase letter and end with a letter or number.
      cluster_id = string

      # Zone where this cluster will be deployed (e.g., "us-central1-a").
      # Each cluster in the instance must be in a different zone.
      # Zones must be Bigtable-capable (see GCP Bigtable locations).
      # Immutable after creation.
      zone = string

      # Fixed number of nodes in this cluster. Mutually exclusive with
      # autoscaling_config. If neither is set, Bigtable auto-allocates
      # nodes based on the data footprint.
      num_nodes = optional(number, 0)

      # Storage type for this cluster.
      # SSD: lower latency, recommended for most workloads.
      # HDD: lower cost, suitable for large batch-analytics workloads
      # where latency is less critical.
      # Default: SSD.
      # Immutable after creation.
      storage_type = optional(string)

      # Cloud KMS encryption key to protect data in this cluster (CMEK).
      # Format: projects/{project}/locations/{location}/keyRings/{keyring}/cryptoKeys/{key}
      # The key region must match the cluster zone's region. All clusters
      # within an instance should use the same CMEK key.
      # Immutable after creation.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_name = optional(string, "")

      # Node scaling factor for this cluster. Controls the granularity of
      # node scaling: 1X scales in increments of 1 node, 2X scales in
      # increments of 2 nodes (for larger workloads that benefit from
      # coarser scaling steps). When using 2X, num_nodes, min_nodes, and
      # max_nodes must all be specified in increments of 2.
      # If not set, GCP defaults to NodeScalingFactor1X.
      # Immutable after creation.
      node_scaling_factor = optional(string, "")

      # Autoscaling configuration for this cluster. Mutually exclusive with
      # num_nodes. When set, Bigtable dynamically adjusts the number of
      # nodes based on CPU and storage utilization targets.
      autoscaling_config = optional(object({
        # Minimum number of nodes for autoscaling. Must be at least 1.
        min_nodes = number

        # Maximum number of nodes for autoscaling. Must be at least 1 and
        # greater than or equal to min_nodes.
        max_nodes = number

        # Target CPU utilization percentage for autoscaling. Bigtable adds nodes
        # when average CPU utilization exceeds this target and removes nodes
        # when utilization drops sufficiently below it. Must be between 10 and 80.
        cpu_target = number

        # Target storage utilization per node in GB. When total storage per node
        # exceeds this target, Bigtable adds nodes. The valid range depends on
        # the cluster's storage_type:
        #   - SSD: 2560 to 5120 (2.5 TiB to 5 TiB per node)
        #   - HDD: 8192 to 16384 (8 TiB to 16 TiB per node)
        # If not set, Bigtable uses the default for the storage type
        # (2560 for SSD, 8192 for HDD).
        storage_target = optional(number, 0)
      }))
    }))

    # User-defined labels to organize and track the instance (instance-level
    # only — GCP has no per-cluster labels). Merged beneath Planton's
    # platform attribution labels (platform keys win on conflict).
    labels = optional(map(string), {})

    # Edition of the instance, gating feature availability. ENTERPRISE
    # (the GCP default when unset) is the standard production edition;
    # ENTERPRISE_PLUS adds enterprise capabilities such as multi-location
    # automated-backup placement for the instance's tables. Upgrading
    # ENTERPRISE -> ENTERPRISE_PLUS applies in place; there is no
    # downgrade path.
    edition = optional(string, "")

    # Resource Manager tags bound to the instance for org-policy and IAM
    # conditions. Keys in the form "tagKeys/{id}" (or the namespaced
    # "project/tag-key"), values "tagValues/{id}" (or
    # "project/tag-key/tag-value"). Create-time only: changing them later
    # replaces the instance — plan tag changes deliberately.
    resource_manager_tags = optional(map(string), {})

    # Deletion policy for the instance — what happens when this resource
    # is destroyed. Applies only once deletion_protection (above) permits
    # a destroy at all:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the instance is deleted with every cluster and table
    #                on it (backups additionally gated by force_destroy)
    #   "PREVENT" -- destroy FAILS; a second wall for the instance a
    #                data platform depends on
    #   "ABANDON" -- the instance is removed from management but left
    #                running (and billing) in GCP with its data intact
    deletion_policy = optional(string, "")
  })
}
