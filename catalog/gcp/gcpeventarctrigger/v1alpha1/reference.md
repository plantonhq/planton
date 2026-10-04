# GcpEventarcTrigger

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpEventarcTriggerSpec defines an Eventarc trigger — the routing rule
"when THIS event happens, call THAT service": events matching the
criteria (a Pub/Sub message, a Cloud Storage object change, an audit-log
entry, a SaaS partner event) are delivered as CloudEvents to a Cloud Run
service, a GKE service, a Workflow execution, or a private HTTP
endpoint.

The trigger's IDENTITY is service_account: it must hold
roles/eventarc.eventReceiver (and, for Cloud Run destinations,
roles/run.invoker on the service). The first trigger in a project also
provisions Eventarc's service agent — expect a few minutes of
propagation before the first delivery succeeds.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpEventarcTrigger
metadata:
  name: my-sample-eventarc-trigger
spec:
  # A region or "global"; must match where the events originate.
  location: us-central1

  # ALL criteria must match; every trigger MUST filter the type
  # attribute (the API's own rule).
  matchingCriteria:
    - attribute: type
      value: google.cloud.pubsub.topic.v1.messagePublished

  # Exactly one destination arm. The service references the Cloud Run
  # fixture the lanes deploy first (its bare service name).
  destination:
    cloudRunService:
      service:
        valueFrom:
          kind: GcpCloudRun
          name: planton-oss-e2e-gcprun-prereq
          fieldPath: status.outputs.service_name

  # What a destroy does: DELETE (event delivery stops immediately),
  # PREVENT (the posture for production routes), or ABANDON (keep
  # delivering unmanaged).
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.triggerName` | `string` |  |  |  |
| `spec.matchingCriteria` | `[]GcpEventarcTriggerMatchingCriterion` | yes |  |  |
| `spec.matchingCriteria[].attribute` | `string` | yes |  |  |
| `spec.matchingCriteria[].value` | `string` | yes |  |  |
| `spec.matchingCriteria[].operator` | `string` |  |  |  |
| `spec.destination` | `GcpEventarcTriggerDestination` | yes |  |  |
| `spec.destination.cloudRunService` | `GcpEventarcTriggerCloudRunDestination` |  |  |  |
| `spec.destination.cloudRunService.service` | `string \| valueFrom` | yes |  | GcpCloudRun (`status.outputs.service_name`) |
| `spec.destination.cloudRunService.region` | `string` |  |  |  |
| `spec.destination.cloudRunService.path` | `string` |  |  |  |
| `spec.destination.gke` | `GcpEventarcTriggerGkeDestination` |  |  |  |
| `spec.destination.gke.cluster` | `string \| valueFrom` | yes |  | GcpGkeCluster (`status.outputs.name`) |
| `spec.destination.gke.location` | `string` | yes |  |  |
| `spec.destination.gke.namespace` | `string` | yes |  |  |
| `spec.destination.gke.service` | `string` | yes |  |  |
| `spec.destination.gke.path` | `string` |  |  |  |
| `spec.destination.workflow` | `string \| valueFrom` |  |  | GcpWorkflow (`status.outputs.workflow_id`) |
| `spec.destination.httpEndpoint` | `GcpEventarcTriggerHttpEndpointDestination` |  |  |  |
| `spec.destination.httpEndpoint.uri` | `string` | yes |  |  |
| `spec.destination.httpEndpoint.networkAttachment` | `string` | yes |  |  |
| `spec.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.transportPubsubTopic` | `string \| valueFrom` |  |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.eventDataContentType` | `string` |  |  |  |
| `spec.retryMaxAttempts` | `int32` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.partnerChannel` | `GcpEventarcTriggerPartnerChannel` |  |  |  |
| `spec.partnerChannel.channelName` | `string` |  |  |  |
| `spec.partnerChannel.thirdPartyProvider` | `string` | yes |  |  |
| `spec.partnerChannel.cryptoKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.googleChannelCryptoKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project to create the trigger in. Can be a literal project ID
or a reference to a GcpProject resource. If omitted, the provider's
default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The location of the trigger: a region (e.g. us-central1) or "global".
Must match where the events originate — Cloud Storage triggers live in
the bucket's region, audit-log triggers are usually global. Immutable.

