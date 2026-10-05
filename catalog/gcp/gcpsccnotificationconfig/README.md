# GCP SCC Notification Config

Streams Security Command Center findings to a Pub/Sub topic as they are created or updated, for a project, a folder, or the whole organization. The topic is where alerting, ticketing, and SOAR pipelines subscribe; the filter narrows the stream to what matters, such as active high-severity findings.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `securitycenter.googleapis.com` on a project config's project (never disabled on destroy)
- **Notification config** -- one `scc_v2_{project,folder,organization}_notification_config`, chosen by the scope

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Security Command Center notification-config admin permissions (`roles/securitycenter.notificationConfigEditor`) at the scope.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Security Command Center

Security Command Center must be activated on the scope -- the organization, or the project on its own (Standard is free). Without activation, creation fails.

### Optional Dependencies

- **`GcpPubSubTopic`** -- the destination topic (`pubsubTopic`). Grant the config's publisher with a `GcpPubSubTopicIamMember` on the topic (role `roles/pubsub.publisher`, `member` referencing the config's `status.outputs.service_account_member`), or notifications are dropped.
- **`GcpProject`** / **`GcpFolder`** -- the scope, by reference.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSccNotificationConfig
metadata:
  name: high-findings
spec:
  configId: high-findings
  pubsubTopic:
    value: projects/my-gcp-project/topics/scc-findings
  filter: state = "ACTIVE" AND severity = "HIGH"
```

```shell
planton apply -f scc-notification-config.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `configId` | `string` | The config's ID: 1-128 letters, digits, hyphens, underscores. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `scope` | `object` | provider project | At most one of `projectId` (`GcpProject` ref), `folderId` (`GcpFolder` ref), `organizationId` (numeric). |
| `pubsubTopic` | `string` / ref | none | `projects/{project}/topics/{topic}` or a `GcpPubSubTopic` ref. Required on folder and organization configs. |
| `filter` | `string` | every finding | Which create and update events are streamed, e.g. `state = "ACTIVE" AND severity = "HIGH"`. |
| `description` | `string` | none | Up to 1024 characters. |
| `location` | `string` | `global` | Where the configuration is stored; a residency location (`eu`, `us`) only if data residency was set up at activation. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE` removes it, `PREVENT` fails destroy, `ABANDON` keeps it in Google. |

### Validation Rules

- `scope` names at most one of project, folder, organization; `organizationId` is numeric without the `organizations/` prefix.
- `location` is `global` or a residency location; `deletionPolicy` takes only Google's values.
- `pubsubTopic` is required off the project scope and has the form `projects/{project}/topics/{topic}`.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `{parent}/locations/{location}/notificationConfigs/{configId}` |
| `service_account` | `string` | The publisher Security Command Center uses, as a bare email |
| `service_account_member` | `string` | The publisher in IAM member form (`serviceAccount:{email}`) -- the `member` a `GcpPubSubTopicIamMember` grants `roles/pubsub.publisher` on the topic |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Grant the publisher.** Creating the config does not check that `service_account` can publish; without `roles/pubsub.publisher` on the topic, notifications are silently dropped. Grant it with a `GcpPubSubTopicIamMember` on the topic (role `roles/pubsub.publisher`, `member` referencing the config's `status.outputs.service_account_member`).
- **Activation first.** Security Command Center must be active on the scope.
- **The filter decides the noise.** An empty filter streams every finding; narrow it to the states and severities your responders act on.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpPubSubTopic** -- the destination topic
- **GcpSccMuteConfig** -- mute noise before it is streamed
- **GcpSccBigQueryExport** -- keep the findings history in BigQuery

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
