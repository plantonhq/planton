# GCP Deploy Custom Target Type

Declares a Cloud Deploy custom target type: how Cloud Deploy renders and deploys a release to a system it does not deploy to natively -- a vendor API, an internal platform, Terraform, a non-Kubernetes runtime. The work is either containers Cloud Deploy runs (`tasks`) or Skaffold custom actions (`customActions`), optionally pulled from shared Skaffold modules in Git, a Cloud Build repository, or Cloud Storage. Each `GcpDeployTarget` that deploys this way points at the type.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `clouddeploy.googleapis.com` on the type's project (never disabled on destroy)
- **Custom target type** -- one `clouddeploy_custom_target_type`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with `roles/clouddeploy.admin` (or the permissions in `iac/permissions.yaml`) on the type's project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Setup

- **The deployer** -- a container image (for `tasks`) or Skaffold custom actions (for `customActions`) that do the deploying. The execution service account of each target that uses the type must be able to pull the image and read any Skaffold module source.

### Optional Dependencies

- **`GcpCloudBuildRepository`** -- a repository holding shared Skaffold modules (`customActions.includeSkaffoldModules[].googleCloudBuildRepo.repository`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpDeployCustomTargetType
metadata:
  name: vendor-deployer
spec:
  projectId:
    value: acme-delivery
  location: us-central1
  description: Deploys releases through the vendor's API
  tasks:
    deploy:
      container:
        image: us-docker.pkg.dev/acme-delivery/deploy/vendor-deployer:1.4
        env:
          VENDOR_REGION: us
```

```shell
planton apply -f custom-target-type.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The region of the targets that use the type. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `string` / ref | provider project | The type's project (`GcpProject` ref). Immutable. |
| `customTargetTypeId` | `string` | `metadata.name` | The type's ID. Immutable. |
| `description` | `string` | -- | Up to 255 characters. |
| `labels` | `map` | -- | Labels on the type. |
| `annotations` | `map` | -- | User annotations; only declared keys are managed. |
| `tasks` | object | -- | `deploy` (required) and `render`, each with a `container`: `image` (required), `command`, `args`, `env`. |
| `customActions` | object | -- | `deployAction` (required), `renderAction`, `includeSkaffoldModules[]` (`configs` plus exactly one of `git`, `googleCloudBuildRepo`, `googleCloudStorage`). |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- At most one of `tasks` or `customActions`.
- Each Skaffold module names exactly one source; a literal Cloud Build repository is a full repository name; a Cloud Storage source starts with `gs://`.
- The type ID matches Google's ID rule.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/customTargetTypes/{custom_target_type_id}` -- what a target's `customTarget.customTargetType` takes |
| `custom_target_type_id` | `string` | The type's ID |
| `uid` | `string` | Google's unique identifier for the type |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Runs as the target.** The containers run in Cloud Build in each target's execution environment -- its service account and worker pool -- not as the type.
- **Same location.** Targets use a type in their own project and location.
- **Changes apply to the next rollout.** Updating the image or actions does not touch releases already rendered.
- **Destroy order.** Delete the targets that use a type before the type; a chart orders them by their reference.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpDeployTarget** -- targets that deploy through the type
- **GcpDeliveryPipeline** -- pipelines whose stages use those targets
- **GcpCloudBuildRepository** -- a source of shared Skaffold modules

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
