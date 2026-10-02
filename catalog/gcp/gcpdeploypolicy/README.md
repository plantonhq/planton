# GCP Deploy Policy

Declares a Cloud Deploy deploy policy: rollout restrictions -- freeze windows -- that block chosen actions on chosen delivery pipelines and targets at chosen times. "No production deploys on weekends" and "nothing ships over the year-end freeze" are each one policy. Selectors pick what it governs by ID, `*`, or labels; rules say which actions are blocked, for whom, and when.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `clouddeploy.googleapis.com` on the policy's project (never disabled on destroy)
- **Deploy policy** -- one `clouddeploy_deploy_policy`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/clouddeploy.admin` (or the permissions in `iac/permissions.yaml`) on the policy's project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpDeliveryPipeline`** / **`GcpDeployTarget`** -- what the selectors name by ID. A policy can also select by `*` or by labels, and may exist before the pipelines and targets it will govern.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeployPolicy
metadata:
  name: prod-weekend-freeze
spec:
  projectId:
    value: acme-delivery
  location: us-central1
  description: No production rollouts on weekends
  rules:
    - rolloutRestriction:
        id: weekend-freeze
        actions:
          - CREATE
        timeWindows:
          timeZone: America/New_York
          weeklyWindows:
            - daysOfWeek:
                - SATURDAY
                - SUNDAY
  selectors:
    - target:
        id:
          valueFrom:
            kind: GcpDeployTarget
            name: web-prod
```

```shell
planton apply -f deploy-policy.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The region of the pipelines and targets the policy governs. Immutable. |
| `rules[]` | list | At least one. `rolloutRestriction`: `id` (required), `actions`, `invokers`, `timeWindows` (required: `timeZone`, `oneTimeWindows`, `weeklyWindows`). |
| `selectors[]` | list | At least one. `deliveryPipeline` and/or `target`, each with `id` (reference, literal ID, or `*`) and `labels`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The policy's project (`GcpProject` ref). Immutable. |
| `deployPolicyId` | `string` | `metadata.name` | The policy's ID. Immutable. |
| `description` | `string` | -- | Up to 255 characters. |
| `labels` | `map` | -- | Labels on the policy itself. |
| `annotations` | `map` | -- | User annotations; only declared keys are managed. |
| `suspended` | `bool` | `false` | Keeps the policy without enforcing it. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

`actions` take `ADVANCE`, `APPROVE`, `CANCEL`, `CREATE`, `IGNORE_JOB`, `RETRY_JOB`, `ROLLBACK`, `TERMINATE_JOBRUN` (empty blocks all); `invokers` take `USER` and `DEPLOY_AUTOMATION` (empty blocks both). A one-time window needs `startDate`, `startTime`, `endDate`, and `endTime`; a weekly window takes `daysOfWeek` (empty means every day) and optionally `startTime` with `endTime` (neither blocks the whole day).

### Validation Rules

- Restriction IDs are unique in the policy and match Google's ID rule.
- Every restriction has time windows with a time zone.
- A weekly window sets both `startTime` and `endTime`, or neither.
- A literal selector ID is a pipeline or target ID, or `*` -- never a full resource name.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/deployPolicies/{deploy_policy_id}` |
| `deploy_policy_id` | `string` | The policy's ID, which an override names |
| `uid` | `string` | Google's unique identifier for the policy |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Same location only.** A policy governs pipelines and targets in its own project and location.
- **Any selector matches.** The policy applies when any selector matches; within a selector, every attribute given must match.
- **Overrides are the emergency door.** A blocked action goes through only when its caller names the policy as an override, which needs `clouddeploy.deployPolicies.override`.
- **Lift a freeze early** with `suspended: true` instead of deleting the policy.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpDeliveryPipeline** -- pipelines the policy governs
- **GcpDeployTarget** -- targets the policy governs

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
