# GCP Shared VPC Host

Enables a Google Cloud project as a Shared VPC HOST: the project whose VPC networks and subnetworks other projects (service projects, `GcpSharedVpcServiceProject`) attach to and deploy their workloads into. Shared VPC is how an organization keeps one network team owning one set of networks while many application teams deploy from their own projects with their own IAM, quotas, and bill. Enabling the host is a one-bit act with its own lifecycle, which is why it is a block of its own rather than a flag on the project.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Shared VPC host enablement** -- the `compute_shared_vpc_host_project` flag on the project

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module whose identity holds `roles/compute.xpnAdmin` on the organization (or a folder above the project). Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Project

- **The project must exist** -- declare it with `GcpProject` and reference its `project_id` output, pass the ID as a literal, or leave `projectId` empty to enable the project the credentials are configured for.
- **`roles/compute.xpnAdmin` is organization-level.** A project-level grant does not allow enabling a host.
- **A project is a host OR a service project**, never both.

## Deploy

### Console

Open the deployment store, find **GCP Shared VPC Host**, and click **Deploy**. The creation wizard asks only for the project. Start from the **Enable Host** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSharedVpcHost
metadata:
  name: network-host
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-network-host
  deletionPolicy: PREVENT
```

```shell
planton apply -f shared-vpc-host.yaml
```

This enables `acme-network-host` as the organization's network host, guarded against accidental destroy. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the host references its `GcpProject` via ValueFromRef, and every `GcpSharedVpcServiceProject` references the host's `host_project_id`:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: network-host
      fieldPath: status.outputs.project_id
```

The InfraPipeline deploys the project first, then enables it as a host, then attaches every service project that points at `status.outputs.host_project_id` -- and destroys in the reverse order, which is the order Google requires.

## Key Configuration

These are the most important decisions when configuring a Shared VPC host. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Project** -- the host. Empty means the provider's default project, so "make the project I am deploying into the host" needs no configuration. Immutable: a different host is a new resource.

**Deletion policy** -- `DELETE` (default) disables the host role and fails while any service project is still attached (destroy the attachments first -- a chart's dependency order does this when they reference the host); `PREVENT` fails the destroy, the guard for the host every service project in the organization depends on; `ABANDON` leaves the project a host with every attachment intact.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `host_project_id` | The project ID enabled as the host (resolved, even when the spec named none) | A `GcpSharedVpcServiceProject`'s `hostProjectId` |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Enable host** -- the whole shape: one project becomes the host. Start from the **Enable Host** preset.

## Works With

- [**GCP Shared VPC Service Project**](/cloud-catalog/gcp-shared-vpc-service-project) -- attaches a service project to this host
- [**GCP Project**](/cloud-catalog/gcp-project) -- the project being enabled
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- the networks the host shares
- [**GCP Project IAM Member**](/cloud-catalog/gcp-project-iam-member) -- the `networkUser` grants service projects need
