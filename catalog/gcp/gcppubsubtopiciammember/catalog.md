# GCP Pub/Sub Topic IAM Member

Grants one role, to one identity, on ONE Pub/Sub topic — the least-privilege way to let a logging sink, a Security Command Center notification config, or a workload publish into a topic. Additive: it merges into the topic's IAM policy without touching any other member's bindings, and removal subtracts only this exact (role, member) pair. It exists as its own resource because the identities that most need publish rights belong to resources that name the topic themselves; a grant declared on the topic would be a dependency cycle, while this one simply lands last.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Topic IAM Member Binding** -- a `google_pubsub_topic_iam_member` merging the (role, member) pair into the target topic's IAM policy

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials permitted to set IAM policy on the target topic (e.g. `roles/pubsub.admin` on the topic or its project). Map it as the default for your environment, or specify it explicitly.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A Pub/Sub topic** whose IAM policy receives the grant. Provide its full name (`projects/<project>/topics/<topic>`) directly or reference a GcpPubSubTopic Infra Component via ValueFromRef.
- **The identity** receiving the grant must already exist — a sink's writer identity exists once the sink is created.

## Deploy

### Console

Open the deployment store, find **GCP Pub/Sub Topic IAM Member**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the grant definition. Start from the **Logging Sink Publisher** preset in the [Presets](#presets) tab for the most common shape.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPubSubTopicIamMember
metadata:
  name: audit-sink-publisher
  org: acme-corp
  env: prod
spec:
  topic:
    value: projects/acme-prod-12345/topics/audit-events
  role:
    value: roles/pubsub.publisher
  member:
    value: serviceAccount:service-123456789@gcp-sa-logging.iam.gserviceaccount.com
```

```shell
planton apply -f gcp-pubsub-topic-iam-member.yaml
```

This merges one binding into the topic's policy. An Infra Job tracks the provisioning in real time.

### InfraChart

The composed form is where this kind shines — topic, sink, and grant wired in one InfraPipeline without a cycle:

```yaml
spec:
  topic:
    valueFrom:
      kind: GcpPubSubTopic
      name: audit-events
      fieldPath: status.outputs.topic_id
  role:
    value: roles/pubsub.publisher
  member:
    valueFrom:
      kind: GcpLoggingSink
      name: audit-export
      fieldPath: status.outputs.writer_identity
```

The InfraPipeline deploys the topic, then the sink that names it, then this grant with both values resolved.

## Key Configuration

These are the most important decisions when configuring a grant. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**The sink pattern** -- a logging sink exporting to Pub/Sub publishes as its writer identity, and Cloud Logging drops entries it is not allowed to publish. Grant that identity `roles/pubsub.publisher` on exactly the destination topic by referencing the sink's `writer_identity` output. A Security Command Center notification config works the same way through its `service_account_member` output.

**Role** -- `roles/pubsub.publisher` to publish, `roles/pubsub.subscriber` to attach subscriptions (including from another project), `roles/pubsub.viewer` to read configuration, or a custom role's full name.

**Member format** -- the prefix declares the identity type: `serviceAccount:<email>`, `user:<email>`, `group:<email>`, `domain:<domain>`, or `principal://`/`principalSet://` federation principals. Format validation runs on literal values; references arrive already in member format.

**No conditions on topics** -- Pub/Sub is not among the services Google lists as accepting conditional role bindings, so this grant has no `condition`. Scope access by choosing the topic and the role instead.

**Everything replaces atomically** -- an IAM grant has no update. Changing topic, role, or member replaces the grant, and for the moment between delete and create the member cannot publish.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpPubSubTopic** | `topic` | `status.outputs.topic_id` |
| **GcpIamCustomRole** (optional) | `role` | `status.outputs.name` |
| **GcpServiceAccount** (optional) | `member` | `status.outputs.member` |
| **GcpLoggingSink** (optional) | `member` | `status.outputs.writer_identity` |
| **GcpSccNotificationConfig** (optional) | `member` | `status.outputs.service_account_member` |

### What This Kind Provides

This kind has no outputs a downstream Infra Component would consume: `status.outputs` records the grant's post-resolution facts — the (`topic`, `role`, `member`) triple after any references were resolved, plus the topic IAM policy `etag` at the moment this grant merged. They exist for audit and drift review, not for ValueFromRef wiring.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Logging sink publisher** -- a sink's writer identity gets publisher on its destination topic; the grant every Pub/Sub log export needs. Start from the **Logging Sink Publisher** preset.

**Security Command Center publisher** -- the notification config's service account gets publisher on the findings topic. Start from the **Security Command Center Publisher** preset.

**Workload subscriber** -- an application's service account may attach subscriptions to a topic, for consumers in another project. Start from the **Workload Subscriber** preset.

## Works With

- [**GCP Pub/Sub Topic**](/infra-catalog/gcp-pub-sub-topic) -- its `topic_id` output feeds the topic field; the topic this grant controls access to
- [**GCP Logging Sink**](/infra-catalog/gcp-logging-sink) -- its `writer_identity` output feeds the member field for Pub/Sub log exports
- [**GCP SCC Notification Config**](/infra-catalog/gcp-scc-notification-config) -- its `service_account_member` output feeds the member field for findings notifications
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- its `member` output feeds the member field for workload access
- [**GCP IAM Custom Role**](/infra-catalog/gcp-iam-custom-role) -- its `name` output feeds the role field for curated permission bundles
