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
  description = "GcpEventarcTrigger specification"
  type = object({
    # The GCP project to create the trigger in. Can be a literal project ID
    # or a reference to a GcpProject resource. If omitted, the provider's
    # default project is used.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    project_id = optional(string, "")

    # The location of the trigger: a region (e.g. us-central1) or "global".
    # Must match where the events originate — Cloud Storage triggers live in
    # the bucket's region, audit-log triggers are usually global. Immutable.
    location = string

    # The trigger name in GCP. Defaults to metadata.name when left empty.
    # Immutable: changing it replaces the trigger.
    trigger_name = optional(string, "")

    # The event filters — ALL criteria must match for an event to be
    # delivered. Every trigger MUST filter the "type" attribute (the
    # CloudEvents type, e.g. google.cloud.pubsub.topic.v1.messagePublished or
    # google.cloud.audit.log.v1.written); the API's catalog of types and
    # their filterable attributes: https://cloud.google.com/eventarc/docs/reference/supported-events
    matching_criteria = list(object({
      # The CloudEvents attribute to filter (e.g. "type", "bucket",
      # "serviceName", "methodName"). Only the attributes the event type
      # declares filterable are accepted by the API.
      attribute = string

      # The value the attribute must match. With operator
      # "match-path-pattern", path patterns like "objects/prefix/*" are
      # matched instead of exact equality.
      value = string

      # Empty for exact match (the default). The only other value the API
      # accepts is "match-path-pattern" (supported on a subset of attributes,
      # e.g. Cloud Storage object names and audit-log resourceName).
      operator = optional(string, "")
    }))

    # Where matching events are delivered. Exactly one arm.
    destination = object({
      # Deliver to a Cloud Run service (the most common destination).
      cloud_run_service = optional(object({
        # The Cloud Run service NAME (bare name, not a URL) — a literal or a
        # reference to a GcpCloudRun resource (its service_name output). Only
        # services in the same project as the trigger can be addressed.
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        service = string

        # The region the Cloud Run service is deployed in. The API REQUIRES
        # this on every create (live-verified: 400 "cloud_run.region is empty"
        # — it never infers the region from the trigger). When empty, the
        # MODULE defaults it to the trigger's location; set it explicitly when
        # the service lives in a different region than the trigger, and always
        # on a "global" trigger (spec-enforced — global is not a Cloud Run
        # region).
        region = optional(string, "")

        # The relative path on the service events are POSTed to (RFC2396 path
        # segment, e.g. "/events" or "route/subroute"). Empty means the root
        # path.
        path = optional(string, "")
      }))

      # Deliver to a service running in a GKE cluster. Requires Eventarc's
      # GKE destination support to be enabled once per project
      # (gcloud eventarc gke-destinations init) — Eventarc then manages an
      # event-forwarder pod in the cluster.
      gke = optional(object({
        # The GKE cluster NAME the service runs in — a literal or a reference to
        # a GcpGkeCluster resource (its name output). Must be in the same
        # project as the trigger. The reference is containment-exempt: the
        # trigger DELIVERS INTO the cluster, it does not live inside it (the
        # trigger's home is its project).
        # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
        cluster = string

        # The cluster's compute location: a zone (us-central1-a) for zonal
        # clusters or a region (us-central1) for regional clusters.
        location = string

        # The Kubernetes namespace the destination service lives in.
        namespace = string

        # The Kubernetes service NAME events are delivered to.
        service = string

        # The relative path on the service events are POSTed to (RFC2396 path
        # segment). Empty means the root path.
        path = optional(string, "")
      }))

      # Trigger a Cloud Workflows EXECUTION per event. The full workflow
      # resource name (projects/{project}/locations/{location}/workflows/{name})
      # — a literal or a reference to a GcpWorkflow resource (its workflow_id
      # output is exactly this value). Must live in the same project as the
      # trigger. This arm REQUIRES spec.service_account (spec-enforced; the
      # API rejects the create without it).
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      workflow = optional(string, "")

      # Deliver to a private HTTP endpoint reachable through a VPC network
      # attachment (Eventarc Standard's bring-your-own-HTTP destination).
      http_endpoint = optional(object({
        # The endpoint URI (RFC2396, e.g.
        # https://svc.us-central1.p.local:8080/route). Only HTTPS is supported.
        uri = string

        # The network attachment that lets Eventarc reach the endpoint's VPC
        # (format:
        # projects/{project}/regions/{region}/networkAttachments/{name}).
        # REQUIRED for HTTP endpoint destinations — the provider models this as
        # a separate network_config block but permits it only with HTTP
        # endpoints, so this spec carries it inside the arm.
        network_attachment = string
      }))
    })

    # The IAM service account email the trigger runs as — it must hold
    # roles/eventarc.eventReceiver, plus roles/run.invoker for authenticated
    # Cloud Run destinations (identity tokens are minted from this account)
    # or roles/workflows.invoker for workflow destinations. Audit-log
    # triggers and workflow destinations REQUIRE a service account (the
    # workflow case is API-enforced at create — live-verified 400 without
    # it). A literal email or a reference to a GcpServiceAccount resource.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    service_account = optional(string, "")

    # For google.cloud.pubsub.topic.v1.messagePublished triggers ONLY: use
    # an EXISTING Pub/Sub topic (projects/{project}/topics/{topic} — a
    # literal or a reference to a GcpPubSubTopic resource) as the transport
    # instead of letting Eventarc create one. The topic is NOT deleted when
    # the trigger is destroyed (Eventarc only manages topics it created).
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    transport_pubsub_topic = optional(string, "")

    # The MIME type the CloudEvent data field is delivered as (e.g.
    # application/json or application/protobuf). The API defaults to
    # application/json when unset. NOT accepted on Pub/Sub
    # (messagePublished) triggers — the API rejects any value there
    # (live-verified 400; a Pub/Sub event's payload format is decided by
    # the publisher). Spec-enforced above.
    event_data_content_type = optional(string, "")

    # Delivery retry ceiling. The provider accepts only the value 1
    # ("The only valid value is 1" — its own schema note), which DISABLES
    # Eventarc's default retries: a failed delivery is not retried. Leave
    # unset (0) for the platform default retry behavior. Can only be set
    # with Cloud Run destinations (provider constraint, enforced above).
    retry_max_attempts = optional(number, 0)

    # User labels attached to the trigger (merged with the platform's
    # standard labels by the module).
    labels = optional(map(string), {})

    # Receive events from an Eventarc SaaS PARTNER (e.g. Datadog): the
    # module creates the partner channel alongside the trigger and wires the
    # trigger to it. The channel's activation_token stack output must be
    # handed to the partner to complete the handshake — until then the
    # channel stays PENDING and delivers nothing.
    partner_channel = optional(object({
      # The channel name in GCP. Defaults to "{trigger name}-channel" when
      # left empty. Immutable: changing it replaces the channel (a NEW
      # activation token — redo the partner handshake).
      channel_name = optional(string, "")

      # The partner provider the channel receives events from (format:
      # projects/{project}/locations/{location}/providers/{provider_id} —
      # list available partners with gcloud eventarc providers list).
      # Immutable: changing it replaces the channel.
      third_party_provider = string

      # CMEK for events in transit through the channel. The full crypto key
      # resource name — a literal or a reference to a GcpKmsKey resource.
      # Omit for Google-managed encryption.
      # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
      crypto_key = optional(string, "")
    }))

    # CMEK for the project/location's GOOGLE channel — the shared conduit
    # ALL non-partner triggers in this project+location deliver through.
    # The full crypto key resource name — a literal or a reference to a
    # GcpKmsKey resource. The module manages the per-project-per-location
    # googleChannelConfig SINGLETON: set this from AT MOST ONE trigger per
    # project+location (a second manager fights over the same singleton).
    # Deleting the config is a state-only no-op in GCP — the singleton
    # always exists; clearing this field reverts it to Google-managed
    # encryption.
    # Accepts a literal value or a reference in the manifest; the CLI resolves it to a plain string before the module runs.
    google_channel_crypto_key = optional(string, "")

    # Deletion policy — what happens when this resource is destroyed (also
    # applied to the partner channel and google-channel config the kind
    # manages):
    #   ""        -- same as "DELETE" (provider default)
    #   "DELETE"  -- the trigger (and any partner channel) is deleted;
    #                events stop being delivered immediately
    #   "PREVENT" -- destroy FAILS; protects a production event route
    #   "ABANDON" -- resources are removed from management but keep
    #                delivering in GCP
    deletion_policy = optional(string, "")
  })
}
