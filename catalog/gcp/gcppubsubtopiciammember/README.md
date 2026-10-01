# GCP Pub/Sub Topic IAM Member

Deploys a single ADDITIVE IAM grant ON a Pub/Sub topic (`google_pubsub_topic_iam_member`) — one role, to one member, on one topic. The grant merges into the topic's IAM policy without touching any other member's bindings, and removal subtracts only this exact pair.

The canonical use is letting a Google-managed identity publish into a topic it delivers to: a `GcpLoggingSink`'s writer identity needs `roles/pubsub.publisher` on its destination topic, and so does a `GcpSccNotificationConfig`'s publishing service account. Those resources name the topic themselves, so the grant cannot live on the topic without a dependency cycle. This kind depends on both sides, so the order is always topic, then sink (or notification config), then grant.

## What Gets Created

When you deploy a GcpPubSubTopicIamMember resource, Planton provisions:

- **Topic IAM Member** — one (role, member) entry merged into the target topic's IAM policy

Nothing else in the policy is read as owned or modified — grants made by other charts, teams, or tools are never clobbered.

## Prerequisites

- **GCP credentials** configured via environment variables or Planton provider config
- **An existing topic** — referenced via `topic` (a GcpPubSubTopic's `topic_id` output)
- **IAM permissions** — see [`iac/permissions.yaml`](iac/permissions.yaml) for the least-privilege permission set the deploying principal needs
- **The member must exist** — a service agent, a sink's writer identity, or a GcpServiceAccount

## Quick Start

Create a file `sink-publisher.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPubSubTopicIamMember
metadata:
  name: audit-sink-publisher
spec:
  topic:
    value: projects/my-gcp-project-123/topics/audit-events
  role:
    value: roles/pubsub.publisher
  member:
    value: serviceAccount:service-123456789@gcp-sa-logging.iam.gserviceaccount.com
```

Deploy:

```shell
planton apply -f sink-publisher.yaml
```

Or compose both sides by reference — the shape this kind exists for:

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

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `topic` | `StringValueOrRef` | The topic whose IAM policy receives the grant, as its full resource name (`projects/<project>/topics/<topic>`). References a GcpPubSubTopic's `topic_id` output by default. The project is read from the name. Immutable. |
| `role` | `StringValueOrRef` | The role to grant on the topic: a predefined role (`roles/pubsub.publisher`, `roles/pubsub.subscriber`, `roles/pubsub.viewer`, ...) or a custom role's full name. References a GcpIamCustomRole's `name` output by default. Immutable. |
| `member` | `StringValueOrRef` | The identity receiving the grant, in IAM member format. References a GcpServiceAccount's `member` output by default; a GcpLoggingSink's `writer_identity` and a GcpSccNotificationConfig's `service_account_member` are the other candidates. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|

There is no project field: the project is embedded in the topic's full name, and a topic in another project is granted the same way.

### Member Formats

| Format | Grants to |
|--------|-----------|
| `serviceAccount:<email>` | A service account, a sink's writer identity, or a Google service agent |
| `user:<email>` | A Google account |
| `group:<email>` | A Google group |
| `domain:<domain>` | Everyone in a Workspace or Cloud Identity domain |
| `principal://...` / `principalSet://...` | Workload identity federation principals |
| `allUsers` / `allAuthenticatedUsers` | Public grants — never appropriate on a topic |

Grants to deleted principals (`deleted:...`) are refused by validation.

## Topic Grant vs Project Grant

A project-level `roles/pubsub.publisher` lets the member publish to every topic in the project. A topic-scoped grant — this component — lets it publish to exactly one. A sink's writer identity needs nothing more than its own destination topic, so the topic grant is the least-privilege shape and the one that keeps the dependency graph honest.

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `topic` | `string` | The topic whose policy received the grant (after reference resolution) |
| `role` | `string` | The granted role (after reference resolution) |
| `member` | `string` | The granted member (after reference resolution) |
| `etag` | `string` | The topic IAM policy etag after the grant — useful for audit correlation |

## Deployment Methods

Planton supports two deployment methods:

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Everything is immutable**: IAM grants have no update — changing topic, role, or member replaces the grant (destroy the old pair, create the new one), which mirrors the underlying API exactly.
- **Only additive grants are modeled**: authoritative per-role bindings and whole-policy writes clobber every grant they do not list and are deliberately not modeled.
- **A sink exports nothing until this grant lands**: Cloud Logging drops entries it cannot publish. Deploy the grant in the same chart as the sink so the export works from its first entry.
- **Recreating a sink can mint a new writer identity**: a reference to `writer_identity` follows it automatically; a literal member does not.

## Related Components

- [GcpPubSubTopic](/docs/catalog/gcp/gcppubsubtopic) — the topic being granted on (its `topic_id` output feeds this component)
- [GcpLoggingSink](/docs/catalog/gcp/gcploggingsink) — a sink exporting to the topic (its `writer_identity` output feeds `member`)
- [GcpSccNotificationConfig](/docs/catalog/gcp/gcpsccnotificationconfig) — Security Command Center notifications published to the topic
- [GcpServiceAccount](/docs/catalog/gcp/gcpserviceaccount) — a grantable identity (its `member` output feeds this component)
- [GcpGcsBucketIamMember](/docs/catalog/gcp/gcpgcsbucketiammember) — the same pattern for a sink exporting to a bucket

## Additional Resources

- [Pub/Sub access control](https://cloud.google.com/pubsub/docs/access-control)
- [Route logs to supported destinations](https://cloud.google.com/logging/docs/export/configure_export_v2)
- [Services that allow conditional role bindings](https://cloud.google.com/iam/docs/resource-types-with-conditional-roles)

## Support

For issues, questions, or contributions, please refer to the Planton documentation or open an issue in the repository.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
