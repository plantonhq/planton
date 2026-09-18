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
  description = "GcpPubSubSubscription specification"
  type = object({
    # GCP project where the subscription will be created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Pub/Sub subscription.
    # Must be 3-255 characters, start with a letter, and contain only letters,
    # numbers, hyphens, underscores, periods, tildes, plus signs, and percent
    # signs. Names beginning with "goog" are reserved by Google and rejected
    # at create time. Immutable after creation.
    subscription_name = string

    # The topic from which this subscription receives messages.
    # Format: projects/{project}/topics/{name} or just the topic name if
    # the topic is in the same project as the subscription.
    # Immutable after creation.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    topic = string

    # Maximum time (in seconds) after a subscriber receives a message before the
    # subscriber should acknowledge the message. After the deadline expires,
    # the message is redelivered.
    # Range: 10 to 600 seconds. Defaults to 10 seconds.
    ack_deadline_seconds = optional(number, 0)

    # How long to retain unacknowledged messages in the subscription's backlog.
    # If retain_acked_messages is true, this also controls retention of acknowledged
    # messages and determines how far back a seek operation can go.
    # Format: duration string (e.g., "604800s" for 7 days).
    # Range: 600s (10 minutes) to 2678400s (31 days). Default: 604800s (7 days).
    message_retention_duration = optional(string, "")

    # When true, acknowledged messages are retained in the backlog until they
    # fall out of the message_retention_duration window. Enables replay via seek.
    retain_acked_messages = optional(bool, false)

    # Expiration policy for the subscription. Controls automatic deletion of
    # inactive subscriptions. If not set, GCP defaults to 31 days TTL.
    expiration_policy = optional(object({
      # Duration after which the subscription expires if inactive.
      # Format: duration string (e.g., "2592000s" for 30 days).
      # Minimum: 86400s (1 day). Set to "" for a subscription that never expires.
      ttl = optional(string, "")
    }))

    # Message attribute filter expression. Only messages matching the filter are
    # delivered; non-matching messages are automatically acknowledged.
    # Maximum length: 256 bytes. Immutable after creation.
    filter = optional(string, "")

    # When true, messages with the same ordering key are delivered to subscribers
    # in the order they were published. Immutable after creation.
    enable_message_ordering = optional(bool, false)

    # When true, Pub/Sub guarantees that a message is not resent before its
    # acknowledgement deadline expires. An acknowledged message will not be
    # resent. Note: subscribers may still receive duplicates if the publisher
    # sends the same message multiple times.
    enable_exactly_once_delivery = optional(bool, false)

    # Dead-letter policy. Messages that cannot be processed after repeated
    # delivery attempts are forwarded to the configured dead-letter topic.
    dead_letter_policy = optional(object({
      # The topic to which dead-letter messages are published.
      # Format: projects/{project}/topics/{topic}
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      dead_letter_topic = optional(string, "")

      # Maximum number of delivery attempts before a message is dead-lettered.
      # A delivery attempt is counted as 1 + (NACKs + ack deadline exceeded events).
      # Range: 5 to 100. Defaults to 5 when set to 0.
      max_delivery_attempts = optional(number, 0)
    }))

    # Retry policy. Controls the backoff between consecutive delivery attempts
    # after a NACK or ack deadline exceeded event.
    retry_policy = optional(object({
      # Minimum delay between consecutive delivery attempts of a given message.
      # Format: duration string (e.g., "10s").
      # Range: 0s to 600s. Defaults to 10s.
      minimum_backoff = optional(string, "")

      # Maximum delay between consecutive delivery attempts of a given message.
      # Format: duration string (e.g., "600s").
      # Range: 0s to 600s. Defaults to 600s.
      maximum_backoff = optional(string, "")
    }))

    # Push delivery configuration. When set, Pub/Sub sends messages as HTTP POST
    # requests to the configured endpoint. Mutually exclusive with bigquery_config
    # and cloud_storage_config.
    push_config = optional(object({
      # URL to which Pub/Sub pushes messages. Must use HTTPS.
      # Accepts a literal URL or a reference to a GcpCloudRun service — pushing to
      # a Cloud Run service in the same environment is the canonical serverless
      # consumer pattern, and the service URL contains a generated suffix that can
      # only be known by reading the deployed service's output. Pair a Cloud Run
      # push endpoint with oidc_token (the service account must hold run.invoker)
      # unless the service allows unauthenticated invocations.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      push_endpoint = string

      # Endpoint configuration attributes. The supported attribute is "x-goog-version"
      # which controls the push message format ("v1beta1" or "v1").
      attributes = optional(map(string), {})

      # OIDC token configuration for authenticating push requests.
      oidc_token = optional(object({
        # Service account used to generate the OIDC token. Accepts a literal
        # email or a reference to a GcpServiceAccount resource.
        # The caller must have iam.serviceAccounts.actAs permission on this account.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account_email = string

        # Audience claim for the OIDC token. Identifies the intended recipient.
        # Defaults to the push endpoint URL if not specified.
        audience = optional(string, "")
      }))

      # The message payload is sent unwrapped (no Pub/Sub envelope).
      no_wrapper = optional(object({
        # When true, Pub/Sub message metadata is written as HTTP headers
        # (x-goog-pubsub-<key>:<value>) and message attributes as plain headers.
        write_metadata = optional(bool, false)
      }))

      # The message is wrapped in the standard Pub/Sub envelope.
      pubsub_wrapper = optional(object({}))
    }))

    # BigQuery delivery configuration. When set, Pub/Sub writes messages directly
    # to a BigQuery table. Mutually exclusive with push_config and
    # cloud_storage_config.
    bigquery_config = optional(object({
      # The BigQuery table to write messages to.
      # Format: {project_id}.{dataset_id}.{table_id}
      # Accepts a literal or a reference to a GcpBigQueryTable resource (its
      # qualified_name output is exactly this dotted form).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      table = string

      # When true, use the Pub/Sub topic's schema to map message fields to BigQuery columns.
      # Only one of use_topic_schema and use_table_schema can be true.
      use_topic_schema = optional(bool, false)

      # When true, use the BigQuery table's schema to determine which message fields
      # to write. Only one of use_topic_schema and use_table_schema can be true.
      use_table_schema = optional(bool, false)

      # When true (and use_topic_schema or use_table_schema is true), message fields
      # not present in the BigQuery table schema are silently dropped. When false,
      # messages with extra fields are not written and remain in the backlog.
      drop_unknown_fields = optional(bool, false)

      # When true, the subscription name, messageId, publishTime, attributes, and
      # orderingKey are written to additional columns in the BigQuery table.
      write_metadata = optional(bool, false)

      # Service account to use for writing to BigQuery. Accepts a literal
      # email or a reference to a GcpServiceAccount resource. Defaults to the
      # Pub/Sub service agent if not specified.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account_email = optional(string, "")
    }))

    # Cloud Storage delivery configuration. When set, Pub/Sub writes messages to
    # Cloud Storage objects in batches. Mutually exclusive with push_config and
    # bigquery_config.
    cloud_storage_config = optional(object({
      # The Cloud Storage bucket to write messages to (without "gs://" prefix).
      # The bucket must already exist.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      bucket = string

      # Prefix for Cloud Storage filenames.
      filename_prefix = optional(string, "")

      # Suffix for Cloud Storage filenames. Must not end in "/".
      filename_suffix = optional(string, "")

      # Format string for datetime in Cloud Storage filenames.
      filename_datetime_format = optional(string, "")

      # Maximum bytes per Cloud Storage file before a new file is created.
      # Range: 1024 (1 KB) to 10737418240 (10 GiB).
      max_bytes = optional(number, 0)

      # Maximum duration before a new Cloud Storage file is created.
      # Format: duration string (e.g., "300s").
      # Range: 60s (1 minute) to 600s (10 minutes). Default: 300s (5 minutes).
      # Must not exceed the subscription's ack_deadline_seconds.
      max_duration = optional(string, "")

      # Maximum number of messages per Cloud Storage file. Minimum: 1000.
      max_messages = optional(number, 0)

      # Messages are written as Avro records, optionally using the topic schema
      # and carrying message metadata.
      avro_config = optional(object({
        # When true, serialize output using the topic schema.
        use_topic_schema = optional(bool, false)

        # When true, include subscription name, messageId, publishTime, attributes,
        # and orderingKey as additional fields in the Avro output.
        write_metadata = optional(bool, false)
      }))

      # Messages are written as raw text.
      text_config = optional(object({}))

      # Service account to use for writing to Cloud Storage. Accepts a literal
      # email or a reference to a GcpServiceAccount resource. Defaults to the
      # Pub/Sub service agent if not specified.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      service_account_email = optional(string, "")
    }))

    # User-defined labels attached to the subscription, for cost attribution
    # and fleet queries. Merged with Planton's platform labels (which win on
    # key conflicts).
    labels = optional(map(string), {})

    # Ordered pipeline of transforms applied to every message before
    # delivery to this subscription — reshape payloads for this consumer
    # without changing what other subscriptions on the topic see.
    # Transforms run in list order; a transform returning null drops the
    # message.
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

    # Resource Manager tags bound to the subscription for org-policy and
    # IAM conditions. Keys in the form "tagKeys/{id}", values
    # "tagValues/{id}". Create-time only: changing them later replaces the
    # subscription (its backlog is lost — plan tag changes deliberately).
    resource_manager_tags = optional(map(string), {})

    # Deletion policy for the subscription — what happens when this
    # resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the subscription is deleted; its unacknowledged
    #                backlog is lost immediately
    #   "PREVENT" -- destroy FAILS; protects a consumer whose backlog
    #                must never be silently dropped
    #   "ABANDON" -- the subscription is removed from management but keeps
    #                accumulating and serving messages in GCP
    deletion_policy = optional(string, "")
  })
}