- rule: {"required":true}

### spec.triggerName

`string`

The trigger name in GCP. Defaults to metadata.name when left empty.
Immutable: changing it replaces the trigger.

### spec.matchingCriteria

`[]GcpEventarcTriggerMatchingCriterion` · required

The event filters — ALL criteria must match for an event to be
delivered. Every trigger MUST filter the "type" attribute (the
CloudEvents type, e.g. google.cloud.pubsub.topic.v1.messagePublished or
google.cloud.audit.log.v1.written); the API's catalog of types and
their filterable attributes: https://cloud.google.com/eventarc/docs/reference/supported-events

- rule: matching_criteria must include a criterion for the 'type' attribute (the API requires it on every trigger)
- rule: {"repeated":{"minItems":"1"}}

### spec.matchingCriteria[].attribute

`string` · required

The CloudEvents attribute to filter (e.g. "type", "bucket",
"serviceName", "methodName"). Only the attributes the event type
declares filterable are accepted by the API.

- rule: {"required":true}

### spec.matchingCriteria[].value

`string` · required

The value the attribute must match. With operator
"match-path-pattern", path patterns like "objects/prefix/*" are
matched instead of exact equality.

- rule: {"required":true}

### spec.matchingCriteria[].operator

`string`

Empty for exact match (the default). The only other value the API
accepts is "match-path-pattern" (supported on a subset of attributes,
e.g. Cloud Storage object names and audit-log resourceName).

