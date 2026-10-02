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
  description = "AwsLambdaEventSourceMapping specification"
  type = object({
    # The AWS region the mapping is created in -- must be the region of
    # both the function and the event source.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # The Lambda function the mapping invokes. Repointing a live mapping
    # to a different function is an in-place update -- consumption
    # continues from the tracked position. Reference an AwsLambda
    # function_arn output or pass a literal function ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    function_arn = string

    # The ARN of the AWS event source: an SQS queue (queue_arn), a
    # Kinesis stream (stream_arn), a DynamoDB stream (the table's
    # stream_arn output), an MSK cluster (cluster_arn -- also set
    # topics), an Amazon MQ broker (also set queue), or a DocumentDB
    # change stream. Create-time immutable -- a different source is a
    # different mapping. The default reference targets AwsSqsQueue;
    # reference other kinds explicitly (e.g. kind AwsKinesisStream
    # fieldPath status.outputs.stream_arn, kind AwsDynamodb fieldPath
    # status.outputs.stream_arn, kind AwsMskCluster fieldPath
    # status.outputs.cluster_arn) or pass a literal ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    event_source_arn = optional(string, "")

    # A self-managed (non-MSK) Apache Kafka cluster as the event source.
    # Create-time immutable.
    self_managed_kafka = optional(object({
      # Kafka bootstrap servers as "host:port" pairs, e.g.
      # ["kafka1.example.com:9092", "kafka2.example.com:9092"].
      bootstrap_servers = list(string)
    }))

    # Pause consumption without deleting the mapping -- the tracked
    # position is retained, so re-enabling resumes where it stopped.
    # False (the default) keeps the mapping actively polling.
    disabled = optional(bool, false)

    # Records per invocation batch. 0 keeps the source's AWS default
    # (10 for SQS and DocumentDB, 100 for Kinesis/DynamoDB/Kafka/MQ).
    # Ceilings are per-source (10,000 for streams and SQS FIFO/standard
    # with long batching; 10 for MQ) -- AWS validates the exact bound.
    # Batches above 10 records for SQS require a batching window.
    batch_size = optional(number, 0)

    # How long (seconds, 0-300) the poller gathers records before
    # invoking, trading latency for fuller batches. 0 keeps the AWS
    # default (invoke as soon as records are available).
    maximum_batching_window_seconds = optional(number, 0)

    # Event-pattern filters applied BEFORE invocation -- records matching
    # no filter are discarded without billing function time. Up to 10
    # patterns, OR-ed together, each an EventBridge-style JSON pattern
    # against the record shape of the source.
    filters = optional(list(object({
      # An EventBridge-style JSON pattern matched against the record,
      # e.g. {"body":{"type":["order.created"]}} for an SQS source.
      pattern = optional(string, "")
    })), [])

    # The KMS key that encrypts the filter criteria at rest. Empty uses
    # an AWS-owned key. Reference an AwsKmsKey key_arn output or pass a
    # literal key ARN.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_arn = optional(string, "")

    # Report per-record failures from the function ("ReportBatchItemFailures")
    # so only the failed records of a batch are retried instead of the
    # whole batch -- the right setting for almost every SQS and stream
    # consumer that processes records independently.
    function_response_types = optional(list(string), [])

    # Maximum concurrent function invocations this SQS mapping may drive
    # -- a per-mapping throttle below the function's own concurrency.
    # Minimum 2; the effective ceiling is the function's concurrency
    # (AWS validates it at deploy time -- it routinely exceeds any fixed
    # bound, so none is imposed here). 0 leaves scaling to AWS. SQS
    # sources only.
    scaling_max_concurrency = optional(number, 0)

    # Emit the mapping's CloudWatch metrics: "EventCount" (records
    # delivered to the function), "ErrorCount" (records that failed
    # processing), and/or "KafkaMetrics" (poller/consumer-lag metrics --
    # Kafka sources only). Off by default; metrics are billed.
    metrics = optional(list(string), [])

    # Where to start reading a stream source, create-time immutable:
    # "TRIM_HORIZON" (oldest available record -- process the backlog),
    # "LATEST" (only new records), or "AT_TIMESTAMP" (from
    # starting_position_timestamp; Kinesis only). Required for Kinesis,
    # DynamoDB Streams, MSK, self-managed Kafka, and DocumentDB; must
    # stay empty for SQS and MQ.
    starting_position = optional(string, "")

    # The UTC RFC3339 instant to start reading from (e.g.
    # "2026-07-04T00:00:00Z"). Required with (and only meaningful for)
    # starting_position AT_TIMESTAMP.
    starting_position_timestamp = optional(string, "")

    # Concurrent batches processed per shard, 1-10 -- multiplies
    # per-shard throughput while preserving per-partition-key ordering.
    # 0 keeps the AWS default (1). Kinesis and DynamoDB streams only.
    parallelization_factor = optional(number, 0)

    # Discard records older than this (seconds, 60-604800), or -1 (the
    # AWS default) to never age records out. 0 keeps the AWS default.
    # Stream sources only.
    maximum_record_age_seconds = optional(number, 0)

    # Retries per failed batch before the batch is discarded (or sent
    # to on_failure_destination_arn), 0-10000, or -1 (the AWS default)
    # to retry until the records expire. 0 means no retries. Stream
    # sources only.
    maximum_retry_attempts = optional(number)

    # On a function error, split the failing batch in two and retry the
    # halves -- isolates a single poison record at the cost of
    # reprocessing its neighbors (make the function idempotent). Stream
    # sources only.
    bisect_batch_on_function_error = optional(bool, false)

    # Group records into fixed windows (seconds, 0-900) for streaming
    # aggregations -- the function receives a rolling state across the
    # window. Stream sources only.
    tumbling_window_seconds = optional(number, 0)

    # Where discarded batches' metadata is sent after retries are
    # exhausted: an SQS queue or SNS topic ARN (Kafka sources may also
    # target an S3 bucket). Reference an AwsSqsQueue queue_arn output or
    # pass an explicit-kind reference / literal ARN. Stream and Kafka
    # sources only.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    on_failure_destination_arn = optional(string, "")

    # The Kafka topics to consume, 1-249 characters each. Required for
    # MSK and self-managed Kafka sources; must stay empty otherwise.
    topics = optional(list(string), [])

    # The Kafka consumer group ID to join, create-time immutable. Empty
    # lets AWS generate one. Setting it lets the mapping resume an
    # existing group's committed offsets (starting_position is then
    # ignored for partitions with committed offsets).
    kafka_consumer_group_id = optional(string, "")

    # Authentication and network access the poller uses to reach the
    # source: SASL credentials, mTLS client certificates, VPC subnets
    # and security groups (self-managed Kafka), or broker credentials
    # (MQ). Each entry pairs a type with the Secrets Manager secret ARN
    # or VPC resource URI it points at.
    source_access_configurations = optional(list(object({
      # The access type: "BASIC_AUTH" (MQ / SASL PLAIN),
      # "SASL_SCRAM_256_AUTH" / "SASL_SCRAM_512_AUTH" (Kafka SASL),
      # "CLIENT_CERTIFICATE_TLS_AUTH" (Kafka mTLS),
      # "SERVER_ROOT_CA_CERTIFICATE" (private CA), "VPC_SUBNET" /
      # "VPC_SECURITY_GROUP" (self-managed Kafka networking), or
      # "VIRTUAL_HOST" (RabbitMQ).
      type = string

      # The value for the type: a Secrets Manager secret ARN (auth types),
      # "subnet:<subnet-id>" / "security_group:<sg-id>" (VPC types), or
      # the virtual host name (VIRTUAL_HOST).
      uri = string
    })), [])

    # Confluent / Glue schema registry integration for Kafka sources:
    # the poller validates and deserializes records against registered
    # schemas before invoking the function.
    schema_registry = optional(object({
      # The registry location: an AWS Glue schema registry ARN or a
      # Confluent registry HTTPS URL.
      uri = string

      # What the function receives: "JSON" (deserialized to JSON) or
      # "SOURCE" (the original serialized bytes with the schema header
      # stripped).
      event_record_format = string

      # Which record parts are validated against the registry: "KEY"
      # and/or "VALUE".
      validation_attributes = optional(list(string), [])

      # How the poller authenticates to the registry: type "BASIC_AUTH" /
      # "CLIENT_CERTIFICATE_TLS_AUTH" / "SERVER_ROOT_CA_CERTIFICATE" with
      # the Secrets Manager secret ARN in uri.
      access_configurations = optional(list(object({
        # The access type: "BASIC_AUTH" (MQ / SASL PLAIN),
        # "SASL_SCRAM_256_AUTH" / "SASL_SCRAM_512_AUTH" (Kafka SASL),
        # "CLIENT_CERTIFICATE_TLS_AUTH" (Kafka mTLS),
        # "SERVER_ROOT_CA_CERTIFICATE" (private CA), "VPC_SUBNET" /
        # "VPC_SECURITY_GROUP" (self-managed Kafka networking), or
        # "VIRTUAL_HOST" (RabbitMQ).
        type = string

        # The value for the type: a Secrets Manager secret ARN (auth types),
        # "subnet:<subnet-id>" / "security_group:<sg-id>" (VPC types), or
        # the virtual host name (VIRTUAL_HOST).
        uri = string
      })), [])
    }))

    # Dedicated pollers for Kafka sources: pin the fleet between
    # minimum_pollers and maximum_pollers for predictable throughput
    # instead of AWS's reactive scaling.
    provisioned_pollers = optional(object({
      # The floor of always-running pollers, 1-200. 0 keeps the AWS
      # default (1).
      minimum_pollers = optional(number, 0)

      # The ceiling of pollers AWS may scale to, 1-2000. 0 keeps the AWS
      # default (200).
      maximum_pollers = optional(number, 0)

      # Share one provisioned poller fleet across mappings by naming a
      # poller group, 1-128 characters: every mapping naming the same group
      # draws from (and is jointly capped by) one fleet instead of
      # provisioning its own -- the cost lever for many low-traffic topics.
      poller_group_name = optional(string, "")
    }))

    # The Amazon MQ broker queue to consume (exactly one). Required for
    # MQ sources; must stay empty otherwise.
    mq_queue = optional(string, "")

    # DocumentDB change-stream options. Required for DocumentDB sources.
    document_db = optional(object({
      # The database whose change stream is consumed.
      database_name = string

      # Consume a single collection's changes. Empty consumes the whole
      # database.
      collection_name = optional(string, "")

      # What change events carry: "UpdateLookup" (the full current
      # document alongside the change) or "Default" (the change delta
      # only). Empty keeps the AWS default (Default).
      full_document = optional(string, "")
    }))
  })
}
