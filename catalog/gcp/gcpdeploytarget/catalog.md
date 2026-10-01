# GCP Deploy Target

Names a place your releases go. A Cloud Deploy target is one environment a delivery pipeline promotes through -- a Cloud Run region, a GKE cluster, a fleet cluster, a group of targets deployed together, or a custom platform -- with an optional approval gate and the identity, bucket, and build machines its deploy jobs use.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- the Cloud Deploy API on the target's project
- **Target** -- the deployment target

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with permission to manage Cloud Deploy targets in the target's project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP Deploy Target**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Cloud Run Target** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeployTarget
metadata:
  name: staging
  org: acme-corp
  env: staging
spec:
  projectId:
    value: acme-delivery
  location: us-central1
  run:
    location: projects/acme-staging/locations/us-central1
```

```shell
planton apply -f deploy-target.yaml
```

This creates a target that deploys Cloud Run services into the staging project's us-central1 region. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a project's `status.outputs.project_id` from `projectId`; a cluster's `status.outputs.cluster_id` from `gke.cluster`; other targets' `status.outputs.target_id` from `multiTarget.targetIds`; a custom target type's `status.outputs.name` from `customTarget.customTargetType`; and a service account's `status.outputs.email`, a bucket's `status.outputs.url`, and a worker pool's `status.outputs.name` from `executionConfigs`.

## Key Configuration

These are the most important decisions when configuring a target. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Target type** -- exactly one: Cloud Run, a GKE cluster, a fleet membership, a multi-target, or a custom target type.

**Approval** -- `requireApproval` makes every rollout to the target wait for a human, the usual gate in front of production.

**Execution** -- which service account deploys, where rendered manifests and logs go, and whether the jobs run on a private worker pool (needed to reach a private GKE control plane).

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpGkeCluster** | `gke.cluster`, `associatedEntities[].gkeClusters[].cluster` | `status.outputs.cluster_id` |
| **GcpGkeFleetMembership** | `anthosCluster.membership` | `status.outputs.name` |
| **GcpGkeCluster** | `anthosCluster.membership` | `status.outputs.fleet_membership` |
| **GcpDeployTarget** | `multiTarget.targetIds` | `status.outputs.target_id` |
| **GcpDeployCustomTargetType** | `customTarget.customTargetType` | `status.outputs.name` |
| **GcpServiceAccount** | `executionConfigs[].serviceAccount` | `status.outputs.email` |
| **GcpGcsBucket** | `executionConfigs[].artifactStorage` | `status.outputs.url` |
| **GcpCloudBuildWorkerPool** | `executionConfigs[].workerPool` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The target's full resource name | Audits, tooling |
| `target_id` | The target's ID | Delivery pipeline stages, multi-targets, deploy policies |
| `uid` | The target's unique identifier | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Cloud Run environment** -- a region in a project. Start from the **Cloud Run Target** preset.

**Production behind an approval** -- a gated GKE target with its own deployer identity. Start from the **Approved GKE Target** preset.

**Multi-region rollout** -- one target that deploys to several at once. Start from the **Multi-Region Target** preset.

## Works With

- [**GCP Delivery Pipeline**](/cloud-catalog/gcp-delivery-pipeline) -- the pipeline that deploys to the target
- [**GCP Deploy Custom Target Type**](/cloud-catalog/gcp-deploy-custom-target-type) -- the type behind a custom target
- [**GCP Deploy Policy**](/cloud-catalog/gcp-deploy-policy) -- rollout restrictions on targets
- [**GCP Cloud Build Worker Pool**](/cloud-catalog/gcp-cloud-build-worker-pool) -- private machines for deploy jobs
- [**GCP GKE Cluster**](/cloud-catalog/gcp-gke-cluster) -- clusters to deploy to
