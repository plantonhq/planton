# GcpSccNotificationConfig

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpSccNotificationConfigSpec streams Security Command Center findings to
a Pub/Sub topic as they are created or updated, at a project, folder, or
organization (`google_scc_v2_project_notification_config`,
`google_scc_v2_folder_notification_config`,
`google_scc_v2_organization_notification_config` -- the scope picks one).

The topic is where alerting, ticketing, and SOAR pipelines subscribe. The
filter narrows the stream (e.g. only active high-severity findings), so a
team routes what matters and leaves the rest in the console.

Two things must be true before findings arrive:
  - Security Command Center is activated on the scope (Standard is free;
    a project can be activated on its own when the organization is not).
  - The config's publisher can publish to the topic: a
    GcpPubSubTopicIamMember on the topic, role roles/pubsub.publisher,
    member referencing this config's service_account_member output.
    Creating the config does not check this; without the grant,
    notifications are silently dropped.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSccNotificationConfig
metadata:
  name: high-findings
spec:
  scope:
    projectId:
      value: my-gcp-project
  configId: high-findings
  pubsubTopic:
    value: projects/my-gcp-project/topics/scc-findings
  filter: state = "ACTIVE" AND severity = "HIGH"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.scope` | `GcpSccNotificationConfigScope` |  |  |  |
| `spec.scope.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.scope.folderId` | `string \| valueFrom` |  |  | GcpFolder (`status.outputs.folder_id`) |
| `spec.scope.organizationId` | `string` |  |  |  |
| `spec.configId` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.pubsubTopic` | `string \| valueFrom` |  |  | GcpPubSubTopic (`status.outputs.topic_id`) |
| `spec.filter` | `string` |  |  |  |
| `spec.location` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.scope

`GcpSccNotificationConfigScope`

Whose findings are streamed. Omit for the provider's default project.

- rule: set at most one of project_id, folder_id, or organization_id (empty means the provider's default project)

### spec.scope.projectId

`string | valueFrom`

A project: a literal project ID or a GcpProject reference.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.scope.folderId

`string | valueFrom`

A folder: the folder's numeric ID -- a literal or a GcpFolder
reference. Covers every project beneath it.

- references: GcpFolder (`status.outputs.folder_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFolder, name: <that resource's name>, fieldPath: status.outputs.folder_id}} -- a bare string does not parse

### spec.scope.organizationId

`string`

The organization: the numeric organization ID, without the
organizations/ prefix.

- rule: organization_id must be the numeric organization ID, without the organizations/ prefix

### spec.configId

`string` · required

The config's ID, unique within its parent: 1-128 letters, digits,
hyphens, or underscores. Immutable.

- rule: config_id must be 1-128 letters, digits, hyphens, or underscores
- rule: {"required":true}

### spec.description

`string`

What the config is for, up to 1024 characters.

- rule: {"string":{"maxLen":"1024"}}

### spec.pubsubTopic

`string | valueFrom`

The topic findings are published to: a literal
projects/{project}/topics/{topic} or a GcpPubSubTopic reference.
Required on folder and organization configs; Google lets a project
config omit it. Grant the config's publisher on it with a
GcpPubSubTopicIamMember (role roles/pubsub.publisher, member
referencing the service_account_member output).

- references: GcpPubSubTopic (`status.outputs.topic_id`)
- rule: pubsub_topic must be projects/{project}/topics/{topic}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpPubSubTopic, name: <that resource's name>, fieldPath: status.outputs.topic_id}} -- a bare string does not parse

### spec.filter

`string`

Which finding create and update events are streamed (Google's
streaming_config.filter). Restrictions of the form
`<field> <operator> <value>`, combined with AND and OR (OR binds
tighter), negated with a leading "-":
  state = "ACTIVE" AND severity = "HIGH"
  category = "OPEN_FIREWALL" AND state = "ACTIVE"
Operators: = for every type; >, <, >=, <= for integers; : for substring
match. Empty streams every finding.

### spec.location

`string`

Where the config is stored. "global", the default, unless Security
Command Center data residency was set up at activation (then the
residency location, e.g. "eu" or "us").

- rule: location must be global or a residency location such as eu or us

### spec.deletionPolicy

`string`

What destroying this block does:
  "" / "DELETE" -- the config is deleted and streaming stops
  "PREVENT"     -- destroy fails
  "ABANDON"     -- the config leaves management and keeps streaming

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `pubsub_topic_required_off_project_scope`: pubsub_topic is required on a folder or organization notification config

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpSccNotificationConfig, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.name` | `string` | Full resource name: {parent}/locations/{location}/notificationConfigs/{config_id}. |
| `status.outputs.service_account` | `string` | The Security Command Center service account that publishes the notifications, as a bare email. It needs roles/pubsub.publisher on the topic, or notifications are silently dropped; grant it through a GcpPubSubTopicIamMember whose member references service_account_member (this email in IAM member form). |
| `status.outputs.service_account_member` | `string` | The publisher in IAM member form, "serviceAccount:" + service_account -- the exact value an IAM member field takes. Reference it from a GcpPubSubTopicIamMember's member, with role roles/pubsub.publisher on the config's topic, so notifications are published rather than dropped. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.scope.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.scope.folderId` | GcpFolder | `status.outputs.folder_id` |
| `spec.pubsubTopic` | GcpPubSubTopic | `status.outputs.topic_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpPubSubTopicIamMember | `spec.member` | `status.outputs.service_account_member` |

## See Also

- [Overview](../README.md)
