# GCP Delivery Pipeline

Declares a Cloud Deploy delivery pipeline: the ordered stages a release is promoted through (dev, then staging, then prod), how each rollout is carried out (a standard deploy with optional verification and jobs, or a canary on Cloud Run or GKE), and the pipeline's automations -- promotions, canary advances, repairs, and scheduled promotions that Cloud Deploy performs on its own as a service account you name.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- `clouddeploy.googleapis.com` on the pipeline's project (never disabled on destroy)
- **Delivery pipeline** -- one `clouddeploy_delivery_pipeline` with its serial stages
- **Automations** -- one `clouddeploy_automation` per `automations` entry

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/clouddeploy.admin` (or the permissions in `iac/permissions.yaml`) on the pipeline's project, plus `roles/iam.serviceAccountUser` on each automation's service account.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpDeployTarget`** -- each stage's target (`serialPipeline.stages[].targetId`), in the pipeline's project and region; also the targets an automation selects or promotes to.
- **`GcpServiceAccount`** -- the identity each automation runs as (`automations[].serviceAccount`).
- **`GcpMonitoringAlertPolicy`** -- alert policies an analysis job watches.

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeliveryPipeline
metadata:
  name: web
spec:
  location: us-central1
  serialPipeline:
    stages:
      - targetId:
          valueFrom:
            kind: GcpDeployTarget
            name: web-dev
            fieldPath: status.outputs.target_id
      - targetId:
          valueFrom:
            kind: GcpDeployTarget
            name: web-prod
            fieldPath: status.outputs.target_id
        strategy:
          canary:
            canaryDeployment:
              percentages: [25, 50]
            runtimeConfig:
              cloudRun:
                automaticTrafficControl: true
  automations:
    - automationId: promote-to-prod
      serviceAccount:
        valueFrom:
          kind: GcpServiceAccount
          name: web-deployer
          fieldPath: status.outputs.email
      selector:
        targets:
          - id:
              value: web-dev
      rules:
        - promoteReleaseRule:
            id: promote
            destinationTargetId:
              value: "@next"
```

```shell
planton apply -f delivery-pipeline.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The region the pipeline (and its targets) live in. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The pipeline's project (`GcpProject` ref). Immutable. |
| `deliveryPipelineId` | `string` | `metadata.name` | The pipeline's ID. Immutable. |
| `description` | `string` | -- | Up to 255 characters. |
| `labels` / `annotations` | `map` | -- | Labels (attribution labels added) and AIP-128 annotations. |
| `suspended` | `bool` | `false` | Block new releases and rollouts. |
| `serialPipeline.stages[]` | list | -- | `targetId` (bare ID or `GcpDeployTarget` ref), `profiles`, `deployParameters[]` (`values`, `matchTargetLabels`), `strategy`. |
| `stages[].strategy.standard` | object | -- | `verify`, `predeploy` / `postdeploy` (`actions` or `tasks`), `verifyConfig.tasks`, `analysis`. |
| `stages[].strategy.canary` | object | -- | `canaryDeployment` (`percentages`, `verify`, jobs, `verifyConfig`, `analysis`) or `customCanaryDeployment.phaseConfigs[]`; `runtimeConfig` with `cloudRun` or `kubernetes` (`gatewayServiceMesh` or `serviceNetworking`). |
| `automations[]` | list | -- | `automationId`, `serviceAccount` (required), `selector.targets[]` (required), `rules[]` (required), `description`, `labels`, `annotations`, `suspended`. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`, fanned to every automation. |

### Validation Rules

- `deliveryPipelineId` and every rule and phase ID are 1-63 lowercase letters, digits, or hyphens, starting with a letter.
- Stage, selector, and destination targets are bare target IDs (selectors also take `*`, destinations `@next`).
- At most one of `canaryDeployment` or `customCanaryDeployment`, of `cloudRun` or `kubernetes`, of `gatewayServiceMesh` or `serviceNetworking`, and of `actions` or `tasks` on a standard job.
- A `canaryDeployment` on Cloud Run requires `automaticTrafficControl: true`.
- Each automation rule sets exactly one rule kind; a repair rule needs at least one repair phase, each exactly one of `retry` or `rollback`.
- Automation IDs, phase IDs, and rule IDs are unique where Google requires it.

## Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/deliveryPipelines/{delivery_pipeline_id}` |
| `delivery_pipeline_id` | `string` | The pipeline's ID |
| `uid` | `string` | Google's unique identifier for the pipeline |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Destroy is forceful.** The provider deletes the pipeline with `force=true`, which also deletes every release, rollout, and automation under it. Set `deletionPolicy: PREVENT` on a pipeline whose history matters.
- **Same project and region.** Stage targets are looked up by bare ID in the pipeline's project and region.
- **Canary jobs take actions only.** The pinned provider declares no `tasks` on the canary paths' predeploy and postdeploy; use Skaffold custom actions there.
- **Automations need two grants.** The deploying identity needs `iam.serviceAccounts.actAs` on the automation's service account, and that account needs permission to create releases and rollouts (`roles/clouddeploy.operator` or narrower).

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Kinds

- **GcpDeployTarget** -- the targets the stages deploy to
- **GcpDeployPolicy** -- restricts rollouts on pipelines it selects (by `delivery_pipeline_id` or labels)
- **GcpServiceAccount** -- the identities automations run as
- **GcpCloudBuildTrigger** -- builds that create releases on the pipeline

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
