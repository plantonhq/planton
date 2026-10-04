# GcpEventarcMessageBus

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpEventarcMessageBusSpec defines an Eventarc ADVANCED message bus with
its satellites — the enterprise eventing hub, modeled as one kind
because the satellites exist only in service of their bus:

  - the BUS is the central conduit messages flow through;
  - GOOGLE API SOURCES feed it (they publish Google-service events INTO
    the bus — each is auto-wired to this kind's own bus, never someone
    else's);
  - PIPELINES deliver messages OUT (to an HTTP endpoint, a Pub/Sub
    topic, a Workflow, or another bus), with per-pipeline auth, payload
    conversion, transformation, and retry policy;
  - ENROLLMENTS bind the two: a CEL match expression selects messages
    from the bus and routes them to one of this kind's pipelines
    (referenced by its pipeline_id — the module renders the full
    resource name).

Eventarc Advanced is distinct from Eventarc Standard (GcpEventarcTrigger)
— Standard routes single event types point-to-point; Advanced is the
many-sources-many-destinations hub with mediation. Advanced is available
in a subset of regions; the API decides at create time.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpEventarcMessageBus
metadata:
  name: my-sample-message-bus
spec:
  # Eventarc Advanced serves a subset of regions — the API decides at
  # create time.
  location: us-central1

  # Platform-log verbosity: INFO is the onboarding posture.
  logSeverity: INFO

  # What a destroy does: DELETE (undelivered messages are lost),
  # PREVENT (the posture once production consumers depend on the hub),
  # or ABANDON (keep the family running unmanaged).
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.messageBusId` | `string` |  |  |  |
| `spec.displayName` | `string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.annotations` | `map<string, string>` |  |  |  |
| `spec.cryptoKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.logSeverity` | `string` |  |  |  |
| `spec.googleApiSources` | `[]GcpEventarcMessageBusGoogleApiSource` |  |  |  |
| `spec.googleApiSources[].sourceId` | `string` | yes |  |  |
| `spec.googleApiSources[].displayName` | `string` |  |  |  |
| `spec.googleApiSources[].labels` | `map<string, string>` |  |  |  |
| `spec.googleApiSources[].annotations` | `map<string, string>` |  |  |  |
| `spec.googleApiSources[].cryptoKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.googleApiSources[].logSeverity` | `string` |  |  |  |
| `spec.pipelines` | `[]GcpEventarcMessageBusPipeline` |  |  |  |
| `spec.pipelines[].pipelineId` | `string` | yes |  |  |
| `spec.pipelines[].destination` | `GcpEventarcMessageBusPipelineDestination` | yes |  |  |
| `spec.pipelines[].destination.httpEndpoint` | `GcpEventarcMessageBusHttpEndpoint` |  |  |  |
| `spec.pipelines[].destination.httpEndpoint.uri` | `string` | yes |  |  |
| `spec.pipelines[].destination.httpEndpoint.messageBindingTemplate` | `string` |  |  |  |
| `spec.pipelines[].destination.httpEndpoint.networkAttachment` | `string` |  |  |  |
| `spec.pipelines[].destination.topic` | `string \| valueFrom` |  |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.pipelines[].destination.workflow` | `string \| valueFrom` |  |  | GcpWorkflow (`status.outputs.workflow_id`) |
| `spec.pipelines[].destination.messageBus` | `string` |  |  |  |
| `spec.pipelines[].authentication` | `GcpEventarcMessageBusPipelineAuthentication` |  |  |  |
| `spec.pipelines[].authentication.googleOidc` | `GcpEventarcMessageBusOidcAuth` |  |  |  |
| `spec.pipelines[].authentication.googleOidc.serviceAccount` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.pipelines[].authentication.googleOidc.audience` | `string` |  |  |  |
| `spec.pipelines[].authentication.oauthToken` | `GcpEventarcMessageBusOauthAuth` |  |  |  |
| `spec.pipelines[].authentication.oauthToken.serviceAccount` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.pipelines[].authentication.oauthToken.scope` | `string` |  |  |  |
| `spec.pipelines[].inputPayloadFormat` | `GcpEventarcMessageBusPayloadFormat` |  |  |  |
| `spec.pipelines[].inputPayloadFormat.avro` | `GcpEventarcMessageBusSchemaFormat` |  |  |  |
| `spec.pipelines[].inputPayloadFormat.avro.schemaDefinition` | `string` |  |  |  |
| `spec.pipelines[].inputPayloadFormat.json` | `bool` |  |  |  |
| `spec.pipelines[].inputPayloadFormat.protobuf` | `GcpEventarcMessageBusSchemaFormat` |  |  |  |
| `spec.pipelines[].inputPayloadFormat.protobuf.schemaDefinition` | `string` |  |  |  |
| `spec.pipelines[].outputPayloadFormat` | `GcpEventarcMessageBusPayloadFormat` |  |  |  |
| `spec.pipelines[].outputPayloadFormat.avro` | `GcpEventarcMessageBusSchemaFormat` |  |  |  |
| `spec.pipelines[].outputPayloadFormat.avro.schemaDefinition` | `string` |  |  |  |
| `spec.pipelines[].outputPayloadFormat.json` | `bool` |  |  |  |
| `spec.pipelines[].outputPayloadFormat.protobuf` | `GcpEventarcMessageBusSchemaFormat` |  |  |  |
| `spec.pipelines[].outputPayloadFormat.protobuf.schemaDefinition` | `string` |  |  |  |
| `spec.pipelines[].mediationTransformationTemplate` | `string` |  |  |  |
| `spec.pipelines[].retryPolicy` | `GcpEventarcMessageBusPipelineRetryPolicy` |  |  |  |
| `spec.pipelines[].retryPolicy.maxAttempts` | `int32` |  |  |  |
| `spec.pipelines[].retryPolicy.minRetryDelay` | `string` |  |  |  |
| `spec.pipelines[].retryPolicy.maxRetryDelay` | `string` |  |  |  |
| `spec.pipelines[].displayName` | `string` |  |  |  |
| `spec.pipelines[].labels` | `map<string, string>` |  |  |  |
| `spec.pipelines[].annotations` | `map<string, string>` |  |  |  |
| `spec.pipelines[].cryptoKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.pipelines[].logSeverity` | `string` |  |  |  |
| `spec.enrollments` | `[]GcpEventarcMessageBusEnrollment` |  |  |  |
| `spec.enrollments[].enrollmentId` | `string` | yes |  |  |
| `spec.enrollments[].celMatch` | `string` | yes |  |  |
| `spec.enrollments[].pipeline` | `string` | yes |  |  |
| `spec.enrollments[].displayName` | `string` |  |  |  |
| `spec.enrollments[].labels` | `map<string, string>` |  |  |  |
| `spec.enrollments[].annotations` | `map<string, string>` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project to create the bus (and all satellites) in. Can be a
literal project ID or a reference to a GcpProject resource. If
omitted, the provider's default project is used.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

The region for the bus and every satellite (e.g. us-central1).
Eventarc Advanced serves a subset of regions
(https://cloud.google.com/eventarc/docs/locations) — the API rejects
unsupported ones at create time. Immutable.

- rule: {"required":true}

### spec.messageBusId

`string`

The bus ID in GCP (also its short name). Defaults to metadata.name
when left empty. Format: 1–63 chars, lowercase letters, digits, and
hyphens; must start with a letter and end alphanumeric (the API's
documented pattern). Immutable: changing it replaces the bus.

- rule: message_bus_id must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and ending alphanumeric

### spec.displayName

`string`

Display name shown in the console.

### spec.labels

`map<string, string>`

User labels attached to the bus (merged with the platform's standard
labels by the module).

### spec.annotations

`map<string, string>`

Free-form annotations attached to the bus
(https://google.aip.dev/128#annotations) — non-identifying metadata
for tools; unlike labels they are not usable in filters or billing.

### spec.cryptoKey

`string | valueFrom`

CMEK for messages at rest in the bus. The full crypto key resource
name (projects/{p}/locations/{l}/keyRings/{r}/cryptoKeys/{k}) — a
literal or a reference to a GcpKmsKey resource. The key must be in the
same region as the bus; grant the Eventarc service agent
roles/cloudkms.cryptoKeyEncrypterDecrypter BEFORE creating. Omit for
Google-managed encryption.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.logSeverity

`string`

The minimum severity of bus activity recorded to Cloud Logging.
Empty uses the API default (NONE — no platform logs). One of: NONE,
DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY
(the provider's ValidateEnum list). INFO is the operational sweet
spot while onboarding sources and pipelines.

- rule: log_severity must be one of: NONE, DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY

### spec.googleApiSources

`[]GcpEventarcMessageBusGoogleApiSource`

Google API sources publishing Google-service events INTO this bus.
Each becomes a google_api_source resource whose destination the module
wires to THIS bus (byte-identically on both IaC engines) — an api
source feeding someone else's bus belongs to that bus's kind instance.

### spec.googleApiSources[].sourceId

`string` · required

The source ID in GCP. Format: 1–63 chars, lowercase letters, digits,
and hyphens; starts with a letter, ends alphanumeric. Immutable:
changing it replaces the source.

- rule: source_id must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and ending alphanumeric
- rule: {"required":true}

### spec.googleApiSources[].displayName

`string`

Display name shown in the console.

### spec.googleApiSources[].labels

`map<string, string>`

User labels attached to the source.

### spec.googleApiSources[].annotations

`map<string, string>`

Free-form annotations attached to the source.

### spec.googleApiSources[].cryptoKey

`string | valueFrom`

CMEK for this source's data. The full crypto key resource name — a
literal or a reference to a GcpKmsKey resource. Omit for
Google-managed encryption.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.googleApiSources[].logSeverity

`string`

The minimum severity of this source's activity recorded to Cloud
Logging (same value set as the bus's log_severity).

- rule: log_severity must be one of: NONE, DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY

### spec.pipelines

`[]GcpEventarcMessageBusPipeline`

Pipelines delivering messages OUT of the bus. Referenced by
enrollments below via pipeline_id.

- rule: set exactly one destination target: http_endpoint, topic, workflow, or message_bus
- rule: http_endpoint destinations require network_attachment (the provider's own rule)

### spec.pipelines[].pipelineId

`string` · required

The pipeline ID in GCP. Format: 1–63 chars, lowercase letters, digits,
and hyphens; starts with a letter, ends alphanumeric. Immutable:
changing it replaces the pipeline.

- rule: pipeline_id must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and ending alphanumeric
- rule: {"required":true}

### spec.pipelines[].destination

`GcpEventarcMessageBusPipelineDestination` · required

Where this pipeline delivers. Exactly one target arm.

- rule: {"required":true}

### spec.pipelines[].destination.httpEndpoint

`GcpEventarcMessageBusHttpEndpoint`

Deliver to an HTTP endpoint reachable through a VPC network
attachment.

### spec.pipelines[].destination.httpEndpoint.uri

`string` · required

The endpoint URI (RFC2396). Only HTTPS is supported (the API rejects
http://), e.g. https://svc.us-central1.p.local:8080/route.

- rule: http_endpoint uri must use https:// (the API supports only the HTTPS protocol)
- rule: {"required":true}

### spec.pipelines[].destination.httpEndpoint.messageBindingTemplate

`string`

A CEL expression shaping the outgoing HTTP request (headers, body
binding) — see
https://cloud.google.com/eventarc/advanced/docs/receive-events/create-message-binding.
Empty uses the default CloudEvents HTTP binding.

### spec.pipelines[].destination.httpEndpoint.networkAttachment

`string`

The network attachment that lets the pipeline reach the endpoint's
VPC (format:
projects/{project}/regions/{region}/networkAttachments/{name}).
REQUIRED for HTTP endpoint destinations; forbidden for every other
target (the provider's own rule, enforced at the pipeline level).

### spec.pipelines[].destination.topic

`string | valueFrom`

Publish to a Pub/Sub topic. A literal topic resource name or a
reference to a GcpPubSubTopic resource (its topic_id output,
projects/{project}/topics/{topic}). The provider documents the form
projects/{project}/locations/{location}/topics/{topic} for this field;
both name the same topic — the live API accepts the canonical Pub/Sub
form (live-verified: a pipeline created with the canonical form went
ACTIVE and re-planned clean on both engines).

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.pipelines[].destination.workflow

`string | valueFrom`

Trigger a Cloud Workflows EXECUTION per message. The full workflow
resource name — a literal or a reference to a GcpWorkflow resource
(its workflow_id output). Must be deployed in the same project as the
pipeline.

- references: GcpWorkflow (`status.outputs.workflow_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpWorkflow, name: <that resource's name>, fieldPath: status.outputs.workflow_id}} -- a bare string does not parse

### spec.pipelines[].destination.messageBus

`string`

Chain into ANOTHER message bus (cross-bus routing). The full bus
resource name
(projects/{project}/locations/{location}/messageBuses/{bus}) — e.g.
another GcpEventarcMessageBus's message_bus_name output. Must be in
the same project as the pipeline.

### spec.pipelines[].authentication

`GcpEventarcMessageBusPipelineAuthentication`

How the pipeline authenticates to the destination (HTTP endpoints
that verify identity). At most one mechanism.

- rule: set at most one of google_oidc or oauth_token

### spec.pipelines[].authentication.googleOidc

`GcpEventarcMessageBusOidcAuth`

Authenticate with a Google OIDC ID token — for endpoints that verify
Google-signed identity tokens (Cloud Run, IAP-protected services).

### spec.pipelines[].authentication.googleOidc.serviceAccount

`string | valueFrom` · required

The service account email the token is minted for — a literal or a
reference to a GcpServiceAccount resource. The Eventarc service agent
must hold roles/iam.serviceAccountTokenCreator on it.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.pipelines[].authentication.googleOidc.audience

`string`

The audience claim in the minted token. Empty uses the destination
URI.

### spec.pipelines[].authentication.oauthToken

`GcpEventarcMessageBusOauthAuth`

Authenticate with an OAuth access token — preferred for Google APIs
that accept OAuth.

### spec.pipelines[].authentication.oauthToken.serviceAccount

`string | valueFrom` · required

The service account email the token is minted for — a literal or a
reference to a GcpServiceAccount resource. The Eventarc service agent
must hold roles/iam.serviceAccountTokenCreator on it.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.pipelines[].authentication.oauthToken.scope

`string`

The OAuth scope in the minted token. Empty uses
https://www.googleapis.com/auth/cloud-platform.

### spec.pipelines[].inputPayloadFormat

`GcpEventarcMessageBusPayloadFormat`

The payload format messages ARRIVE in (from the bus). Set input and
output together to convert between formats; leave both unset to pass
payloads through untouched.

- rule: set exactly one payload format: avro, json, or protobuf

### spec.pipelines[].inputPayloadFormat.avro

`GcpEventarcMessageBusSchemaFormat`

Avro format with its schema definition.

### spec.pipelines[].inputPayloadFormat.avro.schemaDefinition

`string`

The schema definition text (an Avro schema JSON or a protobuf
definition). Required by the API when converting formats.

### spec.pipelines[].inputPayloadFormat.json

`bool`

JSON format (no schema).

### spec.pipelines[].inputPayloadFormat.protobuf

`GcpEventarcMessageBusSchemaFormat`

Protobuf format with its schema definition.

### spec.pipelines[].inputPayloadFormat.protobuf.schemaDefinition

`string`

The schema definition text (an Avro schema JSON or a protobuf
definition). Required by the API when converting formats.

### spec.pipelines[].outputPayloadFormat

`GcpEventarcMessageBusPayloadFormat`

The payload format messages are DELIVERED in. Avro/Protobuf require a
schema_definition; converting between Avro and Protobuf requires both
sides to define schemas.

- rule: set exactly one payload format: avro, json, or protobuf

### spec.pipelines[].outputPayloadFormat.avro

`GcpEventarcMessageBusSchemaFormat`

Avro format with its schema definition.

### spec.pipelines[].outputPayloadFormat.avro.schemaDefinition

`string`

The schema definition text (an Avro schema JSON or a protobuf
definition). Required by the API when converting formats.

### spec.pipelines[].outputPayloadFormat.json

`bool`

JSON format (no schema).

### spec.pipelines[].outputPayloadFormat.protobuf

`GcpEventarcMessageBusSchemaFormat`

Protobuf format with its schema definition.

### spec.pipelines[].outputPayloadFormat.protobuf.schemaDefinition

`string`

The schema definition text (an Avro schema JSON or a protobuf
definition). Required by the API when converting formats.

### spec.pipelines[].mediationTransformationTemplate

`string`

A CEL expression rewriting the message before delivery (e.g.
'message.removeFields(["data.secret"])'). The API allows at most ONE
mediation (transformation) per pipeline — hence a single template
rather than a list.

### spec.pipelines[].retryPolicy

`GcpEventarcMessageBusPipelineRetryPolicy`

Delivery retry policy. Unset uses the API defaults (5 attempts,
5s–60s exponential backoff).

### spec.pipelines[].retryPolicy.maxAttempts

`int32`

Maximum delivery attempts, 1–100 (the API's documented range).
0 means unset — the API default of 5.

- rule: max_attempts must be between 1 and 100 (0 leaves the API default of 5)

### spec.pipelines[].retryPolicy.minRetryDelay

`string`

Minimum wait between attempts, in seconds-suffixed form (e.g. "5s").
The API accepts 1–600 seconds; its default is 5s.

- rule: min_retry_delay must be a seconds-suffixed duration like '5s' (the API accepts 1-600 seconds)

### spec.pipelines[].retryPolicy.maxRetryDelay

`string`

Maximum wait between attempts, in seconds-suffixed form (e.g. "60s").
The API accepts 1–600 seconds; its default is 60s. Note the API quirk
its docs call out: when setting min and max, they may be required to
be EQUAL on some surfaces — live behavior decides.

- rule: max_retry_delay must be a seconds-suffixed duration like '60s' (the API accepts 1-600 seconds)

### spec.pipelines[].displayName

`string`

Display name shown in the console.

### spec.pipelines[].labels

`map<string, string>`

User labels attached to the pipeline.

### spec.pipelines[].annotations

`map<string, string>`

Free-form annotations attached to the pipeline.

### spec.pipelines[].cryptoKey

`string | valueFrom`

CMEK for this pipeline's data. The full crypto key resource name — a
literal or a reference to a GcpKmsKey resource. Omit for
Google-managed encryption.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.pipelines[].logSeverity

`string`

The minimum severity of this pipeline's activity recorded to Cloud
Logging (same value set as the bus's log_severity).

- rule: log_severity must be one of: NONE, DEBUG, INFO, NOTICE, WARNING, ERROR, CRITICAL, ALERT, EMERGENCY

### spec.enrollments

`[]GcpEventarcMessageBusEnrollment`

Enrollments — the routing table: each selects messages from the bus
with a CEL expression and delivers them to one of this spec's
pipelines.

### spec.enrollments[].enrollmentId

`string` · required

The enrollment ID in GCP. Format: 1–63 chars, lowercase letters,
digits, and hyphens; starts with a letter, ends alphanumeric.
Immutable: changing it replaces the enrollment.

- rule: enrollment_id must be 1-63 characters of lowercase letters, digits, and hyphens, starting with a letter and ending alphanumeric
- rule: {"required":true}

### spec.enrollments[].celMatch

`string` · required

The CEL expression selecting which bus messages this enrollment
routes (evaluated against the CloudEvent, e.g.
message.type == "google.cloud.storage.object.v1.finalized"). "true"
routes everything.

- rule: {"required":true}

### spec.enrollments[].pipeline

`string` · required

The pipeline_id of the pipeline (defined in this spec) that delivers
the selected messages. The module renders the full pipeline resource
name the API demands; a spec-level rule rejects ids that match no
sibling pipeline.

- rule: {"required":true}

### spec.enrollments[].displayName

`string`

Display name shown in the console.

### spec.enrollments[].labels

`map<string, string>`

User labels attached to the enrollment.

### spec.enrollments[].annotations

`map<string, string>`

Free-form annotations attached to the enrollment.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed
(applied to the bus and every satellite):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- bus, sources, enrollments, and pipelines are deleted;
               undelivered messages are lost
  "PREVENT" -- destroy FAILS; protects a production eventing hub
  "ABANDON" -- resources are removed from management but keep running
               in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `enrollment_pipeline_must_exist`: every enrollment's pipeline must match the pipeline_id of a pipeline defined in this spec

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpEventarcMessageBus, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.message_bus_name` | `string` | The full bus resource name (projects/{project}/locations/{location}/messageBuses/{bus}) — the value cross-bus pipelines (another bus's destination.message_bus) and external publishers consume. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.cryptoKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.googleApiSources[].cryptoKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.pipelines[].destination.topic` | GcpPubSubTopic | `status.outputs.topic_id` |
| `spec.pipelines[].destination.workflow` | GcpWorkflow | `status.outputs.workflow_id` |
| `spec.pipelines[].authentication.googleOidc.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.pipelines[].authentication.oauthToken.serviceAccount` | GcpServiceAccount | `status.outputs.email` |
| `spec.pipelines[].cryptoKey` | GcpKmsKey | `status.outputs.key_id` |

## See Also

- [Overview](../README.md)
