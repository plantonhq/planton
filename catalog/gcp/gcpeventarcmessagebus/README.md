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
- **IAM**: see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying identity needs.
- **Region**: Eventarc Advanced serves a [subset of regions](https://cloud.google.com/eventarc/docs/locations) — the API rejects unsupported ones at create time.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpEventarcMessageBus
metadata:
  name: central-bus
spec:
  location: us-central1
  logSeverity: INFO
  pipelines:
    - pipelineId: deliver-to-topic
      destination:
        topic:
          value: projects/my-project/topics/downstream
  enrollments:
    - enrollmentId: route-everything
      celMatch: "true"
      pipeline: deliver-to-topic
```

```shell
planton apply -f message-bus.yaml
```

## Outputs

| Output | Description |
|--------|-------------|
| `message_bus_name` | Full bus resource name — the cross-bus / external-publisher handle |

## Works With

- **GcpPubSubTopic** -- pipeline topic destinations
- **GcpWorkflow** -- pipeline workflow destinations (an execution per message)
- **GcpServiceAccount** -- pipeline authentication identities (OIDC / OAuth)
- **GcpKmsKey** -- CMEK on the bus and every satellite
- **GcpProject** -- provides the GCP project the family lives in

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
