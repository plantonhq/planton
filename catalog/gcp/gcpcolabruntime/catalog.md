# GCP Colab Runtime

Hands a person a ready notebook machine. A Colab Enterprise runtime is a notebook VM assigned to one user, built from one of your runtime templates -- so a new hire or a workshop participant opens Colab Enterprise and finds their runtime waiting, already on the right machine, network, and image. The block can also keep a runtime stopped until it is needed, so you pay only for its disk.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project
- **Runtime** -- a `colab.Runtime` assigned to the user from the template

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Required Dependencies

- **GcpColabRuntimeTemplate** -- the template the runtime is built from.

## Deploy

### Console

Open the deployment store, find **GCP Colab Runtime**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Personal Runtime** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpColabRuntime
metadata:
  name: alice-gpu
  org: acme-corp
  env: prod
spec:
  location: us-central1
  runtimeTemplate:
    valueFrom:
      kind: GcpColabRuntimeTemplate
      name: gpu-t4
      fieldPath: status.outputs.name
  runtimeUser: alice@acme.com
  desiredState: STOPPED
```

```shell
planton apply -f colab-runtime.yaml
```

This gives Alice a GPU runtime that waits stopped until she needs it. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a `GcpColabRuntimeTemplate` from `runtimeTemplate`; declare one runtime per person in a team chart.

## Key Configuration

These are the most important decisions when configuring a runtime. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Who it belongs to** -- `runtimeUser` is the only person who can connect notebooks to it.

**Which template** -- the machine, network, and image all come from `runtimeTemplate`.

**Running or stopped** -- `desiredState` starts or stops the runtime on every apply; `STOPPED` keeps the disk and stops the compute bill. Leave it unset to let the user decide.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpColabRuntimeTemplate** | `runtimeTemplate` | `status.outputs.name` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The runtime's full resource name | Console links, automation |
| `runtime_id` | The runtime's id | SDK calls |
| `location` | The runtime's region | Regional clients |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Personal runtime** -- a runtime for one person on the team's standard template, left for them to start and stop. Start from the **Personal Runtime** preset.

**Parked GPU runtime** -- a GPU runtime kept stopped and auto-upgraded, ready on demand. Start from the **Parked GPU** preset.

## Works With

- [**GCP Colab Runtime Template**](/cloud-catalog/gcp-colab-runtime-template) -- the template the runtime is built from
- [**GCP Colab Schedule**](/cloud-catalog/gcp-colab-schedule) -- scheduled notebook runs
