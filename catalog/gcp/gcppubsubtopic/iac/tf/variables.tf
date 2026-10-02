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
  description = "GcpPubSubTopic specification"
  type = object({
    # GCP project where the topic will be created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Pub/Sub topic.
    # Must be 3-255 characters, start with a letter, and contain only letters,
    # numbers, hyphens, underscores, periods, tildes, plus signs, and percent
    # signs. Names beginning with "goog" are reserved by Google and rejected
    # at create time. Immutable after creation.
    topic_name = string

    # Cloud KMS key for encrypting messages at rest (CMEK).
    # Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
    # The Pub/Sub service account must have roles/cloudkms.cryptoKeyEncrypterDecrypter
    # on this key. If not set, messages are encrypted with Google-managed keys.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    kms_key_name = optional(string, "")

    # Minimum duration to retain a message after it is published to the topic.
    # When set, messages published in the last message_retention_duration are always
    # available to subscribers, and any attached subscription can seek to a timestamp
    # within the retention window.
    # Format: duration string (e.g., "604800s" for 7 days).
    # Range: 600s (10 minutes) to 2678400s (31 days).
    # If not set, message retention is controlled by individual subscriptions.
    message_retention_duration = optional(string, "")

    # Policy constraining the set of GCP regions where messages may be stored.
    # When not set, no regional constraints are applied.
    message_storage_policy = optional(object({
      # A list of GCP region IDs where messages may be persisted in storage.
      # Messages published by publishers running in non-allowed regions will be
      # routed for storage in one of the allowed regions.
      # Must contain at least one region when the policy is specified.
      allowed_persistence_regions = list(string)

      # When true, allowed_persistence_regions is also used to enforce in-transit
      # guarantees for messages. Pub/Sub will fail publish operations on this topic
      # and subscribe operations on any subscription attached to this topic in any
      # region not listed in allowed_persistence_regions.
      enforce_in_transit = optional(bool, false)
    }))

    # Schema validation settings for messages published to the topic.
    # When set, all published messages are validated against the specified schema.
    schema_settings = optional(object({
      # The Pub/Sub schema that published messages must conform to.
      # Accepts a literal fully qualified path (projects/{project}/schemas/{schema})
      # or a reference to a GcpPubSubSchema resource. If the schema is later
      # deleted, the topic validates against the "_deleted-schema_" sentinel and
      # publishes fail — destroy topics (or recreate them without validation)
      # before destroying the schema they reference.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      schema = string

      # The encoding of messages validated against the schema.
      # Valid values: "JSON" or "BINARY".
      encoding = optional(string, "")

      # The minimum (inclusive) schema revision accepted for validating
      # messages. When empty, any revision created before last_revision_id
      # (or any revision at all, if that is also empty) is accepted.
      # Accepts a literal revision ID or a reference to a GcpPubSubSchema
      # resource — its revision_id output is the revision the schema's
      # current definition committed, so referencing it pins the topic to
      # the contract this deploy produced.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      first_revision_id = optional(string, "")

      # The maximum (inclusive) schema revision accepted for validating
      # messages. When empty, any revision created after first_revision_id
      # (or any revision at all, if that is also empty) is accepted. Pinning
      # BOTH bounds to the same revision freezes the contract exactly.
      # Accepts a literal revision ID or a reference to a GcpPubSubSchema
      # resource (its revision_id output).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      last_revision_id = optional(string, "")
    }))

    # Settings for ingesting data from external sources into this topic.
    # Supports AWS Kinesis, AWS MSK, Azure Event Hubs, Cloud Storage, and
    # Confluent Cloud. Typically one data source is configured per topic.
    ingestion_data_source_settings = optional(object({
      # Ingest from Amazon Kinesis Data Streams.
      aws_kinesis = optional(object({
        # The ARN of the Kinesis data stream to ingest from.
        stream_arn = string

        # The ARN of the Kinesis consumer to use for Enhanced Fan-Out delivery.
        consumer_arn = string

        # The ARN of the AWS IAM role used for cross-account access to the Kinesis stream.
        aws_role_arn = string

        # The GCP service account used for Federated Identity authentication with AWS.
        # Accepts a literal email or a reference to a GcpServiceAccount resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        gcp_service_account = string
      }))

      # Ingest from Amazon Managed Streaming for Apache Kafka (MSK).
      aws_msk = optional(object({
        # The ARN of the MSK cluster to ingest from.
        cluster_arn = string

        # The name of the Kafka topic in MSK to ingest from.
        topic = string

        # The ARN of the AWS IAM role used for cross-account access to the MSK cluster.
        aws_role_arn = string

        # The GCP service account used for Federated Identity authentication with AWS.
        # Accepts a literal email or a reference to a GcpServiceAccount resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        gcp_service_account = string
      }))

      # Ingest from Azure Event Hubs.
      azure_event_hubs = optional(object({
        # The Azure resource group containing the Event Hubs namespace.
        resource_group = optional(string, "")

        # The Azure Event Hubs namespace.
        namespace = optional(string, "")

        # The name of the Event Hub to ingest from.
        event_hub = optional(string, "")

        # The Azure AD application client ID for authentication.
        client_id = optional(string, "")

        # The Azure AD tenant ID.
        tenant_id = optional(string, "")

        # The Azure subscription ID.
        subscription_id = optional(string, "")

        # The GCP service account used for Federated Identity authentication with Azure.
        # Accepts a literal email or a reference to a GcpServiceAccount resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        gcp_service_account = optional(string, "")
      }))

      # Ingest from Google Cloud Storage.
      cloud_storage = optional(object({
        # The name of the Cloud Storage bucket to ingest from (without "gs://" prefix).
        # See: https://cloud.google.com/storage/docs/buckets#naming
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        bucket = string

        # Glob pattern used to match objects that will be ingested.
        # If unset, all objects in the bucket will be ingested.
        match_glob = optional(string, "")

        # Only ingest objects with a creation timestamp equal to or later than this value.
        # Format: RFC 3339 (e.g., "2024-01-01T00:00:00Z").
        # If unset, all objects are eligible for ingestion regardless of creation time.
        minimum_object_create_time = optional(string, "")

        # Read Cloud Storage data in Avro binary format. The bytes of each object
        # are set to the data field of a Pub/Sub message.
        avro_format = optional(object({}))

        # Read Cloud Storage data written via Cloud Storage subscriptions.
        # Restores the data and attributes of the originally exported Pub/Sub messages.
        pubsub_avro_format = optional(object({}))

        # Read Cloud Storage data in text format. Each line of text (as defined by
        # the delimiter) becomes the data field of a Pub/Sub message.
        text_format = optional(object({
          # The line delimiter. Defaults to newline ("\n") when not set.
          delimiter = optional(string, "")
        }))
      }))

      # Ingest from Confluent Cloud.
      confluent_cloud = optional(object({
        # The Confluent Cloud bootstrap server address. Format: host:port.
        bootstrap_server = string

        # The name of the Confluent Cloud topic to ingest from.
        topic = string

        # The Workload Identity Pool ID used for Federated Identity authentication
        # with Confluent Cloud.
        identity_pool_id = string

        # The GCP service account used for Federated Identity authentication
        # with Confluent Cloud. Accepts a literal email or a reference to a
        # GcpServiceAccount resource.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        gcp_service_account = string

        # The Confluent Cloud cluster ID. Optional.
        cluster_id = optional(string, "")
      }))

      # Platform logging settings for the ingestion pipeline.
      platform_logs_settings = optional(object({
        # The minimum severity level of platform logs that will be written.
        # Valid values: "DISABLED", "DEBUG", "INFO", "WARNING", "ERROR".
        # If unset, no platform logs will be generated.
        severity = optional(string, "")
      }))
    }))

    # User-defined labels attached to the topic, for cost attribution and
    # fleet queries. Merged with Planton's platform labels (which win on
    # key conflicts).
    labels = optional(map(string), {})

    # Ordered pipeline of transforms applied to every published message
    # before it is stored — redact, normalize, or filter at the topic
    # boundary instead of in every subscriber. Transforms run in list
    # order; a transform returning null drops the message.
    message_transforms = optional(list(object({
      # A JavaScript user-defined function transform.
      javascript_udf = optional(object({
        # Name of the JavaScript function to invoke from the code below.
        # Must be unique across all transforms on the resource.
        function_name = string

        # The JavaScript source code defining the function. The function
        # signature is (message, metadata) => message; return null or undefined
        # to drop the message.
        code = string
      }))

      # When true, this transform is kept in the pipeline definition but not
      # applied — the staging lever for rolling a transform in or out without
      # losing its position in the ordered list.
      disabled = optional(bool, false)

      # An AI inference transform backed by a Vertex AI model endpoint.
      ai_inference = optional(object({
        # The Vertex AI model endpoint inference requests are sent to. Accepts a
        # literal path — projects/{project}/locations/{location}/endpoints/{endpoint}
        # for a dedicated endpoint, or
        # projects/{project}/locations/{location}/publishers/{publisher}/models/{model}
        # for a publisher model — or a reference to a GcpVertexAiEndpoint resource
        # (its endpoint_id output is exactly the dedicated-endpoint form).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        endpoint = string

        # The service account used to make prediction requests against the
        # endpoint (it needs Vertex AI invocation permission on the model).
        # Accepts a literal email or a reference to a GcpServiceAccount resource.
        # Defaults to the Pub/Sub service agent if not specified.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account_email = optional(string, "")

        # Configuration for making inferences using arbitrary JSON payloads
        # (rather than a model-specific request schema).
        unstructured_inference = optional(object({
          # A parameters object included in each inference request (e.g. model
          # temperature or system-prompt knobs the endpoint understands). Combined
          # with the message data to form the request body.
          parameters = optional(map(string), {})
        }))
      }))
    })), [])

    # Resource Manager tags bound to the topic for org-policy and IAM
    # conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
    # Create-time only: changing them later replaces the topic (and detaches
    # every subscription with it — plan tag changes deliberately).
    resource_manager_tags = optional(map(string), {})

    # Deletion policy for the topic — what happens when this resource is
    # destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the topic is deleted. Its subscriptions survive but
    #                receive no new messages and drain to empty
    #   "PREVENT" -- destroy FAILS; protects the topic an event pipeline
    #                publishes into
    #   "ABANDON" -- the topic is removed from management but left serving
    #                in GCP (publishers and subscriptions keep working)
    deletion_policy = optional(string, "")
  })
}
