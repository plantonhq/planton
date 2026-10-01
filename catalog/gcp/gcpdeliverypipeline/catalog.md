# GCP Delivery Pipeline

Ships a release from dev to prod the same way every time. A Cloud Deploy delivery pipeline lists the targets a release is promoted through, decides how each rollout happens -- all at once with a verification step, or as a canary that moves traffic on Cloud Run or GKE in steps -- and can promote, advance, repair, and schedule releases on its own.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- the Cloud Deploy API on the pipeline's project
- **Delivery pipeline** -- the pipeline and its ordered stages
- **Automations** -- one per declared automation, each running as the service account you name

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with permission to manage Cloud Deploy pipelines and automations in the target project, and to act as each automation's service account. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP Delivery Pipeline**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Two-Stage Cloud Run Pipeline** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeliveryPipeline
metadata:
  name: web
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-apps
  location: us-central1
  serialPipeline:
    stages:
      - targetId:
          value: web-dev
      - targetId:
          value: web-prod
```

```shell
planton apply -f delivery-pipeline.yaml
```

This creates a pipeline that promotes releases from the web-dev target to web-prod. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference each GcpDeployTarget's `status.outputs.target_id` from `serialPipeline.stages[].targetId`, and a GcpServiceAccount's `status.outputs.email` from `automations[].serviceAccount`.

## Key Configuration

These are the most important decisions when configuring a delivery pipeline. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Stages** -- the targets in promotion order, each in the pipeline's project and region, with optional Skaffold profiles and deploy parameters.

**Strategy** -- per stage, a standard deploy (optionally verified, with jobs before and after) or a canary with fixed percentages or phase-by-phase control.

**Automations** -- promote a release once it succeeds on dev, advance a healthy canary, retry or roll back a failed rollout, or promote on a weekly schedule.

**Destroy** -- deleting the pipeline also deletes its releases, rollouts, and automations; protect it with deletion policy PREVENT.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpDeployTarget** | `serialPipeline.stages[].targetId`, automation selector and destination targets | `status.outputs.target_id` |
| **GcpServiceAccount** | `automations[].serviceAccount` | `status.outputs.email` |
| **GcpMonitoringAlertPolicy** | analysis `googleCloud.alertPolicyChecks[].alertPolicies` | `status.outputs.policy_name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The pipeline's full resource name | Tooling, audits |
| `delivery_pipeline_id` | The pipeline's ID | A deploy policy's pipeline selector; `gcloud deploy releases create --delivery-pipeline` |
| `uid` | The pipeline's unique identifier | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Two-stage Cloud Run pipeline** -- dev then prod, promoted automatically after dev succeeds. Start from the **Two-Stage Cloud Run Pipeline** preset.

**Cloud Run canary** -- prod rolled out at 25% and 50% before going stable, with canary advances and automatic rollback. Start from the **Cloud Run Canary with Repair** preset.

**GKE canary through a service mesh** -- phase-by-phase canary traffic through a Gateway API route. Start from the **GKE Gateway Canary** preset.

## Works With

- [**GCP Deploy Target**](/cloud-catalog/gcp-deploy-target) -- where each stage deploys
- [**GCP Deploy Policy**](/cloud-catalog/gcp-deploy-policy) -- rollout freezes and windows for the pipeline
- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- the identity automations run as
- [**GCP Cloud Build Trigger**](/cloud-catalog/gcp-cloud-build-trigger) -- builds that create releases
