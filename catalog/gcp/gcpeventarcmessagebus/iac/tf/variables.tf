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
  description = "GcpEventarcMessageBus specification"
  type = object({
    # The GCP project to create the bus (and all satellites) in. Can be a
    # literal project ID or a reference to a GcpProject resource. If
    # omitted, the provider's default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The region for the bus and every satellite (e.g. us-central1).
    # Eventarc Advanced serves a subset of regions
    # (https://cloud.google.com/eventarc/docs/locations) — the API rejects
    # unsupported ones at create time. Immutable.
    location = string

    # The bus ID in GCP (also its short name). Defaults to metadata.name
    # when left empty. Format: 1–63 chars, lowercase letters, digits, and
    # hyphens; must start with a letter and end alphanumeric (the API's
    # documented pattern). Immutable: changing it replaces the bus.
    message_bus_id = optional(string, "")

    # Display name shown in the console.
    display_name = optional(string, "")

    # User labels attached to the bus (merged with the platform's standard
    # labels by the module).
    labels = optional(map(string), {})

    # Free-form annotations attached to the bus
    # (https://google.aip.dev/128#annotations) — non-identifying metadata
    # for tools; unlike labels they are not usable in filters or billing.
    annotations = optional(map(string), {})

    # CMEK for messages at rest in the bus. The full crypto key resource
    # name (projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a
    # literal or a reference to a GcpKmsKey resource. The key must be in the
    # same region as the bus; grant the Eventarc service agent
    # roles/cloudkms.cryptoKeyEncrypterDecrypter BEFORE creating. Omit for
    # Google-managed encryption.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    crypto_key = optional(string, "")

    # The minimum severity of bus activity recorded to Cloud Logging.
    # Empty uses the API default (NONE — no platform logs). One of: NONE,
    # DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY
    # (the provider's ValidateEnum list). INFO is the operational sweet
    # spot while onboarding sources and pipelines.
    log_severity = optional(string, "")

    # Google API sources publishing Google-service events INTO this bus.
    # Each becomes a google_api_source resource whose destination the module
    # wires to THIS bus (byte-identically on both IaC engines) — an api
    # source feeding someone else's bus belongs to that bus's kind instance.
    google_api_sources = optional(list(object({
      # The source ID in GCP. Format: 1–63 chars, lowercase letters, digits,
      # and hyphens; starts with a letter, ends alphanumeric. Immutable:
      # changing it replaces the source.
      source_id = string

      # Display name shown in the console.
      display_name = optional(string, "")

      # User labels attached to the source.
      labels = optional(map(string), {})

      # Free-form annotations attached to the source.
      annotations = optional(map(string), {})

      # CMEK for this source's data. The full crypto key resource name — a
      # literal or a reference to a GcpKmsKey resource. Omit for
      # Google-managed encryption.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      crypto_key = optional(string, "")

      # The minimum severity of this source's activity recorded to Cloud
      # Logging (same value set as the bus's log_severity).
      log_severity = optional(string, "")
    })), [])

    # Pipelines delivering messages OUT of the bus. Referenced by
    # enrollments below via pipeline_id.
    pipelines = optional(list(object({
      # The pipeline ID in GCP. Format: 1–63 chars, lowercase letters, digits,
      # and hyphens; starts with a letter, ends alphanumeric. Immutable:
      # changing it replaces the pipeline.
      pipeline_id = string

      # Where this pipeline delivers. Exactly one target arm.
      destination = object({
        # Deliver to an HTTP endpoint reachable through a VPC network
        # attachment.
        http_endpoint = optional(object({
          # The endpoint URI (RFC2396). Only HTTPS is supported (the API rejects
          # http://), e.g. https://svc.us-central1.p.local:8080/route.
          uri = string

          # A CEL expression shaping the outgoing HTTP request (headers, body
          # binding) — see
          # https://cloud.google.com/eventarc/advanced/docs/receive-events/create-message-binding.
          # Empty uses the default CloudEvents HTTP binding.
          message_binding_template = optional(string, "")

          # The network attachment that lets the pipeline reach the endpoint's
          # VPC (format:
          # projects/{project}/regions/{region}/networkAttachments/{name}).
          # REQUIRED for HTTP endpoint destinations; forbidden for every other
          # target (the provider's own rule, enforced at the pipeline level).
          network_attachment = optional(string, "")
        }))

        # Publish to a Pub/Sub topic. A literal topic resource name or a
        # reference to a GcpPubSubTopic resource (its topic_id output,
        # projects/{project}/topics/{topic}). The provider documents the form
        # projects/{project}/locations/{location}/topics/{topic} for this field;
        # both name the same topic — the live API accepts the canonical Pub/Sub
        # form (live-verified: a pipeline created with the canonical form went
        # ACTIVE and re-planned clean on both engines).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        topic = optional(string, "")

        # Trigger a Cloud Workflows EXECUTION per message. The full workflow
        # resource name — a literal or a reference to a GcpWorkflow resource
        # (its workflow_id output). Must be deployed in the same project as the
        # pipeline.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        workflow = optional(string, "")

        # Chain into ANOTHER message bus (cross-bus routing). The full bus
        # resource name
        # (projects/{project}/locations/{location}/messageBuses/{bus}) — e.g.
        # another GcpEventarcMessageBus's message_bus_name output. Must be in
        # the same project as the pipeline.
        message_bus = optional(string, "")
      })

      # How the pipeline authenticates to the destination (HTTP endpoints
      # that verify identity). At most one mechanism.
      authentication = optional(object({
        # Authenticate with a Google OIDC ID token — for endpoints that verify
        # Google-signed identity tokens (Cloud Run, IAP-protected services).
        google_oidc = optional(object({
          # The service account email the token is minted for — a literal or a
          # reference to a GcpServiceAccount resource. The Eventarc service agent
          # must hold roles/iam.serviceAccountTokenCreator on it.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          service_account = string

          # The audience claim in the minted token. Empty uses the destination
          # URI.
          audience = optional(string, "")
        }))

        # Authenticate with an OAuth access token — preferred for Google APIs
        # that accept OAuth.
        oauth_token = optional(object({
          # The service account email the token is minted for — a literal or a
          # reference to a GcpServiceAccount resource. The Eventarc service agent
          # must hold roles/iam.serviceAccountTokenCreator on it.
          # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
          service_account = string

          # The OAuth scope in the minted token. Empty uses
          # https://www.googleapis.com/auth/cloud-platform.
          scope = optional(string, "")
        }))
      }))

      # The payload format messages ARRIVE in (from the bus). Set input and
      # output together to convert between formats; leave both unset to pass
      # payloads through untouched.
      input_payload_format = optional(object({
        # Avro format with its schema definition.
        avro = optional(object({
          # The schema definition text (an Avro schema JSON or a protobuf
          # definition). Required by the API when converting formats.
          schema_definition = optional(string, "")
        }))

        # JSON format (no schema).
        json = optional(bool, false)

        # Protobuf format with its schema definition.
        protobuf = optional(object({
          # The schema definition text (an Avro schema JSON or a protobuf
          # definition). Required by the API when converting formats.
          schema_definition = optional(string, "")
        }))
      }))

      # The payload format messages are DELIVERED in. Avro/Protobuf require a
      # schema_definition; converting between Avro and Protobuf requires both
      # sides to define schemas.
      output_payload_format = optional(object({
        # Avro format with its schema definition.
        avro = optional(object({
          # The schema definition text (an Avro schema JSON or a protobuf
          # definition). Required by the API when converting formats.
          schema_definition = optional(string, "")
        }))

        # JSON format (no schema).
        json = optional(bool, false)

        # Protobuf format with its schema definition.
        protobuf = optional(object({
          # The schema definition text (an Avro schema JSON or a protobuf
          # definition). Required by the API when converting formats.
          schema_definition = optional(string, "")
        }))
      }))

      # A CEL expression rewriting the message before delivery (e.g.
      # 'message.removeFields(["data.secret"])'). The API allows at most ONE
      # mediation (transformation) per pipeline — hence a single template
      # rather than a list.
      mediation_transformation_template = optional(string, "")

      # Delivery retry policy. Unset uses the API defaults (5 attempts,
      # 5s–60s exponential backoff).
      retry_policy = optional(object({
        # Maximum delivery attempts, 1–100 (the API's documented range).
        # 0 means unset — the API default of 5.
        max_attempts = optional(number, 0)

        # Minimum wait between attempts, in seconds-suffixed form (e.g. "5s").
        # The API accepts 1–600 seconds; its default is 5s.
        min_retry_delay = optional(string, "")

        # Maximum wait between attempts, in seconds-suffixed form (e.g. "60s").
        # The API accepts 1–600 seconds; its default is 60s. Note the API quirk
        # its docs call out: when setting min and max, they may be required to
        # be EQUAL on some surfaces — live behavior decides.
        max_retry_delay = optional(string, "")
      }))

      # Display name shown in the console.
      display_name = optional(string, "")

      # User labels attached to the pipeline.
      labels = optional(map(string), {})

      # Free-form annotations attached to the pipeline.
      annotations = optional(map(string), {})

      # CMEK for this pipeline's data. The full crypto key resource name — a
      # literal or a reference to a GcpKmsKey resource. Omit for
      # Google-managed encryption.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      crypto_key = optional(string, "")

      # The minimum severity of this pipeline's activity recorded to Cloud
      # Logging (same value set as the bus's log_severity).
      log_severity = optional(string, "")
    })), [])

    # Enrollments — the routing table: each selects messages from the bus
    # with a CEL expression and delivers them to one of this spec's
    # pipelines.
    enrollments = optional(list(object({
      # The enrollment ID in GCP. Format: 1–63 chars, lowercase letters,
      # digits, and hyphens; starts with a letter, ends alphanumeric.
      # Immutable: changing it replaces the enrollment.
      enrollment_id = string

      # The CEL expression selecting which bus messages this enrollment
      # routes (evaluated against the CloudEvent, e.g.
      # message.type == "google.cloud.storage.object.v1.finalized"). "true"
      # routes everything.
      cel_match = string

      # The pipeline_id of the pipeline (defined in this spec) that delivers
      # the selected messages. The module renders the full pipeline resource
      # name the API demands; a spec-level rule rejects ids that match no
      # sibling pipeline.
      pipeline = string

      # Display name shown in the console.
      display_name = optional(string, "")

      # User labels attached to the enrollment.
      labels = optional(map(string), {})

      # Free-form annotations attached to the enrollment.
      annotations = optional(map(string), {})
    })), [])

    # Deletion policy — what happens when this resource is destroyed
    # (applied to the bus and every satellite):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- bus, sources, enrollments, and pipelines are deleted;
    #                undelivered messages are lost
    #   "PREVENT" -- destroy FAILS; protects a production eventing hub
    #   "ABANDON" -- resources are removed from management but keep running
    #                in GCP
    deletion_policy = optional(string, "")
  })
}
