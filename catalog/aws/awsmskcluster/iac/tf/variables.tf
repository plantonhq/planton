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
  description = "AwsMskCluster specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # kafka_version is the Apache Kafka version for all brokers in the cluster.
    # Examples: "3.6.0", "3.5.1", "3.4.0", "2.8.1".
    # Upgrades are applied in place via rolling restart (an UpdateClusterKafkaVersion
    # operation); a version DOWNGRADE cannot be performed in place and forces cluster
    # replacement.
    kafka_version = string

    # number_of_broker_nodes is the total number of Kafka broker nodes in the cluster.
    # Must be a multiple of the number of subnets provided in subnet_ids so that
    # brokers are evenly distributed across Availability Zones. Broker count can be
    # INCREASED in place (an UpdateBrokerCount operation); AWS does not support
    # decreasing the broker count.
    number_of_broker_nodes = number

    # instance_type determines the compute and memory capacity of each broker node.
    # Standard types: kafka.m5.large, kafka.m5.xlarge, kafka.m5.2xlarge, kafka.m5.4xlarge.
    # Graviton types: kafka.m7g.large, kafka.m7g.xlarge (better price-performance).
    # Express brokers: express.m7g.large and up (AWS-managed storage, faster scaling,
    # intelligent rebalancing -- see rebalancing_status).
    # Small/dev types: kafka.t3.small (no tiered storage, no public access, no
    # provisioned throughput).
    # Updatable in place: changing the instance type is a rolling UpdateBrokerType
    # operation, not a replacement.
    instance_type = string

    # subnet_ids are the VPC subnets where broker nodes are placed.
    # Brokers are distributed round-robin across subnets. The number of broker nodes
    # must be a multiple of the number of subnets for even AZ distribution.
    # ForceNew: changing subnets forces cluster replacement.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    subnet_ids = list(string)

    # security_group_ids are the security groups ATTACHED to the broker network
    # interfaces -- they define what can reach the brokers. Ingress rules for the
    # Kafka listener ports (9092 plaintext, 9094 TLS, 9096 SASL/SCRAM, 9098
    # SASL/IAM) and ZooKeeper (2181-2182) belong on these referenced
    # AwsSecurityGroup nodes.
    # ForceNew: adding or removing entries after creation forces cluster replacement.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    security_group_ids = list(string)

    # public_access_type controls whether the cluster is reachable from the public internet.
    # "DISABLED" (default): brokers are only reachable within the VPC.
    # "SERVICE_PROVIDED_EIPS": AWS assigns public IPs to brokers.
    # AWS only allows turning public access ON for an EXISTING cluster: creation always
    # starts DISABLED, and the provider applies SERVICE_PROVIDED_EIPS as a follow-up
    # connectivity update. Public access also requires real client authentication
    # (SASL/IAM, SASL/SCRAM, or mTLS -- unauthenticated must be off) and TLS-only
    # client_broker_encryption.
    public_access_type = optional(string, "")

    # vpc_connectivity enables multi-VPC private connectivity (AWS PrivateLink) so
    # clients in OTHER VPCs or accounts connect to the brokers without peering or
    # public exposure. Enable at least one authentication scheme here; each scheme
    # requires the same scheme to be enabled in `authentication` (an IAM-only
    # cluster cannot offer SCRAM over PrivateLink). AWS activates PrivateLink
    # connectivity as a follow-up update after the cluster is created; the consumer
    # side (an aws_msk_vpc_connection in the client VPC) is a separate surface.
    vpc_connectivity = optional(object({
      # sasl_iam_enabled offers SASL/IAM authentication to PrivateLink clients.
      sasl_iam_enabled = optional(bool, false)

      # sasl_scram_enabled offers SASL/SCRAM authentication to PrivateLink clients.
      sasl_scram_enabled = optional(bool, false)

      # tls_enabled offers mutual TLS authentication to PrivateLink clients.
      tls_enabled = optional(bool, false)
    }))

    # network_type selects the IP addressing of the broker network interfaces.
    # "IPV4" (default): IPv4-only.
    # "DUAL": dual-stack IPv4 + IPv6 (requires dual-stack subnets).
    # One-way: AWS only supports updating IPV4 -> DUAL; going back forces replacement.
    network_type = optional(string, "")

    # ebs_volume_size_gib is the size of the EBS volume per broker, in GiB.
    # Range: 1-16384. If omitted, AWS uses the instance-type-specific default.
    # Volume size can be increased in place (an UpdateBrokerStorage operation).
    # Not applicable to express.* instance types (AWS manages Express broker storage).
    ebs_volume_size_gib = optional(number)

    # provisioned_throughput_enabled enables provisioned EBS throughput for higher streaming performance.
    # Only supported on kafka.m5.4xlarge and larger instance types with ebs_volume_size_gib >= 10 GiB.
    provisioned_throughput_enabled = optional(bool, false)

    # provisioned_throughput_mbs is the provisioned EBS throughput in MiB/s per broker.
    # Range: 250-2375. Required when provisioned_throughput_enabled is true.
    provisioned_throughput_mbs = optional(number, 0)

    # storage_mode controls the data storage strategy.
    # "LOCAL" (default): all data on broker EBS volumes.
    # "TIERED": hot data on EBS, warm data automatically offloaded to low-cost storage.
    # Tiered storage requires Kafka 2.8.2.tiered+ and supported instance types
    # (not kafka.t3.small). One-way: moving TIERED -> LOCAL forces cluster replacement.
    storage_mode = optional(string, "")

    # kms_key_arn is the KMS key ARN for encrypting data at rest on broker EBS volumes.
    # If omitted, AWS uses the default aws/msk service key.
    # ForceNew: changing the KMS key forces cluster replacement.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_arn = optional(string, "")

    # client_broker_encryption controls encryption for data in transit between clients and brokers.
    # "TLS" (default, recommended): all client-broker traffic is TLS-encrypted (port 9094).
    # "TLS_PLAINTEXT": both TLS (9094) and plaintext (9092) are available.
    # "PLAINTEXT": all client-broker traffic is unencrypted (port 9092).
    client_broker_encryption = optional(string)

    # in_cluster_encryption enables TLS encryption for data in transit between brokers.
    # Strongly recommended for production. ForceNew: changing this forces cluster replacement.
    in_cluster_encryption = optional(bool)

    # authentication configures client authentication methods for the cluster.
    # Multiple methods can be enabled simultaneously (e.g., SASL/IAM + TLS).
    # If no authentication is configured, the cluster accepts unauthenticated connections.
    authentication = optional(object({
      # sasl_iam_enabled enables SASL/IAM authentication.
      # Recommended for most workloads. Clients authenticate using AWS IAM credentials
      # (access key, role assumption, or instance profiles). No password management required.
      # Brokers listen on port 9098 for SASL/IAM connections.
      sasl_iam_enabled = optional(bool, false)

      # sasl_scram_enabled enables SASL/SCRAM-SHA-512 authentication.
      # Clients authenticate with username/password stored in AWS Secrets Manager.
      # Useful for non-AWS clients that cannot use IAM. Brokers listen on port 9096.
      # Associate the credential secrets via scram_secret_arns on the spec.
      sasl_scram_enabled = optional(bool, false)

      # tls_enabled enables mutual TLS (mTLS) authentication.
      # Clients present X.509 certificates signed by a private Certificate Authority.
      # Requires tls_certificate_authority_arns to specify trusted CAs.
      # Brokers listen on port 9094 for TLS connections.
      tls_enabled = optional(bool, false)

      # tls_certificate_authority_arns are the ARNs of ACM Private Certificate Authority resources
      # trusted for client certificate validation. Required when tls_enabled is true.
      # Reference AwsPrivateCa certificate_authority_arn outputs or pass literal ARNs.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      tls_certificate_authority_arns = optional(list(string), [])

      # unauthenticated allows clients to connect without any authentication.
      # Not recommended for production. Useful for development or when network-level
      # security (VPC + security groups) provides sufficient access control.
      unauthenticated = optional(bool, false)
    }))

    # scram_secret_arns are AWS Secrets Manager secret ARNs associated with the
    # cluster for SASL/SCRAM username/password authentication. Each secret holds a
    # {"username": ..., "password": ...} JSON document, MUST be named with the
    # AmazonMSK_ prefix, and MUST be encrypted with a customer-managed KMS key
    # (AWS rejects secrets on the default aws/secretsmanager key). Associations are
    # a cluster-keyed setting (added/removed in place); the secrets themselves are
    # managed outside this resource.
    scram_secret_arns = optional(list(string), [])

    # cluster_policy is a resource-based IAM policy attached to the cluster, as a
    # structured policy document -- the mechanism behind cross-account PrivateLink
    # access (granting kafka:CreateVpcConnection and the Get*ForVpcConnection
    # actions to consumer principals). A cluster setting keyed by the cluster ARN,
    # updated in place. The IaC modules serialize this document to JSON for the
    # provider.
    cluster_policy = optional(any)

    # configuration_arn is the ARN of an externally managed MSK Configuration resource.
    # MSK Configurations hold Apache Kafka server.properties overrides (e.g., replication factor,
    # min ISR, log retention). Mutually exclusive with server_properties.
    configuration_arn = optional(string, "")

    # configuration_revision is the revision number of the external MSK Configuration.
    # Required when configuration_arn is set. Must be >= 1.
    configuration_revision = optional(number, 0)

    # server_properties defines Apache Kafka server.properties overrides as key-value pairs.
    # When provided, a module-managed MSK Configuration resource is created and associated
    # with the cluster (the folded shape; bring an existing one via configuration_arn).
    # Common properties: auto.create.topics.enable, default.replication.factor,
    # min.insync.replicas, num.partitions, log.retention.hours, log.retention.bytes.
    # Mutually exclusive with configuration_arn.
    server_properties = optional(map(string), {})

    # logging configures broker log delivery to one or more destinations.
    # All three destinations (CloudWatch Logs, Kinesis Data Firehose, S3) can be
    # enabled simultaneously for different operational workflows.
    logging = optional(object({
      # cloudwatch_logs configures delivery of broker logs to CloudWatch Logs.
      cloudwatch_logs = optional(object({
        # enabled controls whether broker logs are delivered to CloudWatch Logs.
        enabled = optional(bool, false)

        # log_group is the CloudWatch Logs group that receives broker logs.
        # Required when enabled is true.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        log_group = optional(string, "")
      }))

      # firehose configures delivery of broker logs to a Kinesis Data Firehose stream.
      firehose = optional(object({
        # enabled controls whether broker logs are delivered to Kinesis Data Firehose.
        enabled = optional(bool, false)

        # delivery_stream is the Firehose delivery stream that receives broker logs.
        # Required when enabled is true.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        delivery_stream = optional(string, "")
      }))

      # s3 configures delivery of broker logs to an S3 bucket.
      s3 = optional(object({
        # enabled controls whether broker logs are delivered to S3.
        enabled = optional(bool, false)

        # bucket is the S3 bucket that receives broker logs.
        # Required when enabled is true.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket = optional(string, "")

        # prefix is an optional S3 key prefix for log objects.
        # Example: "msk-logs/my-cluster/" produces keys like "msk-logs/my-cluster/broker-1/...".
        prefix = optional(string, "")
      }))
    }))

    # enhanced_monitoring sets the level of CloudWatch metrics published by the cluster.
    # "DEFAULT": cluster-level and topic-level metrics.
    # "PER_BROKER": adds per-broker metrics.
    # "PER_TOPIC_PER_BROKER": adds per-topic-per-broker metrics.
    # "PER_TOPIC_PER_PARTITION": most granular, adds per-partition metrics.
    enhanced_monitoring = optional(string, "")

    # jmx_exporter_enabled enables the Prometheus JMX Exporter on all brokers.
    # When enabled, JMX metrics are available on port 11001 for Prometheus scraping.
    # Provides detailed JVM and Kafka broker metrics.
    jmx_exporter_enabled = optional(bool, false)

    # node_exporter_enabled enables the Prometheus Node Exporter on all brokers.
    # When enabled, host-level metrics (CPU, memory, disk, network) are available
    # on port 11002 for Prometheus scraping.
    node_exporter_enabled = optional(bool, false)

    # rebalancing_status controls intelligent rebalancing on Express-broker clusters
    # (instance_type express.*): MSK automatically redistributes partitions when
    # brokers are added or removed. "ACTIVE" (the AWS default for Express clusters)
    # or "PAUSED". Not applicable to standard kafka.* instance types.
    rebalancing_status = optional(string, "")

    # topics declares Apache Kafka topics managed WITH the cluster through the
    # MSK topic API (CreateTopic/UpdateTopic/DeleteTopic) -- no Kafka client,
    # bootstrap connectivity, or client authentication setup is needed; the
    # control plane manages them via IAM alone. Each entry is keyed by its
    # topic name and follows the cluster's lifecycle.
    #
    # Declare the topics your applications depend on here (the chart-wiring
    # shape: a cluster and its contract topics deploy as one unit); leave
    # application-owned ad-hoc topics to the applications themselves, or set
    # auto.create.topics.enable in server_properties.
    #
    # Deleting a declared topic entry deletes the topic and its data. Topic
    # deletion requires delete.topic.enable=true on the cluster (MSK's default;
    # only relevant if server_properties overrides it to false, in which case
    # destroy operations on declared topics hang until the server property is
    # restored).
    topics = optional(list(object({
      # name is the Kafka topic name. Letters, digits, '.', '_', and '-' only,
      # up to 249 characters (Kafka's own naming contract). Names beginning with
      # "__" are reserved for Kafka-internal topics (e.g. __consumer_offsets) --
      # avoid declaring them.
      # ForceNew: a Kafka topic cannot be renamed; changing the name replaces the
      # topic and its data.
      name = string

      # partition_count is the number of partitions for the topic.
      # Kafka only supports INCREASING the partition count in place; declaring a
      # lower count than the topic currently has fails the update (decrease is
      # not possible without replacing the topic).
      partition_count = number

      # replication_factor is the number of replicas per partition. Cannot exceed
      # the cluster's number_of_broker_nodes (CEL-enforced on the spec). Use 3 for
      # production durability, matching MSK's default.replication.factor guidance.
      # ForceNew: changing the replication factor replaces the topic.
      replication_factor = number

      # configs are topic-level Apache Kafka configuration overrides, e.g.
      # "retention.ms", "cleanup.policy", "min.insync.replicas",
      # "retention.bytes". Keys and values follow Kafka's topic-config vocabulary;
      # values are strings exactly as Kafka accepts them ("-1" for unlimited
      # retention.bytes). Updated in place via the MSK topic API.
      configs = optional(map(string), {})
    })), [])
  })
}
