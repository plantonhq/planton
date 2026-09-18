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
  description = "GcpCloudTasksQueue specification"
  type = object({
    # GCP project where the queue will be created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Cloud Tasks queue.
    # Must start with a letter, contain only letters, numbers, and hyphens,
    # and be between 1 and 63 characters.
    # Immutable after creation, and deliberately required: a deleted queue's
    # ID is reserved by the API for up to 7 days, so the name deserves an
    # explicit, stable choice rather than a derived default.
    queue_name = string

    # GCP region where the queue will be created (e.g., "us-central1").
    # Immutable after creation.
    location = string

    # Queue-level HTTP target configuration. When set, these settings apply
    # to all HTTP tasks dispatched from this queue, overriding task-level
    # HTTP configuration.
    #
    # This is the recommended pattern for microservices: configure auth and
    # routing at the queue level, then enqueue tasks with just a request body.
    http_target = optional(object({
      # HTTP method override for all tasks in this queue.
      # When specified, overrides the method on individual tasks.
      # Note: if set to GET, the task body is ignored at execution time.
      # Valid values: "POST", "GET", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS",
      # or the API's explicit "HTTP_METHOD_UNSPECIFIED" sentinel (equivalent to
      # leaving the override unset).
      http_method = optional(string, "")

      # HTTP headers to set on all tasks dispatched from this queue.
      # These headers override any task-level headers with the same key.
      # Header size must be less than 80KB total.
      header_overrides = optional(list(object({
        # The header field name.
        key = string

        # The header field value.
        value = string
      })), [])

      # OAuth2 access token configuration for authenticating HTTP task requests.
      # Use for calling Google APIs on *.googleapis.com.
      # Mutually exclusive with oidc_token.
      oauth_token = optional(object({
        # Service account email to generate the OAuth token.
        # The service account must be within the same project as the queue.
        # The caller must have iam.serviceAccounts.actAs on this service account.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account_email = string

        # OAuth scope for the generated access token.
        # If not specified, defaults to "https://www.googleapis.com/auth/cloud-platform".
        scope = optional(string, "")
      }))

      # OIDC token configuration for authenticating HTTP task requests.
      # Use for calling Cloud Run, Cloud Functions, or custom endpoints.
      # Mutually exclusive with oauth_token.
      oidc_token = optional(object({
        # Service account email to generate the OIDC token.
        # The service account must be within the same project as the queue.
        # The caller must have iam.serviceAccounts.actAs on this service account.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account_email = string

        # Audience for the generated OIDC token.
        # If not specified, the URI specified in the target will be used.
        audience = optional(string, "")
      }))

      # URI override settings. When specified, modifies the URI of all tasks
      # dispatched from this queue before dispatch.
      uri_override = optional(object({
        # Scheme override. Replaces the task URI scheme with HTTP or HTTPS.
        # Valid values: "HTTP", "HTTPS".
        scheme = optional(string, "")

        # Host override. Replaces the host part of the task URL.
        # For example, if the task URL is "https://www.google.com" and host is
        # "example.net", the overridden URI becomes "https://example.net".
        # Must not be empty when set (INVALID_ARGUMENT).
        host = optional(string, "")

        # Port override. Replaces the port part of the task URI.
        # Must be a positive integer. Setting to "0" clears the URI port.
        port = optional(string, "")

        # Path override. Replaces the existing path of the task URL.
        # Setting to an empty string clears the URI path segment.
        path = optional(string, "")

        # Query parameters override. Replaces the query part of the task URI.
        # For example: "qparam1=123&qparam2=456".
        # Setting to an empty string clears the URI query segment.
        query_params = optional(string, "")

        # URI Override Enforce Mode.
        # ALWAYS: Always override the task URI (default).
        # IF_NOT_EXISTS: Only apply the override if the task does not already have
        # the corresponding URI component set.
        enforce_mode = optional(string, "")
      }))
    }))

    # App Engine routing override for App Engine tasks in this queue.
    # When set, overrides each task's own App Engine routing so the whole
    # queue targets one service/version/instance. Only relevant for queues
    # dispatching App Engine tasks; ignored for HTTP tasks.
    app_engine_routing_override = optional(object({
      # The App Engine service to route queue tasks to.
      # If not specified, the task is sent to the service that is the default
      # service when the task is attempted.
      service = optional(string, "")

      # The App Engine version to route queue tasks to.
      # If not specified, the task is sent to the version that is the default
      # version when the task is attempted.
      version = optional(string, "")

      # The App Engine instance to route queue tasks to.
      # If not specified, the task is sent to an instance which is available
      # when the task is attempted (subject to the service/version's scaling).
      instance = optional(string, "")
    }))

    # Rate limits for task dispatches. Controls how fast and how many tasks
    # are dispatched concurrently.
    rate_limits = optional(object({
      # Maximum rate at which tasks are dispatched from this queue (tasks/second).
      # If unspecified, Cloud Tasks picks a default based on queue configuration.
      max_dispatches_per_second = optional(number, 0)

      # Maximum number of concurrent tasks that Cloud Tasks allows to be
      # dispatched for this queue. After this threshold, Cloud Tasks stops
      # dispatching until the concurrency drops.
      # If unspecified, Cloud Tasks picks a default.
      max_concurrent_dispatches = optional(number, 0)
    }))

    # Retry configuration for failed task attempts. Controls backoff behavior,
    # maximum attempts, and retry duration.
    retry_config = optional(object({
      # Number of attempts per task. Includes the first attempt.
      # Must be >= -1. Set to -1 for unlimited attempts.
      # If unspecified, Cloud Tasks picks a default.
      max_attempts = optional(number, 0)

      # Maximum time limit for retrying a failed task, measured from the first attempt.
      # Once elapsed, no further attempts are made regardless of max_attempts.
      # Set to "0s" for unlimited retry duration.
      # Format: duration string (e.g., "3600s" for 1 hour).
      max_retry_duration = optional(string, "")

      # Minimum wait time between retry attempts.
      # Format: duration string (e.g., "0.100s" for 100ms).
      min_backoff = optional(string, "")

      # Maximum wait time between retry attempts.
      # Format: duration string (e.g., "3600s" for 1 hour).
      max_backoff = optional(string, "")

      # The number of times the retry interval doubles before becoming constant.
      # The retry interval starts at min_backoff, doubles max_doublings times,
      # then increases linearly until reaching max_backoff.
      max_doublings = optional(number, 0)
    }))

    # Cloud Logging configuration for task dispatch operations.
    stackdriver_logging_config = optional(object({
      # Fraction of operations to log. Must be between 0.0 and 1.0 inclusive.
      # 0.0 means no logging (default), 1.0 means log every dispatch operation.
      sampling_ratio = optional(number, 0)
    }))

    # Dispatch state the queue is held in:
    #   ""        -- same as "RUNNING" (the provider default)
    #   "RUNNING" -- tasks are dispatched to their targets
    #   "PAUSED"  -- tasks accumulate in the queue but none are dispatched —
    #                the safe holding state during target maintenance or
    #                incident response
    # Declarative for spec-driven transitions: editing this value and
    # applying pauses/resumes the queue (live-verified both directions).
    # NOT drift-correcting: the provider treats this as a config-only
    # (virtual) field and never reads the live dispatch state back, so an
    # out-of-band gcloud pause survives applies whose spec value is
    # unchanged. Recover by resuming out-of-band, or by flipping this field
    # PAUSED → apply → RUNNING → apply so the value change triggers the
    # provider's resume call.
    desired_state = optional(string, "")

    # What destroying this resource does to the queue:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the queue and every task still in it are deleted; the
    #                queue's ID is reserved by the API for up to 7 days
    #   "PREVENT" -- destroy FAILS; protects a queue whose backlog must not
    #                be lost
    #   "ABANDON" -- the queue is removed from management but keeps running
    #                (and dispatching) in GCP
    deletion_policy = optional(string, "")
  })
}
