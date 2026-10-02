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
  description = "GcpManagedKafkaCluster specification"
  type = object({
    # The GCP project the cluster lives in: a literal project ID or a
    # GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The region the brokers run in, e.g. "us-central1". Immutable.
    location = string

    # The cluster's ID -- 1-63 characters, RFC 1035 (lowercase letters,
    # digits, hyphens; starts with a letter). Defaults to metadata.name.
    # Immutable.
    cluster_id = optional(string, "")

    # vCPUs and memory for the whole cluster.
    capacity_config = object({
      # vCPUs provisioned across the cluster -- at least 3. Billed per
      # vCPU-hour whether or not traffic flows. Sent as a decimal string.
      # Mutable in place (scaling up can trigger a rebalance, see
      # rebalance_mode).
      vcpu_count = optional(number, 0)

      # Memory provisioned across the cluster, in BYTES as a plain number
      # (3 GiB = 3221225472) -- between 1 GiB and 8 GiB per vCPU. Billed per
      # GiB-hour. Sent as a decimal string. Mutable in place.
      memory_bytes = optional(number, 0)
    })

    # Disk per broker in GiB -- at least 100. Empty leaves Google's default.
    # Sent as a decimal string.
    broker_disk_size_gib = optional(number, 0)

    # The VPC networks the cluster is reachable from -- at least one, at
    # most 10, one subnet per network.
    network_configs = list(object({
      # The subnet the cluster is reachable from: a GcpSubnetwork reference
      # (its self link, which the modules trim to the
      # projects/{project}/regions/{region}/subnetworks/{subnet} form Google
      # requires) or a literal in either form. It must be in the cluster's
      # region; the project may differ (a Shared VPC host). Only one subnet
      # per VPC network.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      subnet = string
    }))

    # A Cloud KMS key encrypting the cluster's data at rest (CMEK): a
    # GcpKmsKey reference or a literal
    # projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}.
    # The key must be in the cluster's region, and the Managed Kafka service
    # agent needs roles/cloudkms.cryptoKeyEncrypterDecrypter on it. Empty
    # uses Google-managed keys. Immutable.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key = optional(string, "")

    # Whether Google moves partitions onto new brokers when the cluster
    # scales up:
    #   NO_REBALANCE               -- never (Google's default when empty)
    #   AUTO_REBALANCE_ON_SCALE_UP -- rebalance automatically after a
    #                                 scale-up, spreading load onto the new
    #                                 brokers
    rebalance_mode = optional(string, "")

    # Mutual TLS for clients. See the message comment for how omitting,
    # declaring, and emptying it differ.
    tls_config = optional(object({
      # How the Distinguished Name of a client certificate becomes the short
      # principal name ACLs match (Kafka's ssl.principal.mapping.rules broker
      # setting, same syntax). Empty keeps Kafka's default behavior. Changing
      # it triggers a rolling restart of the brokers.
      ssl_principal_mapping_rules = optional(string, "")

      # Certificate Authority Service CA pools whose certificates the brokers
      # trust for client authentication -- GcpPrivateCaPool references (their
      # full names) or literals projects/{project}/locations/{location}/caPools/{pool},
      # in any project or location. At most 10. Setting at least one enables
      # mTLS.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      ca_pools = optional(list(string), [])
    }))

    # Labels on the cluster. The platform attribution labels are added on
    # top and win on a key conflict.
    labels = optional(map(string), {})

    # What happens to the cluster when this resource is destroyed:
    #   "" / "DELETE" -- deleted, with every topic and message in it
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- the cluster leaves management and keeps running
    #                    (and billing) in GCP
    deletion_policy = optional(string, "")
  })
}