- rule: operator must be empty (exact match) or 'match-path-pattern' (the API's only other value)

### spec.destination

`GcpEventarcTriggerDestination` · required

Where matching events are delivered. Exactly one arm.

- rule: {"required":true}

### spec.destination.cloudRunService

`GcpEventarcTriggerCloudRunDestination`

Deliver to a Cloud Run service (the most common destination).

### spec.destination.cloudRunService.service

`string | valueFrom` · required

The Cloud Run service NAME (bare name, not a URL) — a literal or a
reference to a GcpCloudRun resource (its service_name output). Only
services in the same project as the trigger can be addressed.

- references: GcpCloudRun (`status.outputs.service_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpCloudRun, name: <that resource's name>, fieldPath: status.outputs.service_name}} -- a bare string does not parse

### spec.destination.cloudRunService.region

`string`

The region the Cloud Run service is deployed in. The API REQUIRES
this on every create (live-verified: 400 "cloud_run.region is empty"
— it never infers the region from the trigger). When empty, the
MODULE defaults it to the trigger's location; set it explicitly when
the service lives in a different region than the trigger, and always
on a "global" trigger (spec-enforced — global is not a Cloud Run
region).

### spec.destination.cloudRunService.path

`string`

The relative path on the service events are POSTed to (RFC2396 path
segment, e.g. "/events" or "route/subroute"). Empty means the root
path.

### spec.destination.gke

`GcpEventarcTriggerGkeDestination`

Deliver to a service running in a GKE cluster. Requires Eventarc's
GKE destination support to be enabled once per project
(gcloud eventarc gke-destinations init) — Eventarc then manages an
event-forwarder pod in the cluster.

### spec.destination.gke.cluster

`string | valueFrom` · required

The GKE cluster NAME the service runs in — a literal or a reference to
a GcpGkeCluster resource (its name output). Must be in the same
project as the trigger. The reference is containment-exempt: the
trigger DELIVERS INTO the cluster, it does not live inside it (the
trigger's home is its project).

- references: GcpGkeCluster (`status.outputs.name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpGkeCluster, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.destination.gke.location

`string` · required

The cluster's compute location: a zone (us-central1-a) for zonal
clusters or a region (us-central1) for regional clusters.

- rule: {"required":true}

### spec.destination.gke.namespace

`string` · required

The Kubernetes namespace the destination service lives in.

- rule: {"required":true}

### spec.destination.gke.service

`string` · required

The Kubernetes service NAME events are delivered to.

- rule: {"required":true}

### spec.destination.gke.path

`string`

The relative path on the service events are POSTed to (RFC2396 path
segment). Empty means the root path.

### spec.destination.workflow

`string | valueFrom`

Trigger a Cloud Workflows EXECUTION per event. The full workflow
resource name (projects/{project}/locations/{location}/workflows/{name})
— a literal or a reference to a GcpWorkflow resource (its workflow_id
output is exactly this value). Must live in the same project as the
trigger. This arm REQUIRES spec.service_account (spec-enforced; the
API rejects the create without it).

- references: GcpWorkflow (`status.outputs.workflow_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpWorkflow, name: <that resource's name>, fieldPath: status.outputs.workflow_id}} -- a bare string does not parse

### spec.destination.httpEndpoint

`GcpEventarcTriggerHttpEndpointDestination`

Deliver to a private HTTP endpoint reachable through a VPC network
attachment (Eventarc Standard's bring-your-own-HTTP destination).

### spec.destination.httpEndpoint.uri

`string` · required

The endpoint URI (RFC2396, e.g.
https://svc.us-central1.p.local:8080/route). Only HTTPS is supported.

- rule: {"required":true}

### spec.destination.httpEndpoint.networkAttachment

`string` · required

The network attachment that lets Eventarc reach the endpoint's VPC
(format:
projects/{project}/regions/{region}/networkAttachments/{name}).
REQUIRED for HTTP endpoint destinations — the provider models this as
a separate network_config block but permits it only with HTTP
endpoints, so this spec carries it inside the arm.

- rule: {"required":true}

### spec.serviceAccount

`string | valueFrom`

The IAM service account email the trigger runs as — it must hold
roles/eventarc.eventReceiver, plus roles/run.invoker for authenticated
Cloud Run destinations (identity tokens are minted from this account)
or roles/workflows.invoker for workflow destinations. Audit-log
triggers and workflow destinations REQUIRE a service account (the
workflow case is API-enforced at create — live-verified 400 without
it). A literal email or a reference to a GcpServiceAccount resource.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.transportPubsubTopic

`string | valueFrom`

For google.cloud.pubsub.topic.v1.messagePublished triggers ONLY: use
an EXISTING Pub/Sub topic (projects/{project}/topics/{topic} — a
literal or a reference to a GcpPubSubTopic resource) as the transport
instead of letting Eventarc create one. The topic is NOT deleted when
the trigger is destroyed (Eventarc only manages topics it created).

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.eventDataContentType

`string`

The MIME type the CloudEvent data field is delivered as (e.g.
application/json or application/protobuf). The API defaults to
application/json when unset. NOT accepted on Pub/Sub
(messagePublished) triggers — the API rejects any value there
(live-verified 400; a Pub/Sub event's payload format is decided by
the publisher). Spec-enforced above.

### spec.retryMaxAttempts

`int32`

Delivery retry ceiling. The provider accepts only the value 1
("The only valid value is 1" — its own schema note), which DISABLES
Eventarc's default retries: a failed delivery is not retried. Leave
unset (0) for the platform default retry behavior. Can only be set
with Cloud Run destinations (provider constraint, enforced above).

- rule: retry_max_attempts accepts only the value 1 (the provider's only valid value; unset means platform default retries)

### spec.labels

`map<string, string>`

User labels attached to the trigger (merged with the platform's
standard labels by the module).

### spec.partnerChannel

`GcpEventarcTriggerPartnerChannel`

Receive events from an Eventarc SaaS PARTNER (e.g. Datadog): the
module creates the partner channel alongside the trigger and wires the
trigger to it. The channel's activation_token output must be
handed to the partner to complete the handshake — until then the
channel stays PENDING and delivers nothing.

### spec.partnerChannel.channelName

`string`

The channel name in GCP. Defaults to "{trigger name}-channel" when
left empty. Immutable: changing it replaces the channel (a NEW
activation token — redo the partner handshake).

### spec.partnerChannel.thirdPartyProvider

`string` · required

The partner provider the channel receives events from (format:
projects/{project}/locations/{location}/providers/{provider_id} —
list available partners with gcloud eventarc providers list).
Immutable: changing it replaces the channel.

- rule: {"required":true}

### spec.partnerChannel.cryptoKey

`string | valueFrom`

CMEK for events in transit through the channel. The full crypto key
resource name — a literal or a reference to a GcpKmsKey resource.
Omit for Google-managed encryption.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.googleChannelCryptoKey

`string | valueFrom`

CMEK for the project/location's GOOGLE channel — the shared conduit
ALL non-partner triggers in this project+location deliver through.
The full crypto key resource name — a literal or a reference to a
GcpKmsKey resource. The module manages the per-project-per-location
googleChannelConfig SINGLETON: set this from AT MOST ONE trigger per
project+location (a second manager fights over the same singleton).
Deleting the config is a state-only no-op in GCP — the singleton
always exists; clearing this field reverts it to Google-managed
encryption.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed (also
applied to the partner channel and google-channel config the kind
manages):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the trigger (and any partner channel) is deleted;
               events stop being delivered immediately
  "PREVENT" -- destroy FAILS; protects a production event route
  "ABANDON" -- resources are removed from management but keep
               delivering in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `exactly_one_destination`: set exactly one destination: cloud_run_service, gke, workflow, or http_endpoint
- `retry_requires_cloud_run_destination`: retry_max_attempts can only be set with a cloud_run_service destination (the provider's own constraint)
- `workflow_destination_requires_service_account`: a workflow destination requires service_account (the API rejects the create without it; the account needs roles/eventarc.eventReceiver + roles/workflows.invoker)
- `global_trigger_cloud_run_region_required`: a cloud_run_service destination on a global trigger needs an explicit region (regional triggers default to their own location)
- `pubsub_trigger_forbids_event_data_content_type`: event_data_content_type cannot be set on a Pub/Sub (messagePublished) trigger — the API rejects it; the payload format is decided at publish time

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpEventarcTrigger, name: <resource-name>, fieldPath: status.outputs.<output>}`. A sensitive output is a secret the resource generates: on Planton it is kept in the organization's secret store and the output holds a `$secret/` reference, so feed it only to a sensitive field.

| Output | Type | Description |
|---|---|---|
| `status.outputs.trigger_name` | `string` | The trigger name as it exists in GCP. |
| `status.outputs.partner_channel_activation_token` | `string` (sensitive) | Partner-channel triggers only: the one-time activation token the SaaS partner needs to complete the channel handshake (hand it to the partner's console/API; the channel stays PENDING until then). Empty for non-partner triggers. Sensitive — treat like a credential. |
| `status.outputs.trigger_id` | `string` | The full trigger resource name (projects/{project}/locations/{location}/triggers/{name}) — the trigger's canonical API handle. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.destination.cloudRunService.service` | GcpCloudRun | `status.outputs.service_name` |
| `spec.destination.gke.cluster` | GcpGkeCluster | `status.outputs.name` |
| `spec.destination.workflow` | GcpWorkflow | `status.outputs.workflow_id` |
| `spec.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.transportPubsubTopic` | GcpPubSubTopic | `status.outputs.topic_id` |
| `spec.partnerChannel.cryptoKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.googleChannelCryptoKey` | GcpKmsKey | `status.outputs.key_id` |

## See Also

- [Overview](../README.md)
