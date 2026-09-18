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
  description = "AwsKinesisFirehose specification"
  type = object({
    # The AWS region where the resource will be created.
    # Example: "us-west-2", "eu-west-1"
    region = string

    # Kinesis Data Stream source configuration. When set, Firehose reads from
    # the specified stream instead of accepting Direct PUT calls. The entire
    # source configuration is ForceNew -- it cannot be changed after creation.
    #
    # Mutually exclusive with msk_source. When a stream source is configured,
    # server-side encryption (sse_enabled) must NOT be set -- the source stream
    # handles its own encryption.
    kinesis_stream_source = optional(object({
      # ARN of the Kinesis Data Stream to read from. Firehose creates an internal
      # consumer and reads all shards. The stream must exist before the delivery
      # stream is created.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      stream_arn = string

      # IAM role ARN that grants Firehose permission to read from the Kinesis
      # stream. The role must have kinesis:GetRecords, kinesis:GetShardIterator,
      # kinesis:DescribeStream, and kinesis:ListShards permissions.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = string
    }))

    # Amazon MSK source configuration. When set, Firehose reads from a topic
    # on the specified MSK cluster instead of accepting Direct PUT calls. The
    # entire source configuration is ForceNew -- it cannot be changed after
    # creation.
    #
    # Mutually exclusive with kinesis_stream_source. When an MSK source is
    # configured, server-side encryption (sse_enabled) must NOT be set -- the
    # source cluster handles its own encryption.
    msk_source = optional(object({
      # ARN of the MSK cluster to read from. The cluster must exist and have
      # IAM access control enabled before the delivery stream is created.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      msk_cluster_arn = string

      # Name of the Kafka topic to read from. The topic must exist on the
      # cluster before the delivery stream is created.
      topic_name = string

      # How Firehose connects to the MSK cluster:
      # - "PRIVATE" -- connect through the cluster's private brokers inside its
      #   VPC (the common case for provisioned and serverless clusters).
      # - "PUBLIC" -- connect through the cluster's public endpoints (requires
      #   public access to be enabled on the cluster).
      connectivity = string

      # IAM role ARN that grants Firehose permission to connect to and read from
      # the MSK cluster. The role must have kafka:GetBootstrapBrokers,
      # kafka:DescribeCluster, kafka:DescribeClusterV2, and the kafka-cluster:*
      # data-plane permissions (Connect, DescribeTopic, ReadData,
      # DescribeGroup) on the cluster, topic, and consumer group.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = string

      # Start reading the topic from this point in time instead of the current
      # offset. RFC 3339 format (e.g., "2026-05-01T00:00:00Z"). When absent,
      # Firehose starts from the latest offset at creation time.
      read_from_timestamp = optional(string, "")
    }))

    # Enable server-side encryption for data at rest in the delivery stream
    # buffer. Only valid for Direct PUT sources -- when using a Kinesis stream
    # or MSK source, encryption is handled by the source.
    #
    # When true and sse_kms_key_arn is absent, uses the AWS-owned CMK.
    # When true and sse_kms_key_arn is present, uses a customer-managed CMK.
    sse_enabled = optional(bool, false)

    # Customer-managed KMS key ARN for server-side encryption. When set,
    # Firehose uses this key instead of the AWS-owned CMK. Requires
    # sse_enabled to be true.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    sse_kms_key_arn = optional(string, "")

    # Extended S3 destination for data lake storage. Supports compression,
    # record transformation, dynamic partitioning, and Parquet/ORC format
    # conversion via AWS Glue Data Catalog. The most feature-rich destination.
    extended_s3 = optional(object({
      # S3 bucket ARN where records are delivered.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      bucket_arn = string

      # IAM role ARN granting Firehose write access to the S3 bucket, KMS key
      # (if encrypted), Lambda function (if processing), and Glue catalog (if
      # format conversion is enabled).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = string

      # S3 key prefix prepended to every delivered object. Supports Firehose
      # expression syntax for dynamic prefixes:
      #   "data/year=!{timestamp:yyyy}/month=!{timestamp:MM}/day=!{timestamp:dd}/"
      #
      # When dynamic partitioning is enabled, use partitioning keys:
      #   "data/customer=!{partitionKeyFromQuery:customer_id}/"
      prefix = optional(string, "")

      # S3 key prefix for records that fail transformation or delivery.
      error_output_prefix = optional(string, "")

      # Compression format applied before writing to S3. When data format
      # conversion is enabled, compression is applied to the converted
      # (Parquet/ORC) output -- in that case, use the format-native compression
      # (configured in data_format_conversion) and leave this as UNCOMPRESSED.
      #
      # Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
      # Default: "UNCOMPRESSED".
      compression_format = optional(string, "")

      # Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      kms_key_arn = optional(string, "")

      # Buffering hints for S3 delivery. Default: 300s interval, 5 MiB size.
      buffering = optional(object({
        # Buffer interval in seconds. Firehose flushes when this time elapses since
        # the last flush, even if the buffer size threshold has not been reached.
        #
        # Range: 0-900 seconds. Default varies by destination (typically 300;
        # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
        # values improve batching efficiency and reduce S3 object count.
        #
        # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
        # destination message carries that rule.
        interval_in_seconds = optional(number, 0)

        # Buffer size in MiB. Firehose flushes when the accumulated data reaches
        # this threshold.
        #
        # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
        # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
        # for query engines); smaller buffers provide faster delivery.
        #
        # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
        # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
        size_in_mbs = optional(number, 0)
      }))

      # IANA time zone for S3 prefix timestamp expressions.
      # Default: "UTC". Example: "US/Eastern", "Europe/London".
      custom_time_zone = optional(string, "")

      # File extension appended to delivered S3 objects (e.g., ".json", ".parquet").
      # Must start with a period. When data format conversion is enabled, the
      # extension is typically set to match the output format.
      file_extension = optional(string, "")

      # S3 backup mode for source records. When "Enabled", a copy of the original
      # (pre-transformation) records is written to s3_backup in addition to the
      # primary destination. Useful for auditing and reprocessing.
      #
      # Valid values: "Disabled" (default), "Enabled".
      s3_backup_mode = optional(string, "")

      # S3 configuration for source record backup. Required when s3_backup_mode
      # is "Enabled".
      s3_backup = optional(object({
        # S3 bucket ARN where records are delivered.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket_arn = string

        # IAM role ARN that grants Firehose permission to write to the S3 bucket.
        # The role must have s3:PutObject, s3:AbortMultipartUpload,
        # s3:GetBucketLocation, and s3:ListBucket permissions.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string

        # S3 key prefix prepended to delivered objects. Supports Firehose expression
        # syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").
        prefix = optional(string, "")

        # S3 key prefix for error output. When Firehose cannot deliver or transform
        # a record, it writes to this prefix. Uses the same expression syntax as prefix.
        error_output_prefix = optional(string, "")

        # Compression format for delivered objects. Applied before writing to S3.
        # Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
        # Default: "UNCOMPRESSED".
        compression_format = optional(string, "")

        # Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
        # When absent, S3 uses its default encryption settings (SSE-S3 or bucket
        # default encryption).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_arn = optional(string, "")

        # Buffering hints for S3 delivery.
        buffering = optional(object({
          # Buffer interval in seconds. Firehose flushes when this time elapses since
          # the last flush, even if the buffer size threshold has not been reached.
          #
          # Range: 0-900 seconds. Default varies by destination (typically 300;
          # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
          # values improve batching efficiency and reduce S3 object count.
          #
          # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
          # destination message carries that rule.
          interval_in_seconds = optional(number, 0)

          # Buffer size in MiB. Firehose flushes when the accumulated data reaches
          # this threshold.
          #
          # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
          # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
          # for query engines); smaller buffers provide faster delivery.
          #
          # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
          # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
          size_in_mbs = optional(number, 0)
        }))

        # CloudWatch logging configuration for S3 delivery errors.
        logging = optional(object({
          # Enable CloudWatch error logging for this delivery target.
          enabled = optional(bool, false)

          # CloudWatch Logs log group name where errors are published.
          # Required when enabled is true. Firehose neither creates nor
          # validates this group: CreateDeliveryStream accepts a nonexistent
          # name (live-verified 2026-08-12), and delivery errors are silently
          # dropped until a log group with exactly this name exists -- create
          # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
          # trail you configured here never materializes.
          log_group_name = optional(string, "")

          # CloudWatch Logs log stream name within the log group.
          # Required when enabled is true.
          log_stream_name = optional(string, "")
        }))
      }))

      # Record-transformation pipeline. Applied before compression and format
      # conversion. For dynamic partitioning, include a metadata_extraction
      # processor to define the partition keys.
      processing = optional(object({
        # Enable the processing pipeline. When true, at least one processor should
        # be configured.
        enabled = optional(bool, false)

        # Ordered list of processors. Each entry configures exactly one processor
        # type; Firehose executes them in order.
        processors = optional(list(object({
          # Invoke an AWS Lambda function to transform records. The function
          # receives batches of records and returns transformed records with a
          # status (Ok, Dropped, ProcessingFailed) per record.
          lambda = optional(object({
            # ARN of the Lambda function that transforms records. May include a
            # version or alias qualifier to pin the deployed transformation.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            lambda_arn = string

            # Buffer size in MiB that Firehose accumulates before invoking Lambda.
            # Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
            # low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
            # destination is Splunk).
            #
            # Smaller buffers invoke Lambda more frequently with smaller batches.
            # Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
            # invocation costs.
            buffer_size_in_mbs = optional(number, 0)

            # Buffer interval in seconds. Firehose invokes Lambda when this interval
            # elapses, even if the buffer size threshold has not been reached.
            # Range: 60-900 seconds. Default: 60 seconds.
            buffer_interval_in_seconds = optional(number, 0)

            # Number of times Firehose retries a failed Lambda invocation before
            # writing the record to the error output prefix.
            # Range: 0-300. Default: 3.
            number_of_retries = optional(number, 0)

            # IAM role ARN Firehose assumes to invoke the Lambda function. When
            # absent, Firehose uses the delivery stream's destination role -- the
            # right choice for almost every pipeline. Set this only when the
            # transformation function must be invoked with a DIFFERENT role than the
            # one that writes to the destination (e.g. the function lives in another
            # account). Note: AWS reports the delivery role back for unset values, and
            # the provider does not store default-valued processor parameters in
            # state -- so set this only to a non-default role, never to the delivery
            # role itself (that would cause perpetual plan diffs).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            role_arn = optional(string, "")
          }))

          # Extract partition keys from JSON records with a JQ expression. Used
          # with Extended S3 dynamic partitioning -- extracted keys are referenced
          # in the S3 prefix as !{partitionKeyFromQuery:<key>}.
          metadata_extraction = optional(object({
            # JQ expression that extracts partition keys from each JSON record.
            # The result must be an object whose keys become partition keys.
            #
            # Example: "{customer_id: .customer_id, event_type: .type}" extracts two
            # keys, referenced in the prefix as
            # "data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".
            query = string

            # JSON parsing engine used to evaluate the query.
            # Valid value: "JQ-1.6" (default -- the only engine AWS supports today).
            json_parsing_engine = optional(string, "")
          }))

          # Decompress GZIP-compressed records before delivery. Typically the first
          # processor when the source sends compressed payloads (e.g., CloudWatch
          # Logs subscription filters).
          decompression = optional(object({
            # Compression format of the incoming records.
            # Valid value: "GZIP" (the only format AWS supports today).
            compression_format = string
          }))

          # Unwrap CloudWatch Logs subscription envelopes into individual log
          # events. Use after a decompression processor when the source is a
          # CloudWatch Logs subscription filter.
          cloudwatch_log_processing = optional(object({
            # When true, extract only the log event message field, discarding the
            # CloudWatch envelope metadata. When false, records pass through with the
            # envelope intact.
            data_message_extraction = optional(bool, false)
          }))

          # Append a delimiter to every record. Use to produce newline-delimited
          # JSON (JSON lines) output for query engines and log consumers.
          #
          # Only supported on the extended_s3 destination -- delimiting is an
          # S3-object formatting concern; other destinations frame records
          # natively.
          append_delimiter = optional(object({
            # Delimiter appended to each record. Use "\\n" for newline-delimited
            # output -- the format Athena, Spark, and most log consumers expect.
            delimiter = string
          }))

          # Split multi-record aggregates (e.g., KPL-aggregated or delimited
          # payloads) into individual records before further processing. Required
          # before dynamic partitioning when producers aggregate records.
          #
          # Only supported on the extended_s3 destination -- AWS rejects it for
          # every other destination type at creation.
          record_deaggregation = optional(object({
            # How records are aggregated in the payload:
            # - "JSON" -- concatenated JSON documents (no delimiter needed).
            # - "DELIMITED" -- records separated by a custom delimiter; requires
            #   the delimiter field.
            sub_record_type = string

            # Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
            # newline). Required when sub_record_type is "DELIMITED".
            delimiter = optional(string, "")
          }))
        })), [])
      }))

      # CloudWatch error logging for S3 delivery failures.
      logging = optional(object({
        # Enable CloudWatch error logging for this delivery target.
        enabled = optional(bool, false)

        # CloudWatch Logs log group name where errors are published.
        # Required when enabled is true. Firehose neither creates nor
        # validates this group: CreateDeliveryStream accepts a nonexistent
        # name (live-verified 2026-08-12), and delivery errors are silently
        # dropped until a log group with exactly this name exists -- create
        # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
        # trail you configured here never materializes.
        log_group_name = optional(string, "")

        # CloudWatch Logs log stream name within the log group.
        # Required when enabled is true.
        log_stream_name = optional(string, "")
      }))

      # Dynamic partitioning configuration. Enables partitioning delivered data
      # by record fields (e.g., customer_id, event_type) for efficient querying
      # with Athena, Spark, or Presto. ForceNew -- cannot be enabled/disabled
      # after creation. Define the partition keys with a metadata_extraction
      # processor (or a Lambda processor emitting partition metadata) and
      # reference them in the prefix.
      dynamic_partitioning = optional(object({
        # Enable dynamic partitioning. ForceNew -- cannot be changed after creation.
        # When enabled, configure partition key expressions in the S3 prefix using
        # !{partitionKeyFromQuery:...} or !{partitionKeyFromLambda:...} syntax,
        # and define the keys with a metadata_extraction (or Lambda) processor.
        enabled = optional(bool, false)

        # Duration in seconds that Firehose retries delivery when a partition key
        # expression fails or the S3 PutObject call is throttled.
        # Range: 0-7200. Default: 300 seconds.
        retry_duration_in_seconds = optional(number, 0)
      }))

      # Data format conversion from JSON to columnar formats (Parquet or ORC)
      # using an AWS Glue Data Catalog schema. Dramatically improves query
      # performance and reduces storage cost for analytics workloads.
      data_format_conversion = optional(object({
        # Enable data format conversion. When true, a deserializer arm, a
        # serializer arm, and schema are required. When false with arms configured, the
        # conversion settings are retained but inactive (AWS permits disabling
        # conversion without discarding its configuration).
        enabled = optional(bool, false)

        # OpenX JSON deserializer arm. Handles most JSON formats including nested
        # objects -- the right choice for general use.
        open_x_json = optional(object({
          # When true (the AWS default), JSON keys are lowercased before
          # deserialization, so keys match case-insensitively against the Glue
          # schema's lowercase column names. Set to false only when the source
          # JSON keys must be matched exactly as sent (combine with
          # column_to_json_key_mappings for columns whose keys differ only by
          # case).
          case_insensitive = optional(bool)

          # Map of Glue schema column names to JSON keys that differ from them.
          # Use when a JSON key is not a valid column name -- e.g. a key that
          # collides with a Hive reserved word: {"ts": "timestamp"} reads the JSON
          # key "timestamp" into the column "ts".
          column_to_json_key_mappings = optional(map(string), {})

          # When true, dots in JSON keys are converted to underscores before
          # matching against the schema ("a.b" reads into column "a_b"). Glue
          # column names cannot contain dots, so enable this when source keys do.
          # Default: false.
          convert_dots_in_json_keys_to_underscores = optional(bool, false)
        }))

        # Apache Hive JSON deserializer arm. Use for Hive-compatible JSON when
        # records carry non-standard timestamp encodings that need explicit
        # parsing patterns.
        hive_json = optional(object({
          # Joda-Time datetime format patterns for parsing timestamp fields from
          # the source JSON (e.g. "yyyy-MM-dd'T'HH:mm:ss"). Include the special
          # value "millis" to parse epoch-millisecond timestamps. When empty,
          # Firehose uses java.sql.Timestamp::valueOf.
          timestamp_formats = optional(list(string), [])
        }))

        # Apache Parquet serializer arm. Best for read-heavy analytical workloads
        # (Athena, Spark, Presto): excellent compression, predicate pushdown, and
        # columnar pruning.
        parquet = optional(object({
          # Compression codec for Parquet pages. "SNAPPY" (the default) balances
          # speed and size; "GZIP" compresses harder at higher CPU cost --
          # preferable when S3 storage/scan cost outweighs write throughput;
          # "UNCOMPRESSED" only when downstream readers cannot decompress.
          compression = optional(string, "")

          # Parquet row-group (block) size in bytes. Firehose and query engines use
          # it for padding and row-group sizing; larger blocks improve scan
          # efficiency at the cost of memory. Minimum: 67108864 (64 MiB).
          # Default: 268435456 (256 MiB).
          block_size_bytes = optional(number, 0)

          # Parquet page size in bytes -- the smallest unit a read must fully
          # decompress. Minimum: 65536 (64 KiB). Default: 1048576 (1 MiB).
          page_size_bytes = optional(number, 0)

          # Maximum padding in bytes when writing row groups (used to align blocks
          # to HDFS-style boundaries; rarely needed on S3). Default: 0.
          max_padding_bytes = optional(number, 0)

          # Enable dictionary compression -- encodes repeated column values through
          # a dictionary. Effective when columns carry low-cardinality values.
          # Default: false.
          enable_dictionary_compression = optional(bool, false)

          # Parquet writer version. "V1" (the default) is readable by every
          # engine; "V2" enables newer encodings -- confirm downstream reader
          # support before switching. Valid values: "V1", "V2".
          writer_version = optional(string, "")
        }))

        # Apache ORC serializer arm. Best for Hive workloads: ACID support, bloom
        # filters, and built-in indexing.
        orc = optional(object({
          # Compression codec for ORC stripes. "SNAPPY" (the default) balances
          # speed and size; "ZLIB" compresses harder at higher CPU cost; "NONE"
          # only when downstream readers cannot decompress.
          compression = optional(string, "")

          # ORC block size in bytes, used for padding calculations and copy-block
          # sizing. Minimum: 67108864 (64 MiB). Default: 268435456 (256 MiB).
          block_size_bytes = optional(number, 0)

          # ORC stripe size in bytes -- the unit of independent reading. Minimum:
          # 8388608 (8 MiB). Default: 67108864 (64 MiB).
          stripe_size_bytes = optional(number, 0)

          # Column names to build bloom filters for. Bloom filters let readers
          # skip stripes that cannot contain a searched value -- list the columns
          # your queries filter on by equality.
          bloom_filter_columns = optional(list(string), [])

          # Bloom filter false-positive probability, between 0 and 1. Lower values
          # make filters more selective but larger. AWS default: 0.05. Explicit 0
          # is AWS-legal, so absence (AWS default) and 0 are distinct states.
          bloom_filter_false_positive_probability = optional(number)

          # Fraction of a column's total distinct keys above which dictionary
          # encoding is abandoned for that column, between 0 and 1. 0 (the AWS
          # default) always dictionary-encodes; 1 never abandons it.
          dictionary_key_threshold = optional(number, 0)

          # Pad stripes to HDFS-style block boundaries. Relevant for HDFS-backed
          # readers; rarely needed for S3. Default: false.
          enable_padding = optional(bool, false)

          # Maximum fraction of a stripe that may be wasted as padding, between 0
          # and 1. Only meaningful when enable_padding is true. AWS default: 0.05.
          # Explicit 0 is AWS-legal, so absence and 0 are distinct states.
          padding_tolerance = optional(number)

          # ORC file format version. "V0_12" (the default) is current; "V0_11" is
          # the legacy Hive 0.11 format -- use only for readers frozen on it.
          # Valid values: "V0_11", "V0_12".
          format_version = optional(string, "")

          # Number of rows between index entries. Must be at least 1000 when set.
          # Default: 10000.
          row_index_stride = optional(number, 0)
        }))

        # AWS Glue Data Catalog schema reference. Defines the table schema used
        # for converting JSON records to the columnar format. Required when
        # data format conversion is enabled.
        schema = optional(object({
          # Glue Data Catalog database name containing the table.
          database_name = string

          # Glue Data Catalog table name defining the record schema.
          table_name = string

          # IAM role ARN granting Firehose permission to access the Glue catalog.
          # Must have glue:GetTable and glue:GetTableVersions permissions.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          role_arn = string

          # Glue Data Catalog ID (AWS account ID). When omitted, defaults to the
          # current AWS account.
          catalog_id = optional(string, "")

          # AWS region of the Glue catalog. When omitted, defaults to the delivery
          # stream's region.
          region = optional(string, "")

          # Table version to use. Default: "LATEST".
          version_id = optional(string, "")
        }))
      }))
    }))

    # OpenSearch destination for direct indexing into an Amazon OpenSearch
    # Service domain. Supports index rotation, VPC delivery, and record
    # transformation. Failed documents are backed up to S3.
    opensearch = optional(object({
      # ARN of the OpenSearch domain. Mutually exclusive with cluster_endpoint.
      # Use this for domains managed within the same AWS account.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      domain_arn = optional(string, "")

      # OpenSearch cluster endpoint URL. Mutually exclusive with domain_arn.
      # Use this for cross-account domains or non-standard endpoints.
      # Format: "https://search-domain-xxxx.us-east-1.es.amazonaws.com"
      cluster_endpoint = optional(string, "")

      # Name of the OpenSearch index to deliver records to. Required.
      # When index_rotation_period is set, this becomes the index prefix and
      # Firehose appends a timestamp suffix (e.g., "logs-2026-02-15").
      index_name = string

      # IAM role ARN granting Firehose permission to write to OpenSearch.
      # Must have es:ESHttpPut and es:ESHttpGet permissions on the domain.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = string

      # Index rotation period. Firehose appends a timestamp suffix to index_name
      # and creates a new index at each rotation boundary.
      #
      # Valid values: "NoRotation", "OneHour", "OneDay" (default), "OneWeek", "OneMonth".
      #
      # "NoRotation" writes all records to the same index (use for small, static datasets).
      # "OneDay" is recommended for most log and analytics use cases.
      index_rotation_period = optional(string, "")

      # OpenSearch document type name. Only relevant for Elasticsearch 6.x and
      # earlier (OpenSearch does not use document types). Leave empty for
      # OpenSearch domains.
      type_name = optional(string, "")

      # How Firehose assigns document IDs:
      # - "FIREHOSE_DEFAULT" -- Firehose generates a unique document ID per
      #   record. Protects against duplicates on retry but disables OpenSearch
      #   ID-based deduplication.
      # - "NO_DOCUMENT_ID" -- no ID is sent; OpenSearch auto-generates one.
      #   Improves indexing throughput and lets retried deliveries deduplicate.
      #
      # When empty, AWS applies its service default (FIREHOSE_DEFAULT).
      default_document_id_format = optional(string, "")

      # Buffering hints for OpenSearch delivery. Default: 300s interval, 5 MiB.
      # Maximum size for OpenSearch destinations: 100 MiB.
      buffering = optional(object({
        # Buffer interval in seconds. Firehose flushes when this time elapses since
        # the last flush, even if the buffer size threshold has not been reached.
        #
        # Range: 0-900 seconds. Default varies by destination (typically 300;
        # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
        # values improve batching efficiency and reduce S3 object count.
        #
        # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
        # destination message carries that rule.
        interval_in_seconds = optional(number, 0)

        # Buffer size in MiB. Firehose flushes when the accumulated data reaches
        # this threshold.
        #
        # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
        # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
        # for query engines); smaller buffers provide faster delivery.
        #
        # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
        # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
        size_in_mbs = optional(number, 0)
      }))

      # Retry duration in seconds for failed OpenSearch index requests.
      # Range: 0-7200. Default: 300 seconds.
      # Set to 0 to disable retries (failed documents go directly to S3 backup).
      retry_duration_in_seconds = optional(number, 0)

      # S3 backup mode for documents. Controls when records are written to S3.
      #
      # Valid values:
      # - "FailedDocumentsOnly" (default) -- only documents that fail indexing
      #   are backed up to S3.
      # - "AllDocuments" -- all documents are backed up to S3 in addition to
      #   being indexed in OpenSearch.
      #
      # ForceNew -- changing the backup mode replaces the delivery stream.
      s3_backup_mode = optional(string, "")

      # S3 configuration for backing up failed (or all) documents. Required.
      s3_config = object({
        # S3 bucket ARN where records are delivered.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket_arn = string

        # IAM role ARN that grants Firehose permission to write to the S3 bucket.
        # The role must have s3:PutObject, s3:AbortMultipartUpload,
        # s3:GetBucketLocation, and s3:ListBucket permissions.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string

        # S3 key prefix prepended to delivered objects. Supports Firehose expression
        # syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").
        prefix = optional(string, "")

        # S3 key prefix for error output. When Firehose cannot deliver or transform
        # a record, it writes to this prefix. Uses the same expression syntax as prefix.
        error_output_prefix = optional(string, "")

        # Compression format for delivered objects. Applied before writing to S3.
        # Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
        # Default: "UNCOMPRESSED".
        compression_format = optional(string, "")

        # Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
        # When absent, S3 uses its default encryption settings (SSE-S3 or bucket
        # default encryption).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_arn = optional(string, "")

        # Buffering hints for S3 delivery.
        buffering = optional(object({
          # Buffer interval in seconds. Firehose flushes when this time elapses since
          # the last flush, even if the buffer size threshold has not been reached.
          #
          # Range: 0-900 seconds. Default varies by destination (typically 300;
          # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
          # values improve batching efficiency and reduce S3 object count.
          #
          # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
          # destination message carries that rule.
          interval_in_seconds = optional(number, 0)

          # Buffer size in MiB. Firehose flushes when the accumulated data reaches
          # this threshold.
          #
          # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
          # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
          # for query engines); smaller buffers provide faster delivery.
          #
          # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
          # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
          size_in_mbs = optional(number, 0)
        }))

        # CloudWatch logging configuration for S3 delivery errors.
        logging = optional(object({
          # Enable CloudWatch error logging for this delivery target.
          enabled = optional(bool, false)

          # CloudWatch Logs log group name where errors are published.
          # Required when enabled is true. Firehose neither creates nor
          # validates this group: CreateDeliveryStream accepts a nonexistent
          # name (live-verified 2026-08-12), and delivery errors are silently
          # dropped until a log group with exactly this name exists -- create
          # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
          # trail you configured here never materializes.
          log_group_name = optional(string, "")

          # CloudWatch Logs log stream name within the log group.
          # Required when enabled is true.
          log_stream_name = optional(string, "")
        }))
      })

      # Record-transformation pipeline applied before indexing.
      processing = optional(object({
        # Enable the processing pipeline. When true, at least one processor should
        # be configured.
        enabled = optional(bool, false)

        # Ordered list of processors. Each entry configures exactly one processor
        # type; Firehose executes them in order.
        processors = optional(list(object({
          # Invoke an AWS Lambda function to transform records. The function
          # receives batches of records and returns transformed records with a
          # status (Ok, Dropped, ProcessingFailed) per record.
          lambda = optional(object({
            # ARN of the Lambda function that transforms records. May include a
            # version or alias qualifier to pin the deployed transformation.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            lambda_arn = string

            # Buffer size in MiB that Firehose accumulates before invoking Lambda.
            # Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
            # low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
            # destination is Splunk).
            #
            # Smaller buffers invoke Lambda more frequently with smaller batches.
            # Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
            # invocation costs.
            buffer_size_in_mbs = optional(number, 0)

            # Buffer interval in seconds. Firehose invokes Lambda when this interval
            # elapses, even if the buffer size threshold has not been reached.
            # Range: 60-900 seconds. Default: 60 seconds.
            buffer_interval_in_seconds = optional(number, 0)

            # Number of times Firehose retries a failed Lambda invocation before
            # writing the record to the error output prefix.
            # Range: 0-300. Default: 3.
            number_of_retries = optional(number, 0)

            # IAM role ARN Firehose assumes to invoke the Lambda function. When
            # absent, Firehose uses the delivery stream's destination role -- the
            # right choice for almost every pipeline. Set this only when the
            # transformation function must be invoked with a DIFFERENT role than the
            # one that writes to the destination (e.g. the function lives in another
            # account). Note: AWS reports the delivery role back for unset values, and
            # the provider does not store default-valued processor parameters in
            # state -- so set this only to a non-default role, never to the delivery
            # role itself (that would cause perpetual plan diffs).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            role_arn = optional(string, "")
          }))

          # Extract partition keys from JSON records with a JQ expression. Used
          # with Extended S3 dynamic partitioning -- extracted keys are referenced
          # in the S3 prefix as !{partitionKeyFromQuery:<key>}.
          metadata_extraction = optional(object({
            # JQ expression that extracts partition keys from each JSON record.
            # The result must be an object whose keys become partition keys.
            #
            # Example: "{customer_id: .customer_id, event_type: .type}" extracts two
            # keys, referenced in the prefix as
            # "data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".
            query = string

            # JSON parsing engine used to evaluate the query.
            # Valid value: "JQ-1.6" (default -- the only engine AWS supports today).
            json_parsing_engine = optional(string, "")
          }))

          # Decompress GZIP-compressed records before delivery. Typically the first
          # processor when the source sends compressed payloads (e.g., CloudWatch
          # Logs subscription filters).
          decompression = optional(object({
            # Compression format of the incoming records.
            # Valid value: "GZIP" (the only format AWS supports today).
            compression_format = string
          }))

          # Unwrap CloudWatch Logs subscription envelopes into individual log
          # events. Use after a decompression processor when the source is a
          # CloudWatch Logs subscription filter.
          cloudwatch_log_processing = optional(object({
            # When true, extract only the log event message field, discarding the
            # CloudWatch envelope metadata. When false, records pass through with the
            # envelope intact.
            data_message_extraction = optional(bool, false)
          }))

          # Append a delimiter to every record. Use to produce newline-delimited
          # JSON (JSON lines) output for query engines and log consumers.
          #
          # Only supported on the extended_s3 destination -- delimiting is an
          # S3-object formatting concern; other destinations frame records
          # natively.
          append_delimiter = optional(object({
            # Delimiter appended to each record. Use "\\n" for newline-delimited
            # output -- the format Athena, Spark, and most log consumers expect.
            delimiter = string
          }))

          # Split multi-record aggregates (e.g., KPL-aggregated or delimited
          # payloads) into individual records before further processing. Required
          # before dynamic partitioning when producers aggregate records.
          #
          # Only supported on the extended_s3 destination -- AWS rejects it for
          # every other destination type at creation.
          record_deaggregation = optional(object({
            # How records are aggregated in the payload:
            # - "JSON" -- concatenated JSON documents (no delimiter needed).
            # - "DELIMITED" -- records separated by a custom delimiter; requires
            #   the delimiter field.
            sub_record_type = string

            # Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
            # newline). Required when sub_record_type is "DELIMITED".
            delimiter = optional(string, "")
          }))
        })), [])
      }))

      # CloudWatch error logging for OpenSearch delivery failures.
      logging = optional(object({
        # Enable CloudWatch error logging for this delivery target.
        enabled = optional(bool, false)

        # CloudWatch Logs log group name where errors are published.
        # Required when enabled is true. Firehose neither creates nor
        # validates this group: CreateDeliveryStream accepts a nonexistent
        # name (live-verified 2026-08-12), and delivery errors are silently
        # dropped until a log group with exactly this name exists -- create
        # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
        # trail you configured here never materializes.
        log_group_name = optional(string, "")

        # CloudWatch Logs log stream name within the log group.
        # Required when enabled is true.
        log_stream_name = optional(string, "")
      }))

      # VPC configuration for delivering to VPC-deployed OpenSearch domains.
      # ForceNew -- the VPC config cannot be changed after creation.
      # When absent, Firehose delivers over the public internet.
      vpc_config = optional(object({
        # Subnet IDs where Firehose creates ENIs for VPC delivery. Provide at
        # least one subnet. For high availability, use subnets in multiple AZs.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        subnet_ids = optional(list(string), [])

        # Security group IDs applied to the ENIs. Must allow outbound HTTPS (443)
        # traffic to the destination.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        security_group_ids = optional(list(string), [])

        # IAM role ARN for Firehose to manage VPC ENIs. The role must have
        # ec2:CreateNetworkInterface, ec2:DescribeNetworkInterfaces,
        # ec2:DeleteNetworkInterface, and ec2:DescribeVpcs permissions.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string
      }))
    }))

    # OpenSearch Serverless destination for indexing into an OpenSearch
    # Serverless collection. Failed documents are backed up to S3.
    opensearch_serverless = optional(object({
      # Endpoint of the OpenSearch Serverless collection.
      # Format: "https://<collection-id>.<region>.aoss.amazonaws.com"
      collection_endpoint = string

      # Name of the index to deliver records to. Required. The index must be
      # permitted by the collection's data access policy for the delivery role.
      index_name = string

      # IAM role ARN granting Firehose permission to write to the collection.
      # The role must be granted aoss:APIAccessAll on the collection and be
      # listed in the collection's data access policy with document-write
      # permission on the index.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = string

      # Buffering hints for delivery. Default: 300s interval, 5 MiB.
      # Maximum size for OpenSearch Serverless destinations: 100 MiB.
      buffering = optional(object({
        # Buffer interval in seconds. Firehose flushes when this time elapses since
        # the last flush, even if the buffer size threshold has not been reached.
        #
        # Range: 0-900 seconds. Default varies by destination (typically 300;
        # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
        # values improve batching efficiency and reduce S3 object count.
        #
        # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
        # destination message carries that rule.
        interval_in_seconds = optional(number, 0)

        # Buffer size in MiB. Firehose flushes when the accumulated data reaches
        # this threshold.
        #
        # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
        # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
        # for query engines); smaller buffers provide faster delivery.
        #
        # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
        # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
        size_in_mbs = optional(number, 0)
      }))

      # Retry duration in seconds for failed index requests.
      # Range: 0-7200. Default: 300 seconds.
      # Set to 0 to disable retries (failed documents go directly to S3 backup).
      retry_duration_in_seconds = optional(number, 0)

      # S3 backup mode for documents. Controls when records are written to S3.
      #
      # Valid values:
      # - "FailedDocumentsOnly" (default) -- only documents that fail indexing
      #   are backed up to S3.
      # - "AllDocuments" -- all documents are backed up to S3 in addition to
      #   being indexed.
      #
      # ForceNew -- changing the backup mode replaces the delivery stream.
      s3_backup_mode = optional(string, "")

      # S3 configuration for backing up failed (or all) documents. Required.
      s3_config = object({
        # S3 bucket ARN where records are delivered.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket_arn = string

        # IAM role ARN that grants Firehose permission to write to the S3 bucket.
        # The role must have s3:PutObject, s3:AbortMultipartUpload,
        # s3:GetBucketLocation, and s3:ListBucket permissions.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string

        # S3 key prefix prepended to delivered objects. Supports Firehose expression
        # syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").
        prefix = optional(string, "")

        # S3 key prefix for error output. When Firehose cannot deliver or transform
        # a record, it writes to this prefix. Uses the same expression syntax as prefix.
        error_output_prefix = optional(string, "")

        # Compression format for delivered objects. Applied before writing to S3.
        # Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
        # Default: "UNCOMPRESSED".
        compression_format = optional(string, "")

        # Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
        # When absent, S3 uses its default encryption settings (SSE-S3 or bucket
        # default encryption).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_arn = optional(string, "")

        # Buffering hints for S3 delivery.
        buffering = optional(object({
          # Buffer interval in seconds. Firehose flushes when this time elapses since
          # the last flush, even if the buffer size threshold has not been reached.
          #
          # Range: 0-900 seconds. Default varies by destination (typically 300;
          # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
          # values improve batching efficiency and reduce S3 object count.
          #
          # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
          # destination message carries that rule.
          interval_in_seconds = optional(number, 0)

          # Buffer size in MiB. Firehose flushes when the accumulated data reaches
          # this threshold.
          #
          # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
          # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
          # for query engines); smaller buffers provide faster delivery.
          #
          # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
          # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
          size_in_mbs = optional(number, 0)
        }))

        # CloudWatch logging configuration for S3 delivery errors.
        logging = optional(object({
          # Enable CloudWatch error logging for this delivery target.
          enabled = optional(bool, false)

          # CloudWatch Logs log group name where errors are published.
          # Required when enabled is true. Firehose neither creates nor
          # validates this group: CreateDeliveryStream accepts a nonexistent
          # name (live-verified 2026-08-12), and delivery errors are silently
          # dropped until a log group with exactly this name exists -- create
          # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
          # trail you configured here never materializes.
          log_group_name = optional(string, "")

          # CloudWatch Logs log stream name within the log group.
          # Required when enabled is true.
          log_stream_name = optional(string, "")
        }))
      })

      # Record-transformation pipeline applied before indexing.
      processing = optional(object({
        # Enable the processing pipeline. When true, at least one processor should
        # be configured.
        enabled = optional(bool, false)

        # Ordered list of processors. Each entry configures exactly one processor
        # type; Firehose executes them in order.
        processors = optional(list(object({
          # Invoke an AWS Lambda function to transform records. The function
          # receives batches of records and returns transformed records with a
          # status (Ok, Dropped, ProcessingFailed) per record.
          lambda = optional(object({
            # ARN of the Lambda function that transforms records. May include a
            # version or alias qualifier to pin the deployed transformation.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            lambda_arn = string

            # Buffer size in MiB that Firehose accumulates before invoking Lambda.
            # Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
            # low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
            # destination is Splunk).
            #
            # Smaller buffers invoke Lambda more frequently with smaller batches.
            # Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
            # invocation costs.
            buffer_size_in_mbs = optional(number, 0)

            # Buffer interval in seconds. Firehose invokes Lambda when this interval
            # elapses, even if the buffer size threshold has not been reached.
            # Range: 60-900 seconds. Default: 60 seconds.
            buffer_interval_in_seconds = optional(number, 0)

            # Number of times Firehose retries a failed Lambda invocation before
            # writing the record to the error output prefix.
            # Range: 0-300. Default: 3.
            number_of_retries = optional(number, 0)

            # IAM role ARN Firehose assumes to invoke the Lambda function. When
            # absent, Firehose uses the delivery stream's destination role -- the
            # right choice for almost every pipeline. Set this only when the
            # transformation function must be invoked with a DIFFERENT role than the
            # one that writes to the destination (e.g. the function lives in another
            # account). Note: AWS reports the delivery role back for unset values, and
            # the provider does not store default-valued processor parameters in
            # state -- so set this only to a non-default role, never to the delivery
            # role itself (that would cause perpetual plan diffs).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            role_arn = optional(string, "")
          }))

          # Extract partition keys from JSON records with a JQ expression. Used
          # with Extended S3 dynamic partitioning -- extracted keys are referenced
          # in the S3 prefix as !{partitionKeyFromQuery:<key>}.
          metadata_extraction = optional(object({
            # JQ expression that extracts partition keys from each JSON record.
            # The result must be an object whose keys become partition keys.
            #
            # Example: "{customer_id: .customer_id, event_type: .type}" extracts two
            # keys, referenced in the prefix as
            # "data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".
            query = string

            # JSON parsing engine used to evaluate the query.
            # Valid value: "JQ-1.6" (default -- the only engine AWS supports today).
            json_parsing_engine = optional(string, "")
          }))

          # Decompress GZIP-compressed records before delivery. Typically the first
          # processor when the source sends compressed payloads (e.g., CloudWatch
          # Logs subscription filters).
          decompression = optional(object({
            # Compression format of the incoming records.
            # Valid value: "GZIP" (the only format AWS supports today).
            compression_format = string
          }))

          # Unwrap CloudWatch Logs subscription envelopes into individual log
          # events. Use after a decompression processor when the source is a
          # CloudWatch Logs subscription filter.
          cloudwatch_log_processing = optional(object({
            # When true, extract only the log event message field, discarding the
            # CloudWatch envelope metadata. When false, records pass through with the
            # envelope intact.
            data_message_extraction = optional(bool, false)
          }))

          # Append a delimiter to every record. Use to produce newline-delimited
          # JSON (JSON lines) output for query engines and log consumers.
          #
          # Only supported on the extended_s3 destination -- delimiting is an
          # S3-object formatting concern; other destinations frame records
          # natively.
          append_delimiter = optional(object({
            # Delimiter appended to each record. Use "\\n" for newline-delimited
            # output -- the format Athena, Spark, and most log consumers expect.
            delimiter = string
          }))

          # Split multi-record aggregates (e.g., KPL-aggregated or delimited
          # payloads) into individual records before further processing. Required
          # before dynamic partitioning when producers aggregate records.
          #
          # Only supported on the extended_s3 destination -- AWS rejects it for
          # every other destination type at creation.
          record_deaggregation = optional(object({
            # How records are aggregated in the payload:
            # - "JSON" -- concatenated JSON documents (no delimiter needed).
            # - "DELIMITED" -- records separated by a custom delimiter; requires
            #   the delimiter field.
            sub_record_type = string

            # Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
            # newline). Required when sub_record_type is "DELIMITED".
            delimiter = optional(string, "")
          }))
        })), [])
      }))

      # CloudWatch error logging for delivery failures.
      logging = optional(object({
        # Enable CloudWatch error logging for this delivery target.
        enabled = optional(bool, false)

        # CloudWatch Logs log group name where errors are published.
        # Required when enabled is true. Firehose neither creates nor
        # validates this group: CreateDeliveryStream accepts a nonexistent
        # name (live-verified 2026-08-12), and delivery errors are silently
        # dropped until a log group with exactly this name exists -- create
        # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
        # trail you configured here never materializes.
        log_group_name = optional(string, "")

        # CloudWatch Logs log stream name within the log group.
        # Required when enabled is true.
        log_stream_name = optional(string, "")
      }))

      # VPC configuration for delivering to collections reached through a VPC
      # endpoint. ForceNew -- the VPC config cannot be changed after creation.
      # When absent, Firehose delivers over the public internet.
      vpc_config = optional(object({
        # Subnet IDs where Firehose creates ENIs for VPC delivery. Provide at
        # least one subnet. For high availability, use subnets in multiple AZs.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        subnet_ids = optional(list(string), [])

        # Security group IDs applied to the ENIs. Must allow outbound HTTPS (443)
        # traffic to the destination.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        security_group_ids = optional(list(string), [])

        # IAM role ARN for Firehose to manage VPC ENIs. The role must have
        # ec2:CreateNetworkInterface, ec2:DescribeNetworkInterfaces,
        # ec2:DeleteNetworkInterface, and ec2:DescribeVpcs permissions.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string
      }))
    }))

    # HTTP endpoint destination for delivery to any HTTPS endpoint. Supports
    # custom headers, content encoding, and record transformation. Commonly
    # used for third-party integrations (Datadog, New Relic, Sumo Logic).
    # Failed deliveries are backed up to S3.
    http_endpoint = optional(object({
      # HTTPS URL of the destination endpoint. Must start with "https://".
      # Maximum length: 1000 characters.
      #
      # Examples:
      # - "https://http-intake.logs.datadoghq.com/v1/input"
      # - "https://api.honeycomb.io/1/kinesis_events/your-dataset"
      # - "https://my-api.example.com/firehose"
      url = string

      # Human-readable name for the endpoint. Appears in the AWS Console and
      # CloudWatch metrics. Maximum 256 characters.
      name = optional(string, "")

      # Access key for endpoint authentication, sent in the
      # X-Amz-Firehose-Access-Key header. Sensitive -- treated as a secret.
      # Maximum 4096 characters.
      #
      # Mutually exclusive with secrets_manager -- prefer Secrets Manager for
      # production so the key never appears in manifests or IaC state. Endpoints
      # that do not require authentication may omit both.
      access_key = optional(string, "")

      # Source the access key from AWS Secrets Manager instead of access_key.
      # The secret shape is {"api_key": "..."}. Setting this block enables
      # Secrets Manager authentication (ForceNew).
      secrets_manager = optional(object({
        # ARN of the Secrets Manager secret holding the destination credential.
        # An ARN reference resolved by Firehose at delivery time -- never the
        # secret material itself.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        secret_arn = string

        # IAM role ARN granting Firehose permission to read the secret
        # (secretsmanager:GetSecretValue). When absent, the destination's
        # delivery role is used.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = optional(string, "")
      }))

      # IAM role ARN granting Firehose permission to deliver to the endpoint
      # and write to the S3 backup bucket. Optional -- the S3 configuration
      # carries its own delivery role.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = optional(string, "")

      # Buffering hints for HTTP delivery. Default: 300s interval, 5 MiB.
      # Maximum size for HTTP endpoint destinations: 100 MiB.
      buffering = optional(object({
        # Buffer interval in seconds. Firehose flushes when this time elapses since
        # the last flush, even if the buffer size threshold has not been reached.
        #
        # Range: 0-900 seconds. Default varies by destination (typically 300;
        # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
        # values improve batching efficiency and reduce S3 object count.
        #
        # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
        # destination message carries that rule.
        interval_in_seconds = optional(number, 0)

        # Buffer size in MiB. Firehose flushes when the accumulated data reaches
        # this threshold.
        #
        # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
        # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
        # for query engines); smaller buffers provide faster delivery.
        #
        # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
        # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
        size_in_mbs = optional(number, 0)
      }))

      # Retry duration in seconds for failed HTTP deliveries (non-2xx responses
      # or timeouts). Range: 0-7200. Default: 300 seconds.
      retry_duration_in_seconds = optional(number, 0)

      # S3 backup mode. Controls when records are written to S3.
      #
      # Valid values:
      # - "FailedDataOnly" (default) -- only records that fail HTTP delivery
      #   are backed up to S3.
      # - "AllData" -- all records are backed up to S3 in addition to being
      #   sent to the HTTP endpoint.
      s3_backup_mode = optional(string, "")

      # S3 configuration for backing up failed (or all) records. Required.
      s3_config = object({
        # S3 bucket ARN where records are delivered.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket_arn = string

        # IAM role ARN that grants Firehose permission to write to the S3 bucket.
        # The role must have s3:PutObject, s3:AbortMultipartUpload,
        # s3:GetBucketLocation, and s3:ListBucket permissions.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string

        # S3 key prefix prepended to delivered objects. Supports Firehose expression
        # syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").
        prefix = optional(string, "")

        # S3 key prefix for error output. When Firehose cannot deliver or transform
        # a record, it writes to this prefix. Uses the same expression syntax as prefix.
        error_output_prefix = optional(string, "")

        # Compression format for delivered objects. Applied before writing to S3.
        # Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
        # Default: "UNCOMPRESSED".
        compression_format = optional(string, "")

        # Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
        # When absent, S3 uses its default encryption settings (SSE-S3 or bucket
        # default encryption).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_arn = optional(string, "")

        # Buffering hints for S3 delivery.
        buffering = optional(object({
          # Buffer interval in seconds. Firehose flushes when this time elapses since
          # the last flush, even if the buffer size threshold has not been reached.
          #
          # Range: 0-900 seconds. Default varies by destination (typically 300;
          # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
          # values improve batching efficiency and reduce S3 object count.
          #
          # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
          # destination message carries that rule.
          interval_in_seconds = optional(number, 0)

          # Buffer size in MiB. Firehose flushes when the accumulated data reaches
          # this threshold.
          #
          # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
          # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
          # for query engines); smaller buffers provide faster delivery.
          #
          # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
          # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
          size_in_mbs = optional(number, 0)
        }))

        # CloudWatch logging configuration for S3 delivery errors.
        logging = optional(object({
          # Enable CloudWatch error logging for this delivery target.
          enabled = optional(bool, false)

          # CloudWatch Logs log group name where errors are published.
          # Required when enabled is true. Firehose neither creates nor
          # validates this group: CreateDeliveryStream accepts a nonexistent
          # name (live-verified 2026-08-12), and delivery errors are silently
          # dropped until a log group with exactly this name exists -- create
          # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
          # trail you configured here never materializes.
          log_group_name = optional(string, "")

          # CloudWatch Logs log stream name within the log group.
          # Required when enabled is true.
          log_stream_name = optional(string, "")
        }))
      })

      # Record-transformation pipeline applied before HTTP delivery.
      processing = optional(object({
        # Enable the processing pipeline. When true, at least one processor should
        # be configured.
        enabled = optional(bool, false)

        # Ordered list of processors. Each entry configures exactly one processor
        # type; Firehose executes them in order.
        processors = optional(list(object({
          # Invoke an AWS Lambda function to transform records. The function
          # receives batches of records and returns transformed records with a
          # status (Ok, Dropped, ProcessingFailed) per record.
          lambda = optional(object({
            # ARN of the Lambda function that transforms records. May include a
            # version or alias qualifier to pin the deployed transformation.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            lambda_arn = string

            # Buffer size in MiB that Firehose accumulates before invoking Lambda.
            # Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
            # low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
            # destination is Splunk).
            #
            # Smaller buffers invoke Lambda more frequently with smaller batches.
            # Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
            # invocation costs.
            buffer_size_in_mbs = optional(number, 0)

            # Buffer interval in seconds. Firehose invokes Lambda when this interval
            # elapses, even if the buffer size threshold has not been reached.
            # Range: 60-900 seconds. Default: 60 seconds.
            buffer_interval_in_seconds = optional(number, 0)

            # Number of times Firehose retries a failed Lambda invocation before
            # writing the record to the error output prefix.
            # Range: 0-300. Default: 3.
            number_of_retries = optional(number, 0)

            # IAM role ARN Firehose assumes to invoke the Lambda function. When
            # absent, Firehose uses the delivery stream's destination role -- the
            # right choice for almost every pipeline. Set this only when the
            # transformation function must be invoked with a DIFFERENT role than the
            # one that writes to the destination (e.g. the function lives in another
            # account). Note: AWS reports the delivery role back for unset values, and
            # the provider does not store default-valued processor parameters in
            # state -- so set this only to a non-default role, never to the delivery
            # role itself (that would cause perpetual plan diffs).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            role_arn = optional(string, "")
          }))

          # Extract partition keys from JSON records with a JQ expression. Used
          # with Extended S3 dynamic partitioning -- extracted keys are referenced
          # in the S3 prefix as !{partitionKeyFromQuery:<key>}.
          metadata_extraction = optional(object({
            # JQ expression that extracts partition keys from each JSON record.
            # The result must be an object whose keys become partition keys.
            #
            # Example: "{customer_id: .customer_id, event_type: .type}" extracts two
            # keys, referenced in the prefix as
            # "data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".
            query = string

            # JSON parsing engine used to evaluate the query.
            # Valid value: "JQ-1.6" (default -- the only engine AWS supports today).
            json_parsing_engine = optional(string, "")
          }))

          # Decompress GZIP-compressed records before delivery. Typically the first
          # processor when the source sends compressed payloads (e.g., CloudWatch
          # Logs subscription filters).
          decompression = optional(object({
            # Compression format of the incoming records.
            # Valid value: "GZIP" (the only format AWS supports today).
            compression_format = string
          }))

          # Unwrap CloudWatch Logs subscription envelopes into individual log
          # events. Use after a decompression processor when the source is a
          # CloudWatch Logs subscription filter.
          cloudwatch_log_processing = optional(object({
            # When true, extract only the log event message field, discarding the
            # CloudWatch envelope metadata. When false, records pass through with the
            # envelope intact.
            data_message_extraction = optional(bool, false)
          }))

          # Append a delimiter to every record. Use to produce newline-delimited
          # JSON (JSON lines) output for query engines and log consumers.
          #
          # Only supported on the extended_s3 destination -- delimiting is an
          # S3-object formatting concern; other destinations frame records
          # natively.
          append_delimiter = optional(object({
            # Delimiter appended to each record. Use "\\n" for newline-delimited
            # output -- the format Athena, Spark, and most log consumers expect.
            delimiter = string
          }))

          # Split multi-record aggregates (e.g., KPL-aggregated or delimited
          # payloads) into individual records before further processing. Required
          # before dynamic partitioning when producers aggregate records.
          #
          # Only supported on the extended_s3 destination -- AWS rejects it for
          # every other destination type at creation.
          record_deaggregation = optional(object({
            # How records are aggregated in the payload:
            # - "JSON" -- concatenated JSON documents (no delimiter needed).
            # - "DELIMITED" -- records separated by a custom delimiter; requires
            #   the delimiter field.
            sub_record_type = string

            # Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
            # newline). Required when sub_record_type is "DELIMITED".
            delimiter = optional(string, "")
          }))
        })), [])
      }))

      # CloudWatch error logging for HTTP delivery failures.
      logging = optional(object({
        # Enable CloudWatch error logging for this delivery target.
        enabled = optional(bool, false)

        # CloudWatch Logs log group name where errors are published.
        # Required when enabled is true. Firehose neither creates nor
        # validates this group: CreateDeliveryStream accepts a nonexistent
        # name (live-verified 2026-08-12), and delivery errors are silently
        # dropped until a log group with exactly this name exists -- create
        # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
        # trail you configured here never materializes.
        log_group_name = optional(string, "")

        # CloudWatch Logs log stream name within the log group.
        # Required when enabled is true.
        log_stream_name = optional(string, "")
      }))

      # Request configuration for customizing the HTTP request format.
      request_config = optional(object({
        # Content encoding for the HTTP request body.
        # Valid values: "NONE" (default), "GZIP".
        # GZIP reduces payload size but adds CPU overhead.
        content_encoding = optional(string, "")

        # Custom key-value pairs sent as HTTP headers with every request.
        # Use this for endpoint-specific metadata (e.g., dataset name,
        # environment identifier, API version).
        common_attributes = optional(list(object({
          # Header name.
          name = string

          # Header value.
          value = string
        })), [])
      }))
    }))

    # Redshift destination for data warehouse loading. Firehose stages data
    # in S3, then issues a Redshift COPY command to load it. Supports record
    # transformation and optional S3 backup of source records.
    redshift = optional(object({
      # JDBC URL of the Redshift cluster. Format:
      #   "jdbc:redshift://<endpoint>:<port>/<database>"
      # Example: "jdbc:redshift://my-cluster.abcdef.us-east-1.redshift.amazonaws.com:5439/mydb"
      cluster_jdbcurl = string

      # IAM role ARN granting Firehose permission to COPY from S3 to Redshift
      # and write to the S3 staging bucket. Must have:
      # - S3 read access to the staging bucket
      # - Redshift COPY permission
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = string

      # Name of the target Redshift table for the COPY command.
      data_table_name = string

      # Comma-separated list of column names for the COPY command. When set,
      # only the specified columns are loaded. When absent, COPY loads into
      # all columns in table order.
      data_table_columns = optional(string, "")

      # Additional COPY command options (e.g., "JSON 'auto'", "GZIP",
      # "DELIMITER ','", "IGNOREHEADER 1"). Appended to the COPY command.
      copy_options = optional(string, "")

      # Redshift database username. Required together with password when
      # authenticating with plaintext credentials; must be empty when
      # secrets_manager is set.
      username = optional(string, "")

      # Redshift database password. Sensitive -- the value lands in IaC state.
      # Prefer secrets_manager for production, which keeps the credential in
      # Secrets Manager entirely. Required together with username when
      # authenticating with plaintext credentials.
      password = optional(string, "")

      # Source the credentials from AWS Secrets Manager instead of
      # username/password. The secret shape is
      # {"username": "...", "password": "..."}. Setting this block enables
      # Secrets Manager authentication (ForceNew).
      secrets_manager = optional(object({
        # ARN of the Secrets Manager secret holding the destination credential.
        # An ARN reference resolved by Firehose at delivery time -- never the
        # secret material itself.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        secret_arn = string

        # IAM role ARN granting Firehose permission to read the secret
        # (secretsmanager:GetSecretValue). When absent, the destination's
        # delivery role is used.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = optional(string, "")
      }))

      # S3 configuration for the intermediate staging bucket. Firehose writes
      # data to this S3 location, then issues a COPY command to load it into
      # Redshift. This is NOT a backup -- it's the primary data path.
      s3_config = object({
        # S3 bucket ARN where records are delivered.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket_arn = string

        # IAM role ARN that grants Firehose permission to write to the S3 bucket.
        # The role must have s3:PutObject, s3:AbortMultipartUpload,
        # s3:GetBucketLocation, and s3:ListBucket permissions.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string

        # S3 key prefix prepended to delivered objects. Supports Firehose expression
        # syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").
        prefix = optional(string, "")

        # S3 key prefix for error output. When Firehose cannot deliver or transform
        # a record, it writes to this prefix. Uses the same expression syntax as prefix.
        error_output_prefix = optional(string, "")

        # Compression format for delivered objects. Applied before writing to S3.
        # Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
        # Default: "UNCOMPRESSED".
        compression_format = optional(string, "")

        # Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
        # When absent, S3 uses its default encryption settings (SSE-S3 or bucket
        # default encryption).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_arn = optional(string, "")

        # Buffering hints for S3 delivery.
        buffering = optional(object({
          # Buffer interval in seconds. Firehose flushes when this time elapses since
          # the last flush, even if the buffer size threshold has not been reached.
          #
          # Range: 0-900 seconds. Default varies by destination (typically 300;
          # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
          # values improve batching efficiency and reduce S3 object count.
          #
          # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
          # destination message carries that rule.
          interval_in_seconds = optional(number, 0)

          # Buffer size in MiB. Firehose flushes when the accumulated data reaches
          # this threshold.
          #
          # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
          # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
          # for query engines); smaller buffers provide faster delivery.
          #
          # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
          # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
          size_in_mbs = optional(number, 0)
        }))

        # CloudWatch logging configuration for S3 delivery errors.
        logging = optional(object({
          # Enable CloudWatch error logging for this delivery target.
          enabled = optional(bool, false)

          # CloudWatch Logs log group name where errors are published.
          # Required when enabled is true. Firehose neither creates nor
          # validates this group: CreateDeliveryStream accepts a nonexistent
          # name (live-verified 2026-08-12), and delivery errors are silently
          # dropped until a log group with exactly this name exists -- create
          # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
          # trail you configured here never materializes.
          log_group_name = optional(string, "")

          # CloudWatch Logs log stream name within the log group.
          # Required when enabled is true.
          log_stream_name = optional(string, "")
        }))
      })

      # Retry duration in seconds for failed Redshift COPY commands.
      # Range: 0-7200. Default: 3600 seconds (1 hour).
      # Redshift COPY can be slow, so a longer default is appropriate.
      retry_duration_in_seconds = optional(number, 0)

      # S3 backup mode for source records (in addition to the staging S3).
      # When "Enabled", a copy of the original records is written to
      # s3_backup. Useful for auditing and reprocessing.
      #
      # Valid values: "Disabled" (default), "Enabled".
      s3_backup_mode = optional(string, "")

      # S3 configuration for source record backup. Required when
      # s3_backup_mode is "Enabled".
      s3_backup = optional(object({
        # S3 bucket ARN where records are delivered.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket_arn = string

        # IAM role ARN that grants Firehose permission to write to the S3 bucket.
        # The role must have s3:PutObject, s3:AbortMultipartUpload,
        # s3:GetBucketLocation, and s3:ListBucket permissions.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string

        # S3 key prefix prepended to delivered objects. Supports Firehose expression
        # syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").
        prefix = optional(string, "")

        # S3 key prefix for error output. When Firehose cannot deliver or transform
        # a record, it writes to this prefix. Uses the same expression syntax as prefix.
        error_output_prefix = optional(string, "")

        # Compression format for delivered objects. Applied before writing to S3.
        # Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
        # Default: "UNCOMPRESSED".
        compression_format = optional(string, "")

        # Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
        # When absent, S3 uses its default encryption settings (SSE-S3 or bucket
        # default encryption).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_arn = optional(string, "")

        # Buffering hints for S3 delivery.
        buffering = optional(object({
          # Buffer interval in seconds. Firehose flushes when this time elapses since
          # the last flush, even if the buffer size threshold has not been reached.
          #
          # Range: 0-900 seconds. Default varies by destination (typically 300;
          # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
          # values improve batching efficiency and reduce S3 object count.
          #
          # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
          # destination message carries that rule.
          interval_in_seconds = optional(number, 0)

          # Buffer size in MiB. Firehose flushes when the accumulated data reaches
          # this threshold.
          #
          # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
          # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
          # for query engines); smaller buffers provide faster delivery.
          #
          # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
          # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
          size_in_mbs = optional(number, 0)
        }))

        # CloudWatch logging configuration for S3 delivery errors.
        logging = optional(object({
          # Enable CloudWatch error logging for this delivery target.
          enabled = optional(bool, false)

          # CloudWatch Logs log group name where errors are published.
          # Required when enabled is true. Firehose neither creates nor
          # validates this group: CreateDeliveryStream accepts a nonexistent
          # name (live-verified 2026-08-12), and delivery errors are silently
          # dropped until a log group with exactly this name exists -- create
          # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
          # trail you configured here never materializes.
          log_group_name = optional(string, "")

          # CloudWatch Logs log stream name within the log group.
          # Required when enabled is true.
          log_stream_name = optional(string, "")
        }))
      }))

      # Record-transformation pipeline applied before staging to S3.
      processing = optional(object({
        # Enable the processing pipeline. When true, at least one processor should
        # be configured.
        enabled = optional(bool, false)

        # Ordered list of processors. Each entry configures exactly one processor
        # type; Firehose executes them in order.
        processors = optional(list(object({
          # Invoke an AWS Lambda function to transform records. The function
          # receives batches of records and returns transformed records with a
          # status (Ok, Dropped, ProcessingFailed) per record.
          lambda = optional(object({
            # ARN of the Lambda function that transforms records. May include a
            # version or alias qualifier to pin the deployed transformation.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            lambda_arn = string

            # Buffer size in MiB that Firehose accumulates before invoking Lambda.
            # Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
            # low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
            # destination is Splunk).
            #
            # Smaller buffers invoke Lambda more frequently with smaller batches.
            # Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
            # invocation costs.
            buffer_size_in_mbs = optional(number, 0)

            # Buffer interval in seconds. Firehose invokes Lambda when this interval
            # elapses, even if the buffer size threshold has not been reached.
            # Range: 60-900 seconds. Default: 60 seconds.
            buffer_interval_in_seconds = optional(number, 0)

            # Number of times Firehose retries a failed Lambda invocation before
            # writing the record to the error output prefix.
            # Range: 0-300. Default: 3.
            number_of_retries = optional(number, 0)

            # IAM role ARN Firehose assumes to invoke the Lambda function. When
            # absent, Firehose uses the delivery stream's destination role -- the
            # right choice for almost every pipeline. Set this only when the
            # transformation function must be invoked with a DIFFERENT role than the
            # one that writes to the destination (e.g. the function lives in another
            # account). Note: AWS reports the delivery role back for unset values, and
            # the provider does not store default-valued processor parameters in
            # state -- so set this only to a non-default role, never to the delivery
            # role itself (that would cause perpetual plan diffs).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            role_arn = optional(string, "")
          }))

          # Extract partition keys from JSON records with a JQ expression. Used
          # with Extended S3 dynamic partitioning -- extracted keys are referenced
          # in the S3 prefix as !{partitionKeyFromQuery:<key>}.
          metadata_extraction = optional(object({
            # JQ expression that extracts partition keys from each JSON record.
            # The result must be an object whose keys become partition keys.
            #
            # Example: "{customer_id: .customer_id, event_type: .type}" extracts two
            # keys, referenced in the prefix as
            # "data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".
            query = string

            # JSON parsing engine used to evaluate the query.
            # Valid value: "JQ-1.6" (default -- the only engine AWS supports today).
            json_parsing_engine = optional(string, "")
          }))

          # Decompress GZIP-compressed records before delivery. Typically the first
          # processor when the source sends compressed payloads (e.g., CloudWatch
          # Logs subscription filters).
          decompression = optional(object({
            # Compression format of the incoming records.
            # Valid value: "GZIP" (the only format AWS supports today).
            compression_format = string
          }))

          # Unwrap CloudWatch Logs subscription envelopes into individual log
          # events. Use after a decompression processor when the source is a
          # CloudWatch Logs subscription filter.
          cloudwatch_log_processing = optional(object({
            # When true, extract only the log event message field, discarding the
            # CloudWatch envelope metadata. When false, records pass through with the
            # envelope intact.
            data_message_extraction = optional(bool, false)
          }))

          # Append a delimiter to every record. Use to produce newline-delimited
          # JSON (JSON lines) output for query engines and log consumers.
          #
          # Only supported on the extended_s3 destination -- delimiting is an
          # S3-object formatting concern; other destinations frame records
          # natively.
          append_delimiter = optional(object({
            # Delimiter appended to each record. Use "\\n" for newline-delimited
            # output -- the format Athena, Spark, and most log consumers expect.
            delimiter = string
          }))

          # Split multi-record aggregates (e.g., KPL-aggregated or delimited
          # payloads) into individual records before further processing. Required
          # before dynamic partitioning when producers aggregate records.
          #
          # Only supported on the extended_s3 destination -- AWS rejects it for
          # every other destination type at creation.
          record_deaggregation = optional(object({
            # How records are aggregated in the payload:
            # - "JSON" -- concatenated JSON documents (no delimiter needed).
            # - "DELIMITED" -- records separated by a custom delimiter; requires
            #   the delimiter field.
            sub_record_type = string

            # Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
            # newline). Required when sub_record_type is "DELIMITED".
            delimiter = optional(string, "")
          }))
        })), [])
      }))

      # CloudWatch error logging for Redshift COPY failures.
      logging = optional(object({
        # Enable CloudWatch error logging for this delivery target.
        enabled = optional(bool, false)

        # CloudWatch Logs log group name where errors are published.
        # Required when enabled is true. Firehose neither creates nor
        # validates this group: CreateDeliveryStream accepts a nonexistent
        # name (live-verified 2026-08-12), and delivery errors are silently
        # dropped until a log group with exactly this name exists -- create
        # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
        # trail you configured here never materializes.
        log_group_name = optional(string, "")

        # CloudWatch Logs log stream name within the log group.
        # Required when enabled is true.
        log_stream_name = optional(string, "")
      }))
    }))

    # Splunk destination for delivery to a Splunk HTTP Event Collector (HEC)
    # endpoint -- Splunk Cloud, Splunk Enterprise, or Splunk-managed AWS.
    # Failed events are backed up to S3.
    splunk = optional(object({
      # Splunk HTTP Event Collector endpoint URL, including the port.
      # Examples:
      # - "https://http-inputs-mycompany.splunkcloud.com:443"
      # - "https://splunk.example.com:8088"
      hec_endpoint = string

      # HEC endpoint type:
      # - "Raw" (default) -- events are sent to the raw endpoint as-is. Use for
      #   preformatted events (the common case for Firehose delivery).
      # - "Event" -- events are sent to the event endpoint and must be JSON
      #   objects in Splunk's event format.
      hec_endpoint_type = optional(string, "")

      # HEC token that authorizes delivery, minted in Splunk when the HEC input
      # is created. Sensitive -- the value lands in IaC state. Prefer
      # secrets_manager for production, which keeps the token in Secrets
      # Manager entirely.
      hec_token = optional(string, "")

      # Source the HEC token from AWS Secrets Manager instead of hec_token.
      # The secret shape is {"hec_token": "..."}. Setting this block enables
      # Secrets Manager authentication (ForceNew).
      secrets_manager = optional(object({
        # ARN of the Secrets Manager secret holding the destination credential.
        # An ARN reference resolved by Firehose at delivery time -- never the
        # secret material itself.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        secret_arn = string

        # IAM role ARN granting Firehose permission to read the secret
        # (secretsmanager:GetSecretValue). When absent, the destination's
        # delivery role is used.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = optional(string, "")
      }))

      # Time in seconds that Firehose waits for the Splunk indexer
      # acknowledgment after sending data. Unacknowledged data is retried or
      # backed up to S3.
      # Range: 180-600. Default: 180 seconds.
      hec_acknowledgment_timeout_in_seconds = optional(number, 0)

      # Buffering hints for Splunk delivery. Default: 60s interval, 5 MiB.
      # Splunk enforces the tightest limits of any destination:
      # interval 0-60 seconds, size 1-5 MiB.
      buffering = optional(object({
        # Buffer interval in seconds. Firehose flushes when this time elapses since
        # the last flush, even if the buffer size threshold has not been reached.
        #
        # Range: 0-900 seconds. Default varies by destination (typically 300;
        # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
        # values improve batching efficiency and reduce S3 object count.
        #
        # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
        # destination message carries that rule.
        interval_in_seconds = optional(number, 0)

        # Buffer size in MiB. Firehose flushes when the accumulated data reaches
        # this threshold.
        #
        # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
        # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
        # for query engines); smaller buffers provide faster delivery.
        #
        # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
        # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
        size_in_mbs = optional(number, 0)
      }))

      # Retry duration in seconds for failed or unacknowledged HEC deliveries.
      # Range: 0-7200. Default: 3600 seconds (1 hour).
      retry_duration_in_seconds = optional(number, 0)

      # S3 backup mode for events. Controls when records are written to S3.
      #
      # Valid values:
      # - "FailedEventsOnly" (default) -- only events that fail HEC delivery
      #   are backed up to S3.
      # - "AllEvents" -- all events are backed up to S3 in addition to being
      #   sent to Splunk.
      s3_backup_mode = optional(string, "")

      # S3 configuration for backing up failed (or all) events. Required.
      s3_config = object({
        # S3 bucket ARN where records are delivered.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket_arn = string

        # IAM role ARN that grants Firehose permission to write to the S3 bucket.
        # The role must have s3:PutObject, s3:AbortMultipartUpload,
        # s3:GetBucketLocation, and s3:ListBucket permissions.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string

        # S3 key prefix prepended to delivered objects. Supports Firehose expression
        # syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").
        prefix = optional(string, "")

        # S3 key prefix for error output. When Firehose cannot deliver or transform
        # a record, it writes to this prefix. Uses the same expression syntax as prefix.
        error_output_prefix = optional(string, "")

        # Compression format for delivered objects. Applied before writing to S3.
        # Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
        # Default: "UNCOMPRESSED".
        compression_format = optional(string, "")

        # Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
        # When absent, S3 uses its default encryption settings (SSE-S3 or bucket
        # default encryption).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_arn = optional(string, "")

        # Buffering hints for S3 delivery.
        buffering = optional(object({
          # Buffer interval in seconds. Firehose flushes when this time elapses since
          # the last flush, even if the buffer size threshold has not been reached.
          #
          # Range: 0-900 seconds. Default varies by destination (typically 300;
          # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
          # values improve batching efficiency and reduce S3 object count.
          #
          # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
          # destination message carries that rule.
          interval_in_seconds = optional(number, 0)

          # Buffer size in MiB. Firehose flushes when the accumulated data reaches
          # this threshold.
          #
          # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
          # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
          # for query engines); smaller buffers provide faster delivery.
          #
          # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
          # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
          size_in_mbs = optional(number, 0)
        }))

        # CloudWatch logging configuration for S3 delivery errors.
        logging = optional(object({
          # Enable CloudWatch error logging for this delivery target.
          enabled = optional(bool, false)

          # CloudWatch Logs log group name where errors are published.
          # Required when enabled is true. Firehose neither creates nor
          # validates this group: CreateDeliveryStream accepts a nonexistent
          # name (live-verified 2026-08-12), and delivery errors are silently
          # dropped until a log group with exactly this name exists -- create
          # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
          # trail you configured here never materializes.
          log_group_name = optional(string, "")

          # CloudWatch Logs log stream name within the log group.
          # Required when enabled is true.
          log_stream_name = optional(string, "")
        }))
      })

      # Record-transformation pipeline applied before HEC delivery.
      processing = optional(object({
        # Enable the processing pipeline. When true, at least one processor should
        # be configured.
        enabled = optional(bool, false)

        # Ordered list of processors. Each entry configures exactly one processor
        # type; Firehose executes them in order.
        processors = optional(list(object({
          # Invoke an AWS Lambda function to transform records. The function
          # receives batches of records and returns transformed records with a
          # status (Ok, Dropped, ProcessingFailed) per record.
          lambda = optional(object({
            # ARN of the Lambda function that transforms records. May include a
            # version or alias qualifier to pin the deployed transformation.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            lambda_arn = string

            # Buffer size in MiB that Firehose accumulates before invoking Lambda.
            # Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
            # low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
            # destination is Splunk).
            #
            # Smaller buffers invoke Lambda more frequently with smaller batches.
            # Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
            # invocation costs.
            buffer_size_in_mbs = optional(number, 0)

            # Buffer interval in seconds. Firehose invokes Lambda when this interval
            # elapses, even if the buffer size threshold has not been reached.
            # Range: 60-900 seconds. Default: 60 seconds.
            buffer_interval_in_seconds = optional(number, 0)

            # Number of times Firehose retries a failed Lambda invocation before
            # writing the record to the error output prefix.
            # Range: 0-300. Default: 3.
            number_of_retries = optional(number, 0)

            # IAM role ARN Firehose assumes to invoke the Lambda function. When
            # absent, Firehose uses the delivery stream's destination role -- the
            # right choice for almost every pipeline. Set this only when the
            # transformation function must be invoked with a DIFFERENT role than the
            # one that writes to the destination (e.g. the function lives in another
            # account). Note: AWS reports the delivery role back for unset values, and
            # the provider does not store default-valued processor parameters in
            # state -- so set this only to a non-default role, never to the delivery
            # role itself (that would cause perpetual plan diffs).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            role_arn = optional(string, "")
          }))

          # Extract partition keys from JSON records with a JQ expression. Used
          # with Extended S3 dynamic partitioning -- extracted keys are referenced
          # in the S3 prefix as !{partitionKeyFromQuery:<key>}.
          metadata_extraction = optional(object({
            # JQ expression that extracts partition keys from each JSON record.
            # The result must be an object whose keys become partition keys.
            #
            # Example: "{customer_id: .customer_id, event_type: .type}" extracts two
            # keys, referenced in the prefix as
            # "data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".
            query = string

            # JSON parsing engine used to evaluate the query.
            # Valid value: "JQ-1.6" (default -- the only engine AWS supports today).
            json_parsing_engine = optional(string, "")
          }))

          # Decompress GZIP-compressed records before delivery. Typically the first
          # processor when the source sends compressed payloads (e.g., CloudWatch
          # Logs subscription filters).
          decompression = optional(object({
            # Compression format of the incoming records.
            # Valid value: "GZIP" (the only format AWS supports today).
            compression_format = string
          }))

          # Unwrap CloudWatch Logs subscription envelopes into individual log
          # events. Use after a decompression processor when the source is a
          # CloudWatch Logs subscription filter.
          cloudwatch_log_processing = optional(object({
            # When true, extract only the log event message field, discarding the
            # CloudWatch envelope metadata. When false, records pass through with the
            # envelope intact.
            data_message_extraction = optional(bool, false)
          }))

          # Append a delimiter to every record. Use to produce newline-delimited
          # JSON (JSON lines) output for query engines and log consumers.
          #
          # Only supported on the extended_s3 destination -- delimiting is an
          # S3-object formatting concern; other destinations frame records
          # natively.
          append_delimiter = optional(object({
            # Delimiter appended to each record. Use "\\n" for newline-delimited
            # output -- the format Athena, Spark, and most log consumers expect.
            delimiter = string
          }))

          # Split multi-record aggregates (e.g., KPL-aggregated or delimited
          # payloads) into individual records before further processing. Required
          # before dynamic partitioning when producers aggregate records.
          #
          # Only supported on the extended_s3 destination -- AWS rejects it for
          # every other destination type at creation.
          record_deaggregation = optional(object({
            # How records are aggregated in the payload:
            # - "JSON" -- concatenated JSON documents (no delimiter needed).
            # - "DELIMITED" -- records separated by a custom delimiter; requires
            #   the delimiter field.
            sub_record_type = string

            # Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
            # newline). Required when sub_record_type is "DELIMITED".
            delimiter = optional(string, "")
          }))
        })), [])
      }))

      # CloudWatch error logging for Splunk delivery failures.
      logging = optional(object({
        # Enable CloudWatch error logging for this delivery target.
        enabled = optional(bool, false)

        # CloudWatch Logs log group name where errors are published.
        # Required when enabled is true. Firehose neither creates nor
        # validates this group: CreateDeliveryStream accepts a nonexistent
        # name (live-verified 2026-08-12), and delivery errors are silently
        # dropped until a log group with exactly this name exists -- create
        # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
        # trail you configured here never materializes.
        log_group_name = optional(string, "")

        # CloudWatch Logs log stream name within the log group.
        # Required when enabled is true.
        log_stream_name = optional(string, "")
      }))
    }))

    # Snowflake destination for direct streaming into a Snowflake table via
    # Snowpipe Streaming. Authenticates with key-pair credentials (or AWS
    # Secrets Manager) and supports PrivateLink. Failed data is backed up
    # to S3.
    snowflake = optional(object({
      # Snowflake account URL.
      # Format: "https://<account-identifier>.snowflakecomputing.com"
      account_url = string

      # Name of the Snowflake database containing the target table.
      database = string

      # Name of the Snowflake schema containing the target table.
      schema = string

      # Name of the target Snowflake table.
      table = string

      # IAM role ARN granting Firehose permission to write to the S3 backup
      # bucket and read the Secrets Manager secret (when used).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = string

      # Snowflake user that owns the key pair. Required together with
      # private_key when authenticating with inline credentials; must be empty
      # when secrets_manager is set.
      user = optional(string, "")

      # RSA private key for key-pair authentication: the PEM body only, without
      # the "-----BEGIN/END PRIVATE KEY-----" header and footer lines.
      # Sensitive -- the value lands in IaC state. Prefer secrets_manager for
      # production, which keeps the key in Secrets Manager entirely.
      private_key = optional(string, "")

      # Passphrase for an encrypted private key. Only set when the private key
      # is encrypted. Length: 7-255 characters. Sensitive.
      key_passphrase = optional(string, "")

      # Source the credentials from AWS Secrets Manager instead of
      # user/private_key. The secret shape is
      # {"user": "...", "private_key": "...", "key_passphrase": "..."}.
      # Setting this block enables Secrets Manager authentication (ForceNew).
      secrets_manager = optional(object({
        # ARN of the Secrets Manager secret holding the destination credential.
        # An ARN reference resolved by Firehose at delivery time -- never the
        # secret material itself.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        secret_arn = string

        # IAM role ARN granting Firehose permission to read the secret
        # (secretsmanager:GetSecretValue). When absent, the destination's
        # delivery role is used.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = optional(string, "")
      }))

      # How records map onto the target table:
      # - "JSON_MAPPING" (default) -- each top-level JSON key maps to a
      #   same-named table column.
      # - "VARIANT_CONTENT_MAPPING" -- the whole record lands in one VARIANT
      #   column (content_column_name).
      # - "VARIANT_CONTENT_AND_METADATA_MAPPING" -- the record lands in a
      #   VARIANT column and Firehose metadata lands in a second VARIANT column
      #   (metadata_column_name).
      data_loading_option = optional(string, "")

      # Name of the VARIANT column that receives record content. Required for
      # the VARIANT_CONTENT_MAPPING and VARIANT_CONTENT_AND_METADATA_MAPPING
      # loading options.
      content_column_name = optional(string, "")

      # Name of the VARIANT column that receives Firehose metadata. Required
      # for the VARIANT_CONTENT_AND_METADATA_MAPPING loading option.
      metadata_column_name = optional(string, "")

      # Snowflake role to assume for the insert. When absent, the user's
      # default role is used. Setting a dedicated ingestion role with
      # insert-only privileges is the recommended least-privilege posture.
      snowflake_role = optional(string, "")

      # AWS PrivateLink VPCE ID for private connectivity to Snowflake
      # (privatelink account URLs). When absent, Firehose connects over the
      # public internet. Format: "com.amazonaws.vpce.<region>.vpce-svc-<id>".
      private_link_vpce_id = optional(string, "")

      # Buffering hints for Snowflake delivery. Default: 0s interval, 1 MiB --
      # near-real-time ingestion via Snowpipe Streaming. Raise the interval to
      # trade latency for fewer, larger inserts.
      buffering = optional(object({
        # Buffer interval in seconds. Firehose flushes when this time elapses since
        # the last flush, even if the buffer size threshold has not been reached.
        #
        # Range: 0-900 seconds. Default varies by destination (typically 300;
        # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
        # values improve batching efficiency and reduce S3 object count.
        #
        # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
        # destination message carries that rule.
        interval_in_seconds = optional(number, 0)

        # Buffer size in MiB. Firehose flushes when the accumulated data reaches
        # this threshold.
        #
        # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
        # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
        # for query engines); smaller buffers provide faster delivery.
        #
        # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
        # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
        size_in_mbs = optional(number, 0)
      }))

      # Retry duration in seconds for failed Snowflake inserts.
      # Range: 0-7200. Default: 60 seconds.
      retry_duration_in_seconds = optional(number, 0)

      # S3 backup mode. Controls when records are written to S3.
      #
      # Valid values:
      # - "FailedDataOnly" (default) -- only records that fail Snowflake
      #   delivery are backed up to S3.
      # - "AllData" -- all records are backed up to S3 in addition to being
      #   delivered to Snowflake.
      s3_backup_mode = optional(string, "")

      # S3 configuration for backing up failed (or all) records. Required.
      s3_config = object({
        # S3 bucket ARN where records are delivered.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket_arn = string

        # IAM role ARN that grants Firehose permission to write to the S3 bucket.
        # The role must have s3:PutObject, s3:AbortMultipartUpload,
        # s3:GetBucketLocation, and s3:ListBucket permissions.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string

        # S3 key prefix prepended to delivered objects. Supports Firehose expression
        # syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").
        prefix = optional(string, "")

        # S3 key prefix for error output. When Firehose cannot deliver or transform
        # a record, it writes to this prefix. Uses the same expression syntax as prefix.
        error_output_prefix = optional(string, "")

        # Compression format for delivered objects. Applied before writing to S3.
        # Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
        # Default: "UNCOMPRESSED".
        compression_format = optional(string, "")

        # Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
        # When absent, S3 uses its default encryption settings (SSE-S3 or bucket
        # default encryption).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_arn = optional(string, "")

        # Buffering hints for S3 delivery.
        buffering = optional(object({
          # Buffer interval in seconds. Firehose flushes when this time elapses since
          # the last flush, even if the buffer size threshold has not been reached.
          #
          # Range: 0-900 seconds. Default varies by destination (typically 300;
          # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
          # values improve batching efficiency and reduce S3 object count.
          #
          # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
          # destination message carries that rule.
          interval_in_seconds = optional(number, 0)

          # Buffer size in MiB. Firehose flushes when the accumulated data reaches
          # this threshold.
          #
          # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
          # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
          # for query engines); smaller buffers provide faster delivery.
          #
          # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
          # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
          size_in_mbs = optional(number, 0)
        }))

        # CloudWatch logging configuration for S3 delivery errors.
        logging = optional(object({
          # Enable CloudWatch error logging for this delivery target.
          enabled = optional(bool, false)

          # CloudWatch Logs log group name where errors are published.
          # Required when enabled is true. Firehose neither creates nor
          # validates this group: CreateDeliveryStream accepts a nonexistent
          # name (live-verified 2026-08-12), and delivery errors are silently
          # dropped until a log group with exactly this name exists -- create
          # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
          # trail you configured here never materializes.
          log_group_name = optional(string, "")

          # CloudWatch Logs log stream name within the log group.
          # Required when enabled is true.
          log_stream_name = optional(string, "")
        }))
      })

      # Record-transformation pipeline applied before delivery.
      processing = optional(object({
        # Enable the processing pipeline. When true, at least one processor should
        # be configured.
        enabled = optional(bool, false)

        # Ordered list of processors. Each entry configures exactly one processor
        # type; Firehose executes them in order.
        processors = optional(list(object({
          # Invoke an AWS Lambda function to transform records. The function
          # receives batches of records and returns transformed records with a
          # status (Ok, Dropped, ProcessingFailed) per record.
          lambda = optional(object({
            # ARN of the Lambda function that transforms records. May include a
            # version or alias qualifier to pin the deployed transformation.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            lambda_arn = string

            # Buffer size in MiB that Firehose accumulates before invoking Lambda.
            # Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
            # low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
            # destination is Splunk).
            #
            # Smaller buffers invoke Lambda more frequently with smaller batches.
            # Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
            # invocation costs.
            buffer_size_in_mbs = optional(number, 0)

            # Buffer interval in seconds. Firehose invokes Lambda when this interval
            # elapses, even if the buffer size threshold has not been reached.
            # Range: 60-900 seconds. Default: 60 seconds.
            buffer_interval_in_seconds = optional(number, 0)

            # Number of times Firehose retries a failed Lambda invocation before
            # writing the record to the error output prefix.
            # Range: 0-300. Default: 3.
            number_of_retries = optional(number, 0)

            # IAM role ARN Firehose assumes to invoke the Lambda function. When
            # absent, Firehose uses the delivery stream's destination role -- the
            # right choice for almost every pipeline. Set this only when the
            # transformation function must be invoked with a DIFFERENT role than the
            # one that writes to the destination (e.g. the function lives in another
            # account). Note: AWS reports the delivery role back for unset values, and
            # the provider does not store default-valued processor parameters in
            # state -- so set this only to a non-default role, never to the delivery
            # role itself (that would cause perpetual plan diffs).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            role_arn = optional(string, "")
          }))

          # Extract partition keys from JSON records with a JQ expression. Used
          # with Extended S3 dynamic partitioning -- extracted keys are referenced
          # in the S3 prefix as !{partitionKeyFromQuery:<key>}.
          metadata_extraction = optional(object({
            # JQ expression that extracts partition keys from each JSON record.
            # The result must be an object whose keys become partition keys.
            #
            # Example: "{customer_id: .customer_id, event_type: .type}" extracts two
            # keys, referenced in the prefix as
            # "data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".
            query = string

            # JSON parsing engine used to evaluate the query.
            # Valid value: "JQ-1.6" (default -- the only engine AWS supports today).
            json_parsing_engine = optional(string, "")
          }))

          # Decompress GZIP-compressed records before delivery. Typically the first
          # processor when the source sends compressed payloads (e.g., CloudWatch
          # Logs subscription filters).
          decompression = optional(object({
            # Compression format of the incoming records.
            # Valid value: "GZIP" (the only format AWS supports today).
            compression_format = string
          }))

          # Unwrap CloudWatch Logs subscription envelopes into individual log
          # events. Use after a decompression processor when the source is a
          # CloudWatch Logs subscription filter.
          cloudwatch_log_processing = optional(object({
            # When true, extract only the log event message field, discarding the
            # CloudWatch envelope metadata. When false, records pass through with the
            # envelope intact.
            data_message_extraction = optional(bool, false)
          }))

          # Append a delimiter to every record. Use to produce newline-delimited
          # JSON (JSON lines) output for query engines and log consumers.
          #
          # Only supported on the extended_s3 destination -- delimiting is an
          # S3-object formatting concern; other destinations frame records
          # natively.
          append_delimiter = optional(object({
            # Delimiter appended to each record. Use "\\n" for newline-delimited
            # output -- the format Athena, Spark, and most log consumers expect.
            delimiter = string
          }))

          # Split multi-record aggregates (e.g., KPL-aggregated or delimited
          # payloads) into individual records before further processing. Required
          # before dynamic partitioning when producers aggregate records.
          #
          # Only supported on the extended_s3 destination -- AWS rejects it for
          # every other destination type at creation.
          record_deaggregation = optional(object({
            # How records are aggregated in the payload:
            # - "JSON" -- concatenated JSON documents (no delimiter needed).
            # - "DELIMITED" -- records separated by a custom delimiter; requires
            #   the delimiter field.
            sub_record_type = string

            # Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
            # newline). Required when sub_record_type is "DELIMITED".
            delimiter = optional(string, "")
          }))
        })), [])
      }))

      # CloudWatch error logging for Snowflake delivery failures.
      logging = optional(object({
        # Enable CloudWatch error logging for this delivery target.
        enabled = optional(bool, false)

        # CloudWatch Logs log group name where errors are published.
        # Required when enabled is true. Firehose neither creates nor
        # validates this group: CreateDeliveryStream accepts a nonexistent
        # name (live-verified 2026-08-12), and delivery errors are silently
        # dropped until a log group with exactly this name exists -- create
        # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
        # trail you configured here never materializes.
        log_group_name = optional(string, "")

        # CloudWatch Logs log stream name within the log group.
        # Required when enabled is true.
        log_stream_name = optional(string, "")
      }))
    }))

    # Iceberg destination for delivery into Apache Iceberg tables managed by
    # the AWS Glue Data Catalog. Supports routing records to multiple tables
    # and update/delete semantics via unique keys. Failed data is backed up
    # to S3.
    iceberg = optional(object({
      # ARN of the Glue Data Catalog that owns the Iceberg tables. The catalog
      # ARN embeds the owning account:
      #   "arn:aws:glue:<region>:<account-id>:catalog"
      # ForceNew -- changing the catalog replaces the delivery stream.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      catalog_arn = string

      # IAM role ARN granting Firehose permission to write to the Iceberg
      # tables: Glue table read/update and S3 read/write on the warehouse
      # location, plus the S3 backup bucket.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      role_arn = string

      # Destination tables for delivered records -- at least one is required
      # (an Iceberg destination with no table cannot route records; AWS
      # rejects it at apply time, so it is rejected here at validation time).
      # When exactly one table is listed, all records land there. When
      # multiple tables are listed, records must carry routing metadata
      # (produced by a metadata_extraction or Lambda processor) selecting the
      # target table per record.
      #
      # ForceNew -- changing the table routing replaces the delivery stream.
      destination_tables = list(object({
        # Glue Data Catalog database containing the Iceberg table.
        database_name = string

        # Name of the Iceberg table.
        table_name = string

        # S3 key prefix for records that fail delivery to this table. Uses the
        # Firehose expression syntax.
        s3_error_output_prefix = optional(string, "")

        # Columns that uniquely identify a row, enabling update/delete semantics:
        # an incoming record whose unique-key values match an existing row
        # updates it instead of appending. Leave empty (with append_only) for
        # pure append workloads.
        unique_keys = optional(list(string), [])
      }))

      # Append-only mode. When true, Firehose only appends new snapshots --
      # update/delete semantics via unique keys are disabled. ForceNew.
      append_only = optional(bool, false)

      # Buffering hints for Iceberg delivery. Default: 300s interval, 5 MiB.
      buffering = optional(object({
        # Buffer interval in seconds. Firehose flushes when this time elapses since
        # the last flush, even if the buffer size threshold has not been reached.
        #
        # Range: 0-900 seconds. Default varies by destination (typically 300;
        # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
        # values improve batching efficiency and reduce S3 object count.
        #
        # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
        # destination message carries that rule.
        interval_in_seconds = optional(number, 0)

        # Buffer size in MiB. Firehose flushes when the accumulated data reaches
        # this threshold.
        #
        # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
        # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
        # for query engines); smaller buffers provide faster delivery.
        #
        # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
        # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
        size_in_mbs = optional(number, 0)
      }))

      # Retry duration in seconds for failed Iceberg commits.
      # Range: 0-7200. Default: 300 seconds.
      retry_duration_in_seconds = optional(number, 0)

      # S3 backup mode. Controls when records are written to S3.
      #
      # Valid values:
      # - "FailedDataOnly" (default) -- only records that fail Iceberg delivery
      #   are backed up to S3.
      # - "AllData" -- all records are backed up to S3 in addition to being
      #   committed to Iceberg.
      s3_backup_mode = optional(string, "")

      # S3 configuration for backing up failed (or all) records. Required.
      s3_config = object({
        # S3 bucket ARN where records are delivered.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket_arn = string

        # IAM role ARN that grants Firehose permission to write to the S3 bucket.
        # The role must have s3:PutObject, s3:AbortMultipartUpload,
        # s3:GetBucketLocation, and s3:ListBucket permissions.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        role_arn = string

        # S3 key prefix prepended to delivered objects. Supports Firehose expression
        # syntax for dynamic prefixes (e.g., "errors/year=!{timestamp:yyyy}/").
        prefix = optional(string, "")

        # S3 key prefix for error output. When Firehose cannot deliver or transform
        # a record, it writes to this prefix. Uses the same expression syntax as prefix.
        error_output_prefix = optional(string, "")

        # Compression format for delivered objects. Applied before writing to S3.
        # Valid values: "UNCOMPRESSED", "GZIP", "ZIP", "Snappy", "HADOOP_SNAPPY".
        # Default: "UNCOMPRESSED".
        compression_format = optional(string, "")

        # Customer-managed KMS key ARN for S3 server-side encryption (SSE-KMS).
        # When absent, S3 uses its default encryption settings (SSE-S3 or bucket
        # default encryption).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        kms_key_arn = optional(string, "")

        # Buffering hints for S3 delivery.
        buffering = optional(object({
          # Buffer interval in seconds. Firehose flushes when this time elapses since
          # the last flush, even if the buffer size threshold has not been reached.
          #
          # Range: 0-900 seconds. Default varies by destination (typically 300;
          # Splunk 60, Snowflake 0). Lower values reduce delivery latency; higher
          # values improve batching efficiency and reduce S3 object count.
          #
          # Some destinations enforce a tighter maximum (Splunk: 60s) -- the
          # destination message carries that rule.
          interval_in_seconds = optional(number, 0)

          # Buffer size in MiB. Firehose flushes when the accumulated data reaches
          # this threshold.
          #
          # Range: 1-128 MiB. Default varies by destination (typically 5 MiB;
          # Snowflake 1 MiB). Larger buffers produce fewer, larger objects (better
          # for query engines); smaller buffers provide faster delivery.
          #
          # Some destinations enforce a tighter maximum (OpenSearch/HTTP endpoint:
          # 100 MiB, Splunk: 5 MiB) -- the destination message carries that rule.
          size_in_mbs = optional(number, 0)
        }))

        # CloudWatch logging configuration for S3 delivery errors.
        logging = optional(object({
          # Enable CloudWatch error logging for this delivery target.
          enabled = optional(bool, false)

          # CloudWatch Logs log group name where errors are published.
          # Required when enabled is true. Firehose neither creates nor
          # validates this group: CreateDeliveryStream accepts a nonexistent
          # name (live-verified 2026-08-12), and delivery errors are silently
          # dropped until a log group with exactly this name exists -- create
          # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
          # trail you configured here never materializes.
          log_group_name = optional(string, "")

          # CloudWatch Logs log stream name within the log group.
          # Required when enabled is true.
          log_stream_name = optional(string, "")
        }))
      })

      # Record-transformation pipeline applied before the Iceberg commit. Use a
      # metadata_extraction processor to produce per-record table routing when
      # multiple destination tables are configured.
      processing = optional(object({
        # Enable the processing pipeline. When true, at least one processor should
        # be configured.
        enabled = optional(bool, false)

        # Ordered list of processors. Each entry configures exactly one processor
        # type; Firehose executes them in order.
        processors = optional(list(object({
          # Invoke an AWS Lambda function to transform records. The function
          # receives batches of records and returns transformed records with a
          # status (Ok, Dropped, ProcessingFailed) per record.
          lambda = optional(object({
            # ARN of the Lambda function that transforms records. May include a
            # version or alias qualifier to pin the deployed transformation.
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            lambda_arn = string

            # Buffer size in MiB that Firehose accumulates before invoking Lambda.
            # Range: 0.2-3 MiB (fractional values are AWS-legal -- e.g. 0.5 for
            # low-latency, small-batch invocation). Default: 1 MiB (256 KiB when the
            # destination is Splunk).
            #
            # Smaller buffers invoke Lambda more frequently with smaller batches.
            # Larger buffers (up to 3 MiB) are more efficient and reduce Lambda
            # invocation costs.
            buffer_size_in_mbs = optional(number, 0)

            # Buffer interval in seconds. Firehose invokes Lambda when this interval
            # elapses, even if the buffer size threshold has not been reached.
            # Range: 60-900 seconds. Default: 60 seconds.
            buffer_interval_in_seconds = optional(number, 0)

            # Number of times Firehose retries a failed Lambda invocation before
            # writing the record to the error output prefix.
            # Range: 0-300. Default: 3.
            number_of_retries = optional(number, 0)

            # IAM role ARN Firehose assumes to invoke the Lambda function. When
            # absent, Firehose uses the delivery stream's destination role -- the
            # right choice for almost every pipeline. Set this only when the
            # transformation function must be invoked with a DIFFERENT role than the
            # one that writes to the destination (e.g. the function lives in another
            # account). Note: AWS reports the delivery role back for unset values, and
            # the provider does not store default-valued processor parameters in
            # state -- so set this only to a non-default role, never to the delivery
            # role itself (that would cause perpetual plan diffs).
            # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
            role_arn = optional(string, "")
          }))

          # Extract partition keys from JSON records with a JQ expression. Used
          # with Extended S3 dynamic partitioning -- extracted keys are referenced
          # in the S3 prefix as !{partitionKeyFromQuery:<key>}.
          metadata_extraction = optional(object({
            # JQ expression that extracts partition keys from each JSON record.
            # The result must be an object whose keys become partition keys.
            #
            # Example: "{customer_id: .customer_id, event_type: .type}" extracts two
            # keys, referenced in the prefix as
            # "data/customer=!{partitionKeyFromQuery:customer_id}/type=!{partitionKeyFromQuery:event_type}/".
            query = string

            # JSON parsing engine used to evaluate the query.
            # Valid value: "JQ-1.6" (default -- the only engine AWS supports today).
            json_parsing_engine = optional(string, "")
          }))

          # Decompress GZIP-compressed records before delivery. Typically the first
          # processor when the source sends compressed payloads (e.g., CloudWatch
          # Logs subscription filters).
          decompression = optional(object({
            # Compression format of the incoming records.
            # Valid value: "GZIP" (the only format AWS supports today).
            compression_format = string
          }))

          # Unwrap CloudWatch Logs subscription envelopes into individual log
          # events. Use after a decompression processor when the source is a
          # CloudWatch Logs subscription filter.
          cloudwatch_log_processing = optional(object({
            # When true, extract only the log event message field, discarding the
            # CloudWatch envelope metadata. When false, records pass through with the
            # envelope intact.
            data_message_extraction = optional(bool, false)
          }))

          # Append a delimiter to every record. Use to produce newline-delimited
          # JSON (JSON lines) output for query engines and log consumers.
          #
          # Only supported on the extended_s3 destination -- delimiting is an
          # S3-object formatting concern; other destinations frame records
          # natively.
          append_delimiter = optional(object({
            # Delimiter appended to each record. Use "\\n" for newline-delimited
            # output -- the format Athena, Spark, and most log consumers expect.
            delimiter = string
          }))

          # Split multi-record aggregates (e.g., KPL-aggregated or delimited
          # payloads) into individual records before further processing. Required
          # before dynamic partitioning when producers aggregate records.
          #
          # Only supported on the extended_s3 destination -- AWS rejects it for
          # every other destination type at creation.
          record_deaggregation = optional(object({
            # How records are aggregated in the payload:
            # - "JSON" -- concatenated JSON documents (no delimiter needed).
            # - "DELIMITED" -- records separated by a custom delimiter; requires
            #   the delimiter field.
            sub_record_type = string

            # Delimiter separating sub-records, base64-encoded (e.g., "Cg==" for a
            # newline). Required when sub_record_type is "DELIMITED".
            delimiter = optional(string, "")
          }))
        })), [])
      }))

      # CloudWatch error logging for Iceberg delivery failures.
      logging = optional(object({
        # Enable CloudWatch error logging for this delivery target.
        enabled = optional(bool, false)

        # CloudWatch Logs log group name where errors are published.
        # Required when enabled is true. Firehose neither creates nor
        # validates this group: CreateDeliveryStream accepts a nonexistent
        # name (live-verified 2026-08-12), and delivery errors are silently
        # dropped until a log group with exactly this name exists -- create
        # it yourself (e.g. an AwsCloudwatchLogGroup resource) or the error
        # trail you configured here never materializes.
        log_group_name = optional(string, "")

        # CloudWatch Logs log stream name within the log group.
        # Required when enabled is true.
        log_stream_name = optional(string, "")
      }))
    }))
  })
}
