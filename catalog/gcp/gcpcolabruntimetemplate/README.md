# GCP Colab Runtime Template

A Colab Enterprise runtime template -- the machine, disk, network, image, and security settings every notebook runtime created from it gets. Admins publish templates; data scientists pick one in Colab Enterprise (or a `GcpColabRuntime` / `GcpColabSchedule` names it) and get a runtime that already meets the team's rules: the right machine and GPU, a private network, customer-managed encryption, and idle shutdown to cap spend.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- `aiplatform.googleapis.com` on the project (Colab Enterprise's API; never disabled on destroy)
- **Runtime template** -- a `colab_runtime_template` with the declared machine, disk, network, idle, security, and software settings

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with Vertex AI admin permissions on the project.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### Optional Dependencies

- **`GcpVpcNetwork`** / **`GcpSubnetwork`** -- a private network for runtimes (`networkSpec.network`, `networkSpec.subnetwork`).
- **`GcpKmsKey`** -- customer-managed encryption for runtime disks (`kmsKeyName`).

## Deploy

### CLI

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpColabRuntimeTemplate
metadata:
  name: standard-runtime
spec:
  location: us-central1
  machineSpec:
    machineType: e2-standard-4
  idleTimeout: 3600s
```

```shell
planton apply -f colab-runtime-template.yaml
```

## Configuration Reference

### Required Fields

| Field | Type | Description |
|-------|------|-------------|
| `location` | `string` | The Colab Enterprise region, e.g. `us-central1`. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `projectId` | `StringValueOrRef` | provider project | The project, as a literal or a `GcpProject` reference. |
| `runtimeTemplateId` | `string` | `metadata.name` | Lowercase letters, digits, hyphens. Immutable. |
| `displayName` | `string` | `metadata.name` | Up to 128 characters; mutable. |
| `description` | `string` | none | Immutable. |
| `labels` | `map<string,string>` | none | Merged under the platform attribution labels. A change replaces the template. |
| `machineSpec` | `object` | Google default | `machineType`, `acceleratorType`, `acceleratorCount`. Immutable. |
| `dataPersistentDiskSpec` | `object` | Google default | `diskType` (`pd-standard`, `pd-balanced`, `pd-ssd`, `pd-extreme`), `diskSizeGb` (10-65536). Immutable. |
| `networkSpec` | `object` | default network | `enableInternetAccess`, `network` (`GcpVpcNetwork` ref), `subnetwork` (`GcpSubnetwork` ref). Immutable. |
| `idleTimeout` | `string` | Google default | Seconds ending in `s`; `0s` disables, otherwise 600s-86400s. Immutable. |
| `eucDisabled` | `bool` | `false` | Block user credentials inside the runtime. Immutable. |
| `enableSecureBoot` | `bool` | `false` | Shielded VM Secure Boot. Immutable. |
| `networkTags` | `string[]` | none | Firewall targeting tags. Immutable. |
| `kmsKeyName` | `StringValueOrRef` | Google-managed | A `GcpKmsKey` reference or literal key path; mutable (new runtimes). |
| `softwareConfig` | `object` | none | `env[]`, `postStartupScriptConfig`, `colabImage.releaseName`; mutable. |
| `deletionPolicy` | `string` | `DELETE` | `DELETE`, `PREVENT`, or `ABANDON`. |

### Validation Rules

- `location` is a region; ids are lowercase letters, digits, hyphens.
- `acceleratorType` needs `acceleratorCount`; `diskSizeGb` needs `diskType`.
- Environment variable names are C identifiers.

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `name` | `string` | `projects/{project}/locations/{location}/notebookRuntimeTemplates/{runtime_template_id}` -- what runtimes and schedules take |
| `runtime_template_id` | `string` | The template's id |
| `location` | `string` | The template's region |

## Deployment Methods

### Pulumi (Go)

See [`iac/pulumi/README.md`](iac/pulumi/README.md) for Pulumi-specific deployment instructions.

### Terraform

See [`iac/tf/README.md`](iac/tf/README.md) for Terraform-specific deployment instructions.

## Important Notes

- **Most settings are fixed at creation.** Only the display name, the key, and the software config update in place. Changing anything else -- including a label -- replaces the template (existing runtimes keep running).
- **Idle shutdown is the cost control.** A notebook left open on a GPU bills until it stops; set `idleTimeout`.
- **Private runtimes need a way out.** With `enableInternetAccess: false`, give the subnetwork Private Google Access (or Cloud NAT) so notebooks can reach Google APIs and package mirrors.
- **Environment values are plain text.** Never put secrets in `softwareConfig.env`.

## Examples

For a complete example, see `e2e/manifest.yaml`. Scenario variants live under `e2e/scenarios/`.

## Related Components

- **GcpColabRuntime** -- a runtime assigned to a user from this template
- **GcpColabSchedule** -- scheduled notebook runs on this template's machine
- **GcpVpcNetwork** / **GcpSubnetwork** -- private networking for runtimes
- **GcpKmsKey** -- customer-managed encryption

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
