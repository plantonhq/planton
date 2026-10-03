# GCP Eventarc Message Bus

Creates an Eventarc ADVANCED message bus with its satellites — the enterprise eventing hub: Google API sources publish events INTO the bus, enrollments select messages with CEL expressions, and pipelines deliver them OUT (to HTTP endpoints, Pub/Sub topics, Workflows, or other buses) with per-pipeline auth, payload conversion, transformation, and retries. Distinct from Eventarc Standard (GcpEventarcTrigger): Standard routes single event types point-to-point; Advanced is the many-sources-many-destinations hub with mediation.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Message bus** -- an `eventarc.MessageBus` (the central conduit)
- **Google API sources** -- one `eventarc.GoogleApiSource` per spec entry, auto-wired to THIS bus
- **Pipelines** -- one `eventarc.Pipeline` per spec entry (destination, auth, payload formats, mediation, retries)
- **Enrollments** -- one `eventarc.Enrollment` per spec entry, binding CEL-selected messages to a sibling pipeline
- **Eventarc API enablement** -- `eventarc.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** to host the family (directly or via a GcpProject reference).
- **IAM**: the deploying identity needs `roles/eventarc.admin`.
- **Region**: Eventarc Advanced serves a subset of regions — the API rejects unsupported ones at create time.
- **CMEK grant** (only for `cryptoKey`) — the key must be in the same region as the bus, and the Eventarc service agent needs `roles/cloudkms.cryptoKeyEncrypterDecrypter` on it before creating.

## Deploy

### Console

Open the deployment store, find **GCP Eventarc Message Bus**, and click **Deploy**. The creation wizard walks you through project and location, the bus itself, then its satellites: Google API sources, pipelines with their destinations and auth, and the enrollments that bind them. Start from the **Bus with Topic Pipeline** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpEventarcMessageBus
metadata:
  name: central-bus
  org: acme-corp
  env: prod
spec:
  location: us-central1
  logSeverity: INFO
  pipelines:
    - pipelineId: deliver-to-topic
      destination:
        topic:
          value: projects/acme-prod/topics/downstream
  enrollments:
    - enrollmentId: route-everything
      celMatch: "true"
      pipeline: deliver-to-topic
```

```shell
planton apply -f message-bus.yaml
```

This creates the bus, one pipeline publishing to a Pub/Sub topic, and one enrollment routing every bus message to it. An Infra Job tracks the provisioning in real time.

### InfraChart

When the pipeline's destination is a topic deployed in the same InfraPipeline, wire it with ValueFromRef:

```yaml
spec:
  location: us-central1
  pipelines:
    - pipelineId: deliver-to-topic
      destination:
        topic:
          valueFrom:
            kind: GcpPubSubTopic
            name: downstream-topic
            fieldPath: status.outputs.topic_id
  enrollments:
    - enrollmentId: route-everything
      celMatch: "true"
      pipeline: deliver-to-topic
```

The InfraPipeline deploys the topic first, then provisions the bus family with the resolved topic name.

## Key Configuration

These are the most important decisions when configuring a message bus. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**googleApiSources** -- publish Google-service events into the bus. Each source is AUTO-WIRED to this kind's own bus — a source feeding another bus belongs to that bus's manifest.

**enrollments** -- the routing table: `celMatch` selects messages (evaluated against the CloudEvent; `"true"` routes everything), `pipeline` names a sibling pipeline by its `pipelineId` (validated at manifest time — dangling ids never reach the API).

**pipelines** -- one destination each (API truth): an HTTPS endpoint via a VPC network attachment, a Pub/Sub topic, a Workflow, or another bus. Per-pipeline `authentication` (OIDC for Cloud Run/IAP, OAuth for Google APIs), avro/json/protobuf payload conversion, a single CEL `mediationTransformationTemplate`, and 1–100 attempt retries with 1–600s backoff.

**logSeverity** -- platform-log verbosity per resource. Empty means the API default (NONE — no platform logs at all): run `INFO` while onboarding sources and pipelines, then scale back.

**deletionPolicy** -- one lever applied to the bus and every satellite. The default (`DELETE`) removes everything on destroy, losing undelivered messages; `PREVENT` makes destroy fail — the posture for a production hub; `ABANDON` unmanages the resources but leaves them running in GCP.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpPubSubTopic** (pipeline arm) | `pipelines[].destination.topic` | `status.outputs.topic_id` |
| **GcpWorkflow** (pipeline arm) | `pipelines[].destination.workflow` | `status.outputs.workflow_id` |
| **GcpServiceAccount** (auth) | `pipelines[].authentication.*.serviceAccount` | `status.outputs.email` |
| **GcpKmsKey** (optional) | `cryptoKey` (bus/source/pipeline) | `status.outputs.key_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `message_bus_name` | Full bus resource name | Another bus's `pipelines[].destination.messageBus` (cross-bus chaining); external publishers |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Bus with topic pipeline** -- everything into the bus, everything out to Pub/Sub; the smallest useful hub. Start from the **Bus with Topic Pipeline** preset.

**Audit fan-out** -- an audit-log API source with CEL-split enrollments routing storage events and IAM events to different pipelines. Start from the **Audit Fan-Out** preset.

## Works With

- [**GCP Pub/Sub Topic**](/infra-catalog/gcp-pub-sub-topic) -- pipeline topic destinations
- [**GCP Workflow**](/infra-catalog/gcp-workflow) -- pipeline workflow destinations
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- pipeline auth identities
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- CMEK on bus and satellites
- [**GCP Eventarc Trigger**](/infra-catalog/gcp-eventarc-trigger) -- Eventarc Standard, for single-type point-to-point routes
