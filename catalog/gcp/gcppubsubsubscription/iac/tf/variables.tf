variable "metadata" {
  description = "Metadata for the resource, including name and labels"
  type = object({
    name    = string,
    id      = optional(string),
    org     = optional(string),
    env     = optional(string),
    labels  = optional(map(string)),
    tags    = optional(list(string)),
    version = optional(object({ id = string, message = string }))
  })
}

variable "spec" {
  description = "Specification for the GCP Pub/Sub subscription"
  type = object({
    # StringValueOrRef fields arrive from the proto→tfvars converter as
    # plain strings (already resolved), never as object({value}). If
    # project_id is empty, the provider's default project is used
    # (see locals.tf).
    project_id = optional(string, "")

    # Subscription name (the GCP resource name). Immutable (ForceNew).
    subscription_name = string

    # The parent topic (resolved from a GcpPubSubTopic reference to the
    # fully qualified projects/{project}/topics/{name} path). Immutable
    # (ForceNew) — repointing replaces the subscription.
    topic = string

    # 0 accepts the API default (10s); otherwise 10-600 seconds.
    ack_deadline_seconds = optional(number, 0)

    # Backlog retention window (e.g. "604800s" for 7 days).
    message_retention_duration = optional(string, "")

    retain_acked_messages = optional(bool, false)

    # The attribute filter. Immutable (ForceNew).
    filter = optional(string, "")

    # Immutable (ForceNew) — changes how Pub/Sub stores the backlog.
    enable_message_ordering = optional(bool, false)

    enable_exactly_once_delivery = optional(bool, false)

    # User labels, merged beneath the platform attribution labels
    # (see locals.tf).
    labels = optional(map(string), {})

    expiration_policy = optional(object({
      # "" means the subscription never expires.
      ttl = optional(string, "")
    }), null)

    dead_letter_policy = optional(object({
      # Resolved from a GcpPubSubTopic reference to the fully qualified
      # topic path.
      dead_letter_topic     = optional(string, "")
      max_delivery_attempts = optional(number, 0)
    }), null)

    retry_policy = optional(object({
      minimum_backoff = optional(string, "")
      maximum_backoff = optional(string, "")
    }), null)

    push_config = optional(object({
      # Resolved from a GcpCloudRun reference to a literal HTTPS URL.
      push_endpoint = string
      attributes    = optional(map(string), {})
      oidc_token = optional(object({
        # Resolved from a GcpServiceAccount reference to a literal email.
        service_account_email = string
        audience              = optional(string, "")
      }), null)
      # The wrapper choice: no_wrapper renders the provider block; the
      # pubsub_wrapper arm is the provider's default envelope and is consumed
      # here, never forwarded.
      no_wrapper = optional(object({
        write_metadata = optional(bool, false)
      }), null)
      pubsub_wrapper = optional(object({}), null)
    }), null)

    bigquery_config = optional(object({
      # Resolved from a GcpBigQueryTable reference to the dotted
      # {project}.{dataset}.{table} form the Pub/Sub API expects.
      table               = string
      use_topic_schema    = optional(bool, false)
      use_table_schema    = optional(bool, false)
      drop_unknown_fields = optional(bool, false)
      write_metadata      = optional(bool, false)
      # Resolved from a GcpServiceAccount reference to a literal email.
      service_account_email = optional(string, "")
    }), null)

    cloud_storage_config = optional(object({
      # Resolved from a GcpGcsBucket reference (no "gs://" prefix).
      bucket                   = string
      filename_prefix          = optional(string, "")
      filename_suffix          = optional(string, "")
      filename_datetime_format = optional(string, "")
      max_bytes                = optional(number, 0)
      max_duration             = optional(string, "")
      max_messages             = optional(number, 0)
      # The output format: avro_config renders the provider block; the
      # text_config arm is the provider's default text output and is consumed
      # here, never forwarded.
      avro_config = optional(object({
        use_topic_schema = optional(bool, false)
        write_metadata   = optional(bool, false)
      }), null)
      text_config = optional(object({}), null)
      # Resolved from a GcpServiceAccount reference to a literal email.
      service_account_email = optional(string, "")
    }), null)

    # Ordered transform pipeline applied to every message before delivery.
    # Each step carries exactly one transform arm (spec-enforced):
    # javascript_udf or ai_inference.
    message_transforms = optional(list(object({
      javascript_udf = optional(object({
        function_name = string
        code          = string
      }), null)
      ai_inference = optional(object({
        # Vertex AI endpoint path (resolved from a GcpVertexAiEndpoint
        # reference or given as a literal dedicated-endpoint or
        # publisher-model path).
        endpoint              = string
        service_account_email = optional(string, "")
        unstructured_inference = optional(object({
          parameters = optional(map(string), {})
        }), null)
      }), null)
      disabled = optional(bool, false)
    })), [])

    # Resource Manager tags bound at create time (tagKeys/{id} =>
    # tagValues/{id}). Changing them later replaces the subscription.
    resource_manager_tags = optional(map(string), {})

    # Deletion policy: "", "DELETE" (default), "PREVENT" (destroy fails),
    # or "ABANDON" (remove from management, leave serving in GCP).
    deletion_policy = optional(string, "")
  })

  validation {
    condition     = var.spec.subscription_name != ""
    error_message = "subscription_name is required."
  }

  validation {
    condition     = var.spec.topic != ""
    error_message = "topic is required."
  }

  validation {
    condition     = contains(["", "DELETE", "PREVENT", "ABANDON"], var.spec.deletion_policy)
    error_message = "deletion_policy must be one of: DELETE, PREVENT, ABANDON."
  }
}
