# GCP Deploy Custom Target Type

Lets Cloud Deploy ship to anything. Out of the box Cloud Deploy deploys to GKE, Cloud Run, and fleet clusters; a custom target type teaches it to deploy somewhere else -- a vendor's API, an internal platform, Terraform -- with a container you provide or with Skaffold custom actions. Your pipelines keep their promotions, approvals, and rollbacks; only the last step changes.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **API enablement** -- the Cloud Deploy API on the type's project
- **Custom target type** -- the render and deploy definition targets point at

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with permission to manage Cloud Deploy custom target types in the target project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

## Deploy

### Console

Open the deployment store, find **GCP Deploy Custom Target Type**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Container Deployer** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeployCustomTargetType
metadata:
  name: vendor-deployer
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-delivery
  location: us-central1
  tasks:
    deploy:
      container:
        image: us-docker.pkg.dev/acme-delivery/deploy/vendor-deployer:1.4
```

```shell
planton apply -f custom-target-type.yaml
```

This defines a target type whose deploy step runs your deployer image. An Infra Job tracks the provisioning in real time.

### InfraChart

Reference a Cloud Build repository's `status.outputs.name` from `customActions.includeSkaffoldModules[].googleCloudBuildRepo.repository`, then reference this type's `status.outputs.name` from each `GcpDeployTarget`'s `customTarget.customTargetType`.

## Key Configuration

These are the most important decisions when configuring a custom target type. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Tasks or custom actions** -- tasks run a container you name, with no Skaffold knowledge needed; custom actions name Skaffold custom actions, which suits teams already standardized on Skaffold.

**Shared modules** -- custom actions can come from a shared Git repository, Cloud Build repository, or Cloud Storage path, so every application does not copy them.

**Render** -- leave it unset to let Cloud Deploy render normally, or supply your own render container or action.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpCloudBuildRepository** | `customActions.includeSkaffoldModules[].googleCloudBuildRepo.repository` | `status.outputs.name` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The type's full resource name | A deploy target's `customTarget.customTargetType` |
| `custom_target_type_id` | The type's ID | Tooling |
| `uid` | The type's unique identifier | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Container deployer** -- a deploy container, no Skaffold. Start from the **Container Deployer** preset.

**Shared Skaffold actions** -- custom actions from a shared Git repository. Start from the **Shared Skaffold Actions** preset.

## Works With

- [**GCP Deploy Target**](/infra-catalog/gcp-deploy-target) -- targets that deploy through the type
- [**GCP Delivery Pipeline**](/infra-catalog/gcp-delivery-pipeline) -- pipelines that promote across those targets
- [**GCP Cloud Build Repository**](/infra-catalog/gcp-cloud-build-repository) -- shared Skaffold modules
