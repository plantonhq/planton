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
- **IAM**: the deploying identity needs `roles/eventarc.admin` (or narrower); the trigger's `serviceAccount` needs `roles/eventarc.eventReceiver`, plus `roles/run.invoker` for authenticated Cloud Run destinations.

## Deploy

### Console

Open the deployment store, find **GCP Eventarc Trigger**, and click **Deploy**. The creation wizard walks you through project and location, the matching criteria, the destination arm, and the trigger's service-account identity. Start from the **Pub/Sub to Cloud Run** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpEventarcTrigger
metadata:
  name: order-events
  org: acme-corp
  env: prod
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
    value: eventarc-invoker@acme-prod.iam.gserviceaccount.com
```

```shell
planton apply -f trigger.yaml
```

This creates a trigger that delivers every Pub/Sub message published to its Eventarc-minted transport topic to the `order-processor` Cloud Run service in the same region. An Infra Job tracks the provisioning in real time.

### InfraChart

When the destination and identity are deployed in the same InfraPipeline, wire them with ValueFromRef:

```yaml
spec:
  location: us-central1
  matchingCriteria:
    - attribute: type
      value: google.cloud.pubsub.topic.v1.messagePublished
  destination:
    cloudRunService:
      service:
        valueFrom:
          kind: GcpCloudRun
          name: order-processor
          fieldPath: status.outputs.service_name
  serviceAccount:
    valueFrom:
      kind: GcpServiceAccount
      name: eventarc-invoker
      fieldPath: status.outputs.email
```

The InfraPipeline deploys the Cloud Run service and service account first, then provisions the trigger against them.

## Key Configuration

These are the most important decisions when configuring a trigger. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**matchingCriteria** -- ALL criteria must match. Every trigger MUST filter the `type` attribute (the CloudEvents type, e.g. `google.cloud.pubsub.topic.v1.messagePublished` or `google.cloud.audit.log.v1.written`); only the attributes the event type declares filterable are accepted. `match-path-pattern` is the only non-exact operator (Storage object names, audit-log resourceName).

**destination** -- exactly one arm: `cloudRunService` (the workhorse), `gke` (needs one-time `gcloud eventarc gke-destinations init` per project), `workflow` (references a GcpWorkflow), or `httpEndpoint` (private endpoints via a VPC network attachment).

**serviceAccount** -- the trigger's identity: `roles/eventarc.eventReceiver` always; `roles/run.invoker` for authenticated Cloud Run destinations; REQUIRED for audit-log triggers.

**transportPubsubTopic** -- messagePublished triggers only: consume an EXISTING topic instead of letting Eventarc mint one. The topic is never deleted with the trigger.

**partnerChannel** -- receive events from a SaaS partner: the module creates the channel and wires the trigger; hand the `partner_channel_activation_token` output to the partner to complete the handshake — the channel stays PENDING and delivers nothing until then. The channel name is immutable: changing it replaces the channel and mints a NEW token, redoing the handshake.

**googleChannelCryptoKey** -- CMEK for the shared Google channel ALL non-partner triggers in this project+location deliver through. The module manages the per-project-per-location singleton: set this from AT MOST ONE trigger per project+location — a second manager fights over the same singleton.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpCloudRun** (destination arm) | `destination.cloudRunService.service` | `status.outputs.service_name` |
| **GcpGkeCluster** (destination arm) | `destination.gke.cluster` | `status.outputs.name` |
| **GcpWorkflow** (destination arm) | `destination.workflow` | `status.outputs.workflow_id` |
| **GcpPubSubTopic** (optional) | `transportPubsubTopic` | `status.outputs.topic_id` |
| **GcpServiceAccount** (optional) | `serviceAccount` | `status.outputs.email` |
| **GcpKmsKey** (optional) | `partnerChannel.cryptoKey`, `googleChannelCryptoKey` | `status.outputs.key_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `trigger_name` | The trigger name | gcloud commands, debugging |
| `trigger_id` | Full trigger resource name | Monitoring filters, the canonical API handle |
| `partner_channel_activation_token` | One-time partner handshake token (sensitive) | Handed to the SaaS partner's console |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Pub/Sub to Cloud Run** -- the workhorse: a message published to a topic invokes a service. Start from the **Pub/Sub to Cloud Run** preset.

**Audit-log to Workflow** -- react to control-plane changes (a bucket created, an IAM grant) with an orchestrated response. Start from the **Audit Log to Workflow** preset.

## Works With

- [**GCP Cloud Run**](/infra-catalog/gcp-cloud-run) -- the most common destination
- [**GCP Workflow**](/infra-catalog/gcp-workflow) -- orchestrated event handling
- [**GCP GKE Cluster**](/infra-catalog/gcp-gke-cluster) -- GKE service destinations
- [**GCP Pub/Sub Topic**](/infra-catalog/gcp-pub-sub-topic) -- bring-your-own transport
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- the trigger's identity
- [**GCP KMS Key**](/infra-catalog/gcp-kms-key) -- channel CMEK
