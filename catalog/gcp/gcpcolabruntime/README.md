# GCP Colab Runtime

A Colab Enterprise runtime -- a notebook VM assigned to one user, built from a `GcpColabRuntimeTemplate`. Declaring runtimes gives a team ready, correctly configured machines (a GPU box for a new hire, a runtime per course participant) and lets the block start or stop them on purpose: `desiredState: STOPPED` keeps the disk and stops the compute bill.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project (never disabled on destroy)
- **Runtime** -- a `colab_runtime` assigned to `runtimeUser` from the template, started or stopped to match `desiredState`

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **`GcpColabRuntimeTemplate`** -- the template the runtime is assigned from (`runtimeTemplate`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpColabRuntime
metadata:
  name: alice-notebooks
spec:
  location: us-central1
  runtimeTemplate:
    valueFrom:
      kind: GcpColabRuntimeTemplate
      name: standard-cpu
      fieldPath: status.outputs.name
  runtimeUser: alice@example.com
```

```shell
planton apply -f colab-runtime.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Colab Enterprise region (the template's). Immutable. |
| `runtimeTemplate` | `StringValueOrRef` | A `GcpColabRuntimeTemplate` reference or literal template name. Immutable. |
| `runtimeUser` | `string` | The email of the user the runtime belongs to. Immutable. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `runtimeId` | `string` | `metadata.name` | Lowercase letters, digits, hyphens. Immutable. |
| `displayName` | `string` | `metadata.name` | Up to 128 characters. Immutable. |
| `description` | `string` | none | Immutable. |
| `desiredState` | `string` | running | `RUNNING` or `STOPPED`; enforced on every apply. |
| `autoUpgrade` | `bool` | `false` | Upgrade to the latest image on start when upgradable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `runtimeUser` is an email address; `desiredState` is `RUNNING` or `STOPPED`.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/notebookRuntimes/{runtime_id}` |
| `runtime_id` | `string` | The runtime's id |
| `location` | `string` | The runtime's region |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **`desiredState` is enforced on every apply.** A runtime the user started in the console is stopped again on the next apply if the manifest says `STOPPED` (and vice versa). Leave it unset to let users start and stop freely.
- **Only the runtime user can connect.** Assign one runtime per person.
- **A stopped runtime still bills its disk.** Destroy it when it is no longer needed.
- **The template's idle shutdown still applies** to a running runtime.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpColabRuntimeTemplate** -- the template the runtime is built from
- **GcpColabSchedule** -- scheduled notebook runs instead of an interactive runtime

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
