# GcpPubSubTopicIamMember

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpPubSubTopicIamMemberSpec defines a single ADDITIVE IAM grant ON a Pub/Sub
topic: one role, to one member, on one topic.

The canonical use is letting a Google-managed identity publish to a topic
it delivers into:
  roles/pubsub.publisher -- what a GcpLoggingSink's writer identity needs
    on its destination topic, what a GcpSccNotificationConfig's service
    account needs on its notification topic, and what any workload's
    service account needs to publish.
  roles/pubsub.subscriber -- attach subscriptions to the topic (pull or
    push consumers in another project).
  roles/pubsub.viewer -- read the topic's configuration.

Why a standalone grant and not a field on GcpPubSubTopic: the identities
that most need publish rights belong to resources that name the topic
themselves (a sink's destination, a notification config's topic). A grant
declared on the topic that referenced those identities would make the
topic depend on the sink while the sink depends on the topic -- a cycle no
chart can deploy. This grant depends on both, so the order is always
topic, then sink, then grant.

Additive semantics: the grant merges into the topic's IAM policy without
touching any other member's bindings on the same role, and removal
subtracts only this exact (role, member) entry. Grants from
other charts, teams, or tools never fight over the policy. (Authoritative
binding and policy management, which clobbers everything not listed, is
deliberately not modeled.)

Every field is immutable, mirroring the API: an IAM grant has no update --
changing the topic, role, or member replaces the grant.

No IAM Condition: Pub/Sub topics do not accept conditional role bindings
(the provider reads and writes a topic's policy without requesting policy
version 3, which conditions require, and documents no condition support
for topic IAM). Scope access by the topic and the role; a time-boxed grant
belongs on the project.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpPubSubTopicIamMember
metadata:
  name: my-sample-topic-grant
spec:
  # The topic whose IAM policy receives this grant, as its full resource
  # name. Reference a GcpPubSubTopic -- its `topic_id` output is exactly
  # this value.
  topic:
    value: projects/my-gcp-project-123/topics/audit-events

  # The role to grant on the topic: what a sink's writer identity needs.
  role:
    value: roles/pubsub.publisher

  # The identity receiving the grant, in GCP IAM member format -- here a
  # project's Cloud Logging writer identity. Reference a GcpLoggingSink's
  # `writer_identity` output to follow the sink's own identity.
  member:
    value: serviceAccount:service-123456789@gcp-sa-logging.iam.gserviceaccount.com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.topic` | `string \| valueFrom` | yes |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.role` | `string \| valueFrom` | yes |  | GcpIamCustomRole (`status.outputs.name`) |
| `spec.member` | `string \| valueFrom` | yes |  | GcpServiceAccount (`status.outputs.member`), GcpLoggingSink (`status.outputs.writer_identity`), GcpSccNotificationConfig (`status.outputs.service_account_member`) |

## Field Details

### spec.topic

`string | valueFrom` · required

The topic whose IAM policy receives this grant, by its full resource
name: projects/<project>/topics/<topic>. Reference a GcpPubSubTopic --
its `topic_id` output is exactly this value. The project is read from
the name, so there is no separate project field, and a topic in another
project is granted the same way.

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: a literal topic must be its full resource name: projects/{project}/topics/{topic}
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.role

`string | valueFrom` · required

The role to grant on the topic. A predefined role
("roles/pubsub.publisher", "roles/pubsub.subscriber",
"roles/pubsub.viewer", "roles/pubsub.editor", "roles/pubsub.admin") or a
custom role's full name ("projects/<project>/roles/<role_id>" or
"organizations/<org>/roles/<role_id>"). Reference a GcpIamCustomRole to
grant a custom role -- its `name` output is exactly this value.

- references: GcpIamCustomRole (`status.outputs.name`)
- rule: a literal role must be roles/<name>, projects/<project>/roles/<id>, or organizations/<org>/roles/<id>
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpIamCustomRole, name: <that resource's name>, fieldPath: status.outputs.name}} -- a bare string does not parse

### spec.member

`string | valueFrom` · required

The identity receiving the grant, in IAM member format:
  serviceAccount:<email>  -- a service account or Google service agent.
                             Reference GcpServiceAccount (`member`), a
                             GcpLoggingSink (`writer_identity`, the
                             sink's writer, already in this format), or
                             a GcpSccNotificationConfig
                             (`service_account_member`, the identity
                             that publishes its notifications).
  user:<email>, group:<email>, domain:<domain>
  principal://... / principalSet://... -- workload identity federation
  allUsers / allAuthenticatedUsers     -- public grants; avoid on topics
Grants to deleted principals ("deleted:...") are not supported.

- references: GcpServiceAccount (`status.outputs.member`), GcpLoggingSink (`status.outputs.writer_identity`), GcpSccNotificationConfig (`status.outputs.service_account_member`)
- rule: a literal member must be <type>:<value> (serviceAccount:, user:, group:, domain:, principal://, principalSet://) or allUsers / allAuthenticatedUsers, and never a deleted: principal
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.member}} -- a bare string does not parse

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpPubSubTopicIamMember, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.topic` | `string` | The topic whose IAM policy received the grant, as configured (projects/<project>/topics/<topic>), after reference resolution. |
| `status.outputs.role` | `string` | The role that was granted, after reference resolution. |
| `status.outputs.member` | `string` | The member the role was granted to, in IAM member format, after reference resolution. |
| `status.outputs.etag` | `string` | The etag of the topic's IAM policy after this grant was applied -- a fingerprint of the policy version, useful for audit correlation. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.topic` | GcpPubSubTopic | `status.outputs.topic_id` |
| `spec.role` | GcpIamCustomRole | `status.outputs.name` |
| `spec.member` | GcpServiceAccount | `status.outputs.member` |
| `spec.member` | GcpLoggingSink | `status.outputs.writer_identity` |
| `spec.member` | GcpSccNotificationConfig | `status.outputs.service_account_member` |

## See Also

- [Overview](../README.md)
