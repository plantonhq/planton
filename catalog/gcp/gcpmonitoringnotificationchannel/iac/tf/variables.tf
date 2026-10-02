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
  description = "GcpMonitoringNotificationChannel specification"
  type = object({
    # The GCP project that owns the notification channel.
    # Can be a literal project ID or a reference to a GcpProject resource.
    # If omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The channel type — which delivery mechanism this channel uses. GCP
    # validates the value server-side against its live channel-type catalog
    # (there is no fixed client-side list; new types appear without provider
    # releases). Common values:
    #   email              -- channel_labels: email_address
    #   sms                -- channel_labels: number (E.164, e.g. +15551234567)
    #   slack              -- channel_labels: channel_name (e.g. #alerts);
    #                         sensitive_labels: auth_token
    #   pagerduty          -- sensitive_labels: service_key
    #   webhook_tokenauth  -- channel_labels: url; the token rides the URL
    #   webhook_basicauth  -- channel_labels: url, username;
    #                         sensitive_labels: password
    #   pubsub             -- channel_labels: topic (projects/{p}/topics/{t})
    # Immutable in practice for most types: GCP updates a channel's type only
    # by delete-and-recreate semantics on the provider side.
    type = string

    # Human-readable name shown in the Cloud Monitoring console and in
    # notification footers. Defaults to metadata.name when left empty.
    # Limited to 512 Unicode characters by the API.
    display_name = optional(string, "")

    # Why this channel exists and who owns the endpoint — write it for the
    # operator triaging a 3am page. Limited to 1024 bytes by the API.
    description = optional(string, "")

    # Type-specific, NON-SECRET configuration keys (maps to the provider's
    # `labels` argument — distinct from the `labels` field below, which is
    # user metadata). Which keys apply depends on `type`; see the type list
    # above. GCP rejects unknown keys for the chosen type at apply time.
    # Credentials (auth_token, password, service_key) are refused here by
    # validation — they belong in sensitive_labels.
    channel_labels = optional(map(string), {})

    # Credentials for channel types that authenticate to an external service.
    # Each field is a secret: the platform stores it as a managed-secret
    # reference and resolves it just-in-time at deploy — it never sits in
    # plaintext in the control plane.
    sensitive_labels = optional(object({
      # OAuth token for the slack channel type (from the Slack app
      # installation).
      auth_token = optional(string, "")

      # HTTP basic-auth password for the webhook_basicauth channel type.
      password = optional(string, "")

      # Integration/service key for the pagerduty channel type (from the
      # PagerDuty service's integration settings).
      service_key = optional(string, "")
    }))

    # Whether notifications are forwarded to the described channel (default
    # true). A disabled channel keeps its configuration and its references
    # from alert policies but delivers nothing — the safe way to silence an
    # endpoint temporarily without rewiring policies. Both IaC engines send
    # the value explicitly so behavior is identical regardless of engine.
    enabled = optional(bool)

    # If true, deleting the channel proceeds even when alert policies still
    # reference it — GCP removes the channel from those policies in the same
    # operation. If false (default), the delete FAILS while references exist,
    # which is the safer posture: a dangling policy silently loses its
    # delivery endpoint.
    force_delete = optional(bool, false)

    # User labels attached to the channel for organizing and identifying it
    # (maps to the provider's user_labels), merged with Planton's platform
    # labels (which win on key conflicts). Keys and values may contain only
    # lowercase letters, numerals, underscores, and dashes; keys must begin
    # with a letter.
    labels = optional(map(string), {})

    # Deletion policy — what happens when this resource is destroyed:
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the channel is deleted (fails while alert policies still
    #                reference it unless force_delete is true)
    #   "PREVENT" -- destroy FAILS; protects the paging path of a production
    #                alerting setup from accidental teardown
    #   "ABANDON" -- the channel is removed from management but keeps
    #                delivering notifications in GCP
    deletion_policy = optional(string, "")
  })
}
