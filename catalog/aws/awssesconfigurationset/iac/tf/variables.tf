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
  description = "AwsSesConfigurationSet specification"
  type = object({
    # The AWS region where the configuration set is created. SES resources
    # are regional: identities can only reference configuration sets in
    # their own region.
    # Example: "us-west-2", "eu-west-1".
    region = string

    # Transport and IP-pool controls for messages sent under this set.
    delivery_options = optional(object({
      # TLS enforcement for delivery to receiving mail servers:
      #   REQUIRE  -- deliver only over a TLS-protected connection; fail the
      #               send if the receiver cannot negotiate TLS. The right
      #               choice for anything carrying sensitive content.
      #   OPTIONAL -- attempt TLS but fall back to plaintext (the SMTP
      #               default, and AWS's default when unset).
      tls_policy = optional(string)

      # How long, in seconds (300-50400, i.e. 5 minutes to 14 hours), SES
      # keeps retrying delivery of a message before giving up and returning a
      # bounce. Shorter values suit time-sensitive mail (one-time passcodes
      # are useless after 10 minutes); the AWS default is 14 hours.
      max_delivery_seconds = optional(number)

      # The dedicated IP pool messages under this set are sent from. Pass the
      # name of a dedicated IP pool created outside this catalog (dedicated
      # IPs are a paid capacity surface with their own lifecycle). Leave
      # unset to send from the shared SES IP space.
      sending_pool_name = optional(string, "")
    }))

    # Whether SES publishes reputation metrics (bounce and complaint rates)
    # for this set to CloudWatch. Off by default AWS-side; turn it on for
    # any production sender -- reputation is what SES suspends accounts
    # over, and this is the built-in way to watch it.
    reputation_metrics_enabled = optional(bool, false)

    # Whether email sending is enabled for this configuration set. This is
    # the per-set kill switch: flip it off to immediately stop all sends
    # that use the set (e.g. when a compromised key is spraying spam)
    # without touching identities or application config. Defaults to true.
    sending_enabled = optional(bool)

    # The account-level suppression-list reasons this set honors. When a
    # recipient address is on the account suppression list for one of these
    # reasons, SES silently skips sending to it under this set:
    #   BOUNCE    -- addresses that previously hard-bounced.
    #   COMPLAINT -- addresses whose owners marked mail as spam.
    # Both are strongly recommended for any non-transactional sender; leave
    # the list empty to inherit the account-level default instead of
    # overriding it.
    suppressed_reasons = optional(list(string), [])

    # Custom open/click tracking configuration. SES rewrites links and
    # embeds a tracking pixel through its own domain by default; setting a
    # custom redirect domain keeps tracking URLs on YOUR domain (a CNAME to
    # SES's tracking endpoint), which looks trustworthy to recipients and
    # corporate link scanners.
    tracking_options = optional(object({
      # The custom subdomain SES rewrites tracking links through. The domain
      # must CNAME to the regional SES tracking endpoint (e.g.
      # "click.example.com" -> "r.us-west-2.awstrack.me") -- compose the CNAME
      # with an AwsRoute53DnsRecord. Required when tracking options are set.
      custom_redirect_domain = string

      # The scheme SES uses for the rewritten tracking links:
      #   REQUIRE           -- https for both open and click tracking.
      #   REQUIRE_OPEN_ONLY -- https for the open-tracking pixel only.
      #   OPTIONAL          -- match the scheme of the original link (AWS's
      #                        default when unset).
      https_policy = optional(string)
    }))

    # Virtual Deliverability Manager (VDM) toggles for this set, overriding
    # the account-level VDM configuration.
    vdm_options = optional(object({
      # Whether VDM engagement metrics (opens/clicks broken down by ISP,
      # subject line, and sending identity in the VDM dashboard) are
      # collected for mail sent under this set. Unset is DISABLED.
      engagement_metrics_enabled = optional(bool)

      # Whether VDM optimized shared delivery is used: SES picks the shared
      # IP with the best standing for each receiving ISP instead of rotating
      # blindly. Only meaningful when sending from the shared IP space (not a
      # dedicated pool). Unset is DISABLED.
      optimized_shared_delivery_enabled = optional(bool)
    }))

    # Named event destinations that publish email events (sends, bounces,
    # complaints, opens, clicks, ...) from this set into other AWS
    # services. Each destination is its own AWS sub-resource keyed by name,
    # materialized per-name by the modules, so entries can be added and
    # removed independently without touching the set. Names must be unique.
    event_destinations = optional(list(object({
      # The destination's name, unique within the configuration set (used as
      # the AWS event-destination identifier; changing it replaces the
      # destination).
      name = string

      # Whether the destination actively publishes events. Defaults to true
      # -- a created-but-disabled destination is almost never what a manifest
      # means (AWS's own default is false, a common source of silently
      # missing events; the modules always send this value explicitly).
      enabled = optional(bool)

      # The email events this destination receives:
      #   SEND              -- the send API call was accepted.
      #   REJECT            -- SES refused the message (e.g. a virus was
      #                        detected).
      #   BOUNCE            -- the receiver rejected the message (hard
      #                        bounces; the reputation killer).
      #   COMPLAINT         -- the recipient marked the message as spam.
      #   DELIVERY          -- the receiver accepted the message.
      #   OPEN              -- the recipient opened the message (needs
      #                        tracking).
      #   CLICK             -- the recipient clicked a tracked link.
      #   RENDERING_FAILURE -- a template variable failed to render (template
      #                        sends only).
      #   DELIVERY_DELAY    -- delivery is being retried (soft bounce,
      #                        mailbox full, ...).
      #   SUBSCRIPTION      -- the recipient changed subscription preferences
      #                        (list-management senders).
      matching_event_types = list(string)

      # Publish events as CloudWatch metrics, dimensioned by the configured
      # sources -- the zero-infrastructure way to alarm on bounce rate.
      cloud_watch = optional(object({
        # The dimensions attached to each published metric. At least one is
        # required by AWS.
        dimensions = list(object({
          # The dimension name as it appears in CloudWatch (1-256 characters).
          # Example: "campaign", "sender".
          name = string

          # Where the dimension's value comes from on each message:
          #   MESSAGE_TAG  -- an X-SES-MESSAGE-TAGS tag set by the sender at send
          #                   time (the common choice: tag sends with campaign or
          #                   tenant and slice metrics by it).
          #   EMAIL_HEADER -- a header on the outgoing message.
          #   LINK_TAG     -- a query-string tag on the clicked link (CLICK
          #                   events only).
          value_source = string

          # The value used when the message carries no value for this dimension
          # (1-256 characters). Example: "none".
          default_value = string
        }))
      }))

      # Publish events onto an EventBridge bus for rule-based routing.
      # Reference an AwsEventBridgeBus's bus_arn output or pass a literal
      # ARN. NOTE: AWS only supports the account's DEFAULT bus for SES event
      # publishing today; rules on the default bus can then forward anywhere.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      event_bus = optional(string, "")

      # Stream events into a Kinesis Data Firehose delivery stream -- the
      # firehose-to-S3/Redshift/OpenSearch path for durable event analytics.
      firehose = optional(object({
        # The Firehose delivery stream that receives the events. Reference an
        # AwsKinesisFirehose's delivery_stream_arn output or pass a literal ARN.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        delivery_stream = string

        # The IAM role SES assumes to put records into the delivery stream
        # (trust principal ses.amazonaws.com, with firehose:PutRecordBatch on
        # the stream). Reference an AwsIamRole's role_arn output or pass a
        # literal ARN.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        iam_role = string
      }))

      # Publish each event as an SNS message -- the classic bounce/complaint
      # feedback-loop wiring (subscribe a queue or webhook to the topic).
      # Reference an AwsSnsTopic's topic_arn output or pass a literal ARN.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      sns_topic = optional(string, "")

      # Send engagement events to an Amazon Pinpoint project (literal ARN --
      # Pinpoint is not modeled in this catalog).
      pinpoint_application_arn = optional(string, "")
    })), [])
  })
}
