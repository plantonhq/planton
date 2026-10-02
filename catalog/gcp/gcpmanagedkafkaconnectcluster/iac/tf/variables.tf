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
  description = "GcpManagedKafkaConnectCluster specification"
  type = object({
    # The GCP project the Connect cluster lives in: a literal project ID or
    # a GcpProject reference. If omitted, the provider's default project is
    # used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The region the workers run in, e.g. "us-central1". Immutable.
    location = string

    # The Connect cluster's ID. Defaults to metadata.name. Immutable.
    connect_cluster_id = optional(string, "")

    # The Kafka cluster the workers attach to (their internal topics and the
    # connectors' default cluster): a GcpManagedKafkaCluster reference (its
    # name output) or a literal full path
    # projects/{project}/locations/{location}/clusters/{cluster}. Mutable in
    # place.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kafka_cluster = string

    # vCPUs and memory for the workers.
    capacity_config = object({
      # vCPUs for the Connect workers -- at least 3. Billed per vCPU-hour.
      # Sent as a decimal string.
      vcpu_count = optional(number, 0)

      # Memory for the Connect workers, in BYTES as a plain number -- at least
      # 3 GiB (3221225472), at a vCPU:GiB ratio between 1:1 and 1:8. Billed per
      # GiB-hour. Sent as a decimal string.
      memory_bytes = optional(number, 0)
    })

    # The VPC networks the workers reach -- at least one, at most 10.
    network_configs = list(object({
      # The subnet the workers' interface is placed in: a GcpSubnetwork
      # reference (its self link, which the modules trim to the
      # projects/{project}/regions/{region}/subnetworks/{subnet} form Google
      # requires) or a literal in either form. It must be in the Connect
      # cluster's region, and its CIDR range must be private (RFC 1918).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      primary_subnet = string

      # Extra DNS domains from this network the workers may resolve. For
      # MirrorMaker 2, add the target Kafka cluster's DNS domain -- the
      # bootstrap address without its leading "bootstrap." label and its port
      # (e.g. my-cluster.us-central1.managedkafka.my-project.cloud.goog).
      dns_domain_names = optional(list(string), [])
    }))

    # Labels on the Connect cluster. The platform attribution labels are
    # added on top and win on a key conflict.
    labels = optional(map(string), {})

    # What happens to the Connect cluster when this resource is destroyed:
    #   "" / "DELETE" -- deleted, with its connectors
    #   "PREVENT"     -- destroy fails
    #   "ABANDON"     -- it leaves management and keeps running (and
    #                    billing) in GCP
    deletion_policy = optional(string, "")
  })
}
