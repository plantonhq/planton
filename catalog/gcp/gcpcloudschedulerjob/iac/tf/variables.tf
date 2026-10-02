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
  description = "GcpCloudSchedulerJob specification"
  type = object({
    # GCP project where the scheduler job will be created.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # Name of the Cloud Scheduler job.
    # If not specified, defaults to metadata.name.
    # Immutable after creation.
    job_name = optional(string, "")

    # GCP region where the scheduler job will be created (e.g., "us-central1").
    # Immutable after creation.
    location = string

    # Cron schedule on which the job will be executed.
    # Uses unix-cron format (e.g., "*/5 * * * *" for every 5 minutes,
    # "0 9 * * 1" for every Monday at 9:00 AM).
    # The schedule is interpreted in the time zone specified by time_zone.
    schedule = string

    # Time zone name from the tz database (e.g., "America/New_York", "Europe/London").
    # If not specified, defaults to "Etc/UTC".
    # See: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
    time_zone = optional(string, "")

    # Human-readable description of the job.
    # Maximum 500 characters.
    description = optional(string, "")

    # The deadline for job attempts. If the request handler does not respond
    # by this deadline, the request is cancelled and the attempt is marked
    # as a DEADLINE_EXCEEDED failure.
    #
    # For HTTP targets: between 15 seconds and 30 minutes.
    # For App Engine targets: between 15 seconds and 24 hours 15 seconds.
    # For Pub/Sub targets: this field is ignored.
    #
    # Format: duration string (e.g., "180s", "30m").
    # If not specified, defaults to "180s" (3 minutes).
    attempt_deadline = optional(string, "")

    # If true, the job will be created in a paused state (will not execute
    # on schedule until resumed). Defaults to false (job starts enabled).
    paused = optional(bool, false)

    # HTTP target configuration. Dispatches the job to an HTTP endpoint.
    # This is the most common target type, used for triggering Cloud Run
    # services, Cloud Functions, webhooks, or any HTTP-accessible endpoint.
    # Exactly one target must be specified.
    http_target = optional(object({
      # The full URI of the HTTP target.
      # Required. Must be a valid HTTP or HTTPS URL.
      uri = string

      # HTTP request method.
      # If not specified, defaults to POST.
      # Valid values: "POST", "GET", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS".
      http_method = optional(string, "")

      # HTTP request body.
      # A request body is allowed only if the HTTP method is POST, PUT, or PATCH.
      # Must be base64-encoded.
      body = optional(string, "")

      # HTTP request headers.
      # The following headers cannot be set: Content-Length, User-Agent,
      # and headers matching X-Google-* or X-AppEngine-*.
      headers = optional(map(string), {})

      # OAuth2 access token configuration for authenticating requests.
      # Use for calling Google APIs on *.googleapis.com.
      # Mutually exclusive with oidc_token.
      oauth_token = optional(object({
        # Service account email to generate the OAuth token.
        # The service account must be within the same project as the job.
        # The caller must have iam.serviceAccounts.actAs on this service account.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account_email = string

        # OAuth scope for the generated access token.
        # If not specified, defaults to "https://www.googleapis.com/auth/cloud-platform".
        scope = optional(string, "")
      }))

      # OIDC token configuration for authenticating requests.
      # Use for calling Cloud Run, Cloud Functions, or custom endpoints.
      # Mutually exclusive with oauth_token.
      oidc_token = optional(object({
        # Service account email to generate the OIDC token.
        # The service account must be within the same project as the job.
        # The caller must have iam.serviceAccounts.actAs on this service account.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service_account_email = string

        # Audience for the generated OIDC token.
        # If not specified, the URI of the HTTP target will be used.
        audience = optional(string, "")
      }))
    }))

    # Pub/Sub target configuration. Publishes a message to a Pub/Sub topic
    # when the job executes. Use this for event-driven architectures where
    # downstream consumers process messages asynchronously.
    # Exactly one target must be specified.
    pubsub_target = optional(object({
      # The fully qualified Pub/Sub topic name to publish to.
      # Format: projects/{project}/topics/{topic}
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      topic_name = string

      # The message payload for the Pub/Sub message.
      # Must be base64-encoded.
      # The message must contain either non-empty data OR at least one attribute.
      data = optional(string, "")

      # Attributes for the Pub/Sub message.
      # Key-value pairs attached to the message as metadata.
      # The message must contain either non-empty data OR at least one attribute.
      attributes = optional(map(string), {})
    }))

    # App Engine HTTP target configuration. Dispatches the job to an
    # App Engine handler within the same project. Use this when the
    # target handler runs on App Engine.
    # Exactly one target must be specified.
    app_engine_http_target = optional(object({
      # The relative URI of the App Engine handler.
      # Must begin with "/" and have a maximum length of 2083 characters.
      relative_uri = string

      # HTTP request method.
      # If not specified, defaults to POST.
      # Valid values: "POST", "GET", "HEAD", "PUT", "DELETE", "PATCH", "OPTIONS".
      http_method = optional(string, "")

      # HTTP request body.
      # A request body is allowed only if the HTTP method is POST or PUT.
      # Must be base64-encoded.
      body = optional(string, "")

      # HTTP request headers.
      # The following headers cannot be set: Content-Length, Host, User-Agent,
      # and headers matching X-Google-* or X-AppEngine-*.
      headers = optional(map(string), {})

      # App Engine routing configuration.
      # Controls which App Engine service, version, and instance handles
      # the request. If not set, the default service and version are used.
      app_engine_routing = optional(object({
        # The App Engine service to route the request to.
        # If not specified, the default service is used.
        service = optional(string, "")

        # The App Engine version to route the request to.
        # If not specified, the default version is used.
        version = optional(string, "")

        # The App Engine instance to route the request to.
        # If not specified, the request is routed according to the service/version
        # traffic splitting configuration.
        instance = optional(string, "")
      }))
    }))

    # Retry configuration for failed job attempts.
    # Controls exponential backoff behavior, maximum attempts, and
    # retry duration limits.
    retry_config = optional(object({
      # The number of attempts that the system will make to run a job using
      # the exponential backoff procedure. Values greater than 5 and negative
      # values are not allowed.
      retry_count = optional(number, 0)

      # The time limit for retrying a failed job, measured from when the job
      # was first attempted. Once elapsed, no further attempts are made.
      # Format: duration string (e.g., "3600s" for 1 hour).
      # Set to "0s" for unlimited retry duration.
      max_retry_duration = optional(string, "")

      # The minimum amount of time to wait before retrying a job after it fails.
      # Format: duration string (e.g., "5s" for 5 seconds).
      min_backoff_duration = optional(string, "")

      # The maximum amount of time to wait before retrying a job after it fails.
      # Format: duration string (e.g., "3600s" for 1 hour).
      max_backoff_duration = optional(string, "")

      # The number of times that the retry interval doubles before becoming
      # constant. The retry interval starts at min_backoff_duration, then
      # doubles max_doublings times, and increases linearly thereafter.
      max_doublings = optional(number, 0)
    }))

    # What destroying this resource does to the job:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the job is deleted and its schedule stops firing
    #   "PREVENT" -- destroy FAILS; protects a job whose missed runs would
    #                break downstream systems
    #   "ABANDON" -- the job is removed from management but keeps firing on
    #                schedule in GCP
    deletion_policy = optional(string, "")
  })
}
