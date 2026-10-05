# GCP Eventarc Trigger

Creates an Eventarc trigger — the routing rule "when THIS event happens, call THAT service": events matching the criteria (a Pub/Sub message, a Cloud Storage object change, an audit-log entry, a SaaS partner event) are delivered as CloudEvents to a Cloud Run service, a GKE service, a Workflow execution, or a private HTTP endpoint.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Eventarc trigger** -- an `eventarc.Trigger` with the configured criteria, destination, identity, and transport
- **Partner channel** (when `partnerChannel` is set) -- an `eventarc.Channel` the trigger is wired to, with its one-time activation token exported
- **Google channel config** (when `googleChannelCryptoKey` is set) -- the per-project-per-location CMEK singleton
- **Eventarc API enablement** -- `eventarc.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** to host the trigger (directly or via a GcpProject reference).
- **IAM**: see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying identity needs; the trigger's `serviceAccount` needs `roles/eventarc.eventReceiver`, plus `roles/run.invoker` for authenticated Cloud Run destinations.
- **First trigger in a project**: Eventarc's service agent is provisioned on first use — expect a few minutes before the first delivery succeeds.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpEventarcTrigger
metadata:
  name: order-events
spec:
  location: us-central1
  matchingCriteria:
    - attribute: type
      value: google.cloud.pubsub.topic.v1.messagePublished
  destination:
    cloudRunService:
      service:
        value: order-processor
  serviceAccount:
    value: eventarc-invoker@my-project.iam.gserviceaccount.com
```

```shell
planton apply -f trigger.yaml
```

## Outputs

| Output | Description |
|--------|-------------|
| `trigger_name` | The trigger name in GCP |
| `trigger_id` | Full trigger resource name — the canonical API handle |
| `partner_channel_activation_token` | Partner triggers only: the one-time partner handshake token (sensitive) |

## Works With

- **GcpCloudRun** -- the most common destination (`destination.cloudRunService.service`)
- **GcpWorkflow** -- `destination.workflow` starts an execution per event
- **GcpGkeCluster** -- GKE service destinations
- **GcpPubSubTopic** -- bring-your-own transport topic for messagePublished triggers
- **GcpServiceAccount** -- the trigger's identity
- **GcpKmsKey** -- CMEK for partner channels and the google channel config

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
