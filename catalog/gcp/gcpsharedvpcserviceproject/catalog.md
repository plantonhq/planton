# GCP Shared VPC Service Project

Attaches a Google Cloud project to a Shared VPC host (`GcpSharedVpcHost`) as a SERVICE project, so its workloads can be placed in the host's subnetworks. The application team keeps its own project, IAM, quotas, and bill; the network team keeps the one set of networks every team shares. One block per attachment; a host may have many.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Shared VPC attachment** -- the `compute_shared_vpc_service_project` binding a service project to its host

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module whose identity holds `roles/compute.xpnAdmin` on the organization (or a folder above both projects). Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.

### GCP Shared VPC Host and Project

- **The host must be enabled** -- declare it with `GcpSharedVpcHost` and reference its `host_project_id` output, or pass the host's project ID as a literal.
- **The service project must exist** -- declare it with `GcpProject` and reference its `project_id` output, or pass the ID as a literal.
- **A project can be a service project of at most one host**, and a host cannot also be a service project.

## Deploy

### Console

Open the deployment store, find **GCP Shared VPC Service Project**, and click **Deploy**. The creation wizard asks for the host and the project. Start from the **Attach Service Project** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpSharedVpcServiceProject
metadata:
  name: payments-attach
  org: acme-corp
  env: prod
spec:
  hostProjectId:
    value: acme-network-host
  serviceProjectId:
    value: acme-payments-prod
```

```shell
planton apply -f shared-vpc-service-project.yaml
```

This attaches the payments project to the organization's network host. A Stack Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, the attachment references the host KIND (not the project) and the service `GcpProject` via ValueFromRef:

```yaml
spec:
  hostProjectId:
    valueFrom:
      kind: GcpSharedVpcHost
      name: network-host
      fieldPath: status.outputs.host_project_id
  serviceProjectId:
    valueFrom:
      kind: GcpProject
      name: payments-prod
      fieldPath: status.outputs.project_id
```

Referencing the host kind is what orders the chart: the project is enabled as a host before anything attaches to it, and detached before the host is disabled.

## Key Configuration

These are the most important decisions when configuring an attachment. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Host** -- the `GcpSharedVpcHost` (its `host_project_id` output). Immutable.

**Service project** -- the `GcpProject` being attached. Immutable: moving to another host is a detach and an attach.

**Deletion policy** -- empty (default) detaches on destroy, which Google refuses while any resource in the service project still uses a host subnetwork; `ABANDON` leaves the attachment in place for a project whose workloads must outlive the chart that attached it. This resource accepts no other value.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpSharedVpcHost** | `hostProjectId` | `status.outputs.host_project_id` |
| **GcpProject** | `serviceProjectId` | `status.outputs.project_id` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `service_project_id` | The attached service project's ID | Tooling |
| `host_project_id` | The host project's ID | Tooling |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Attach service project** -- the everyday shape: one project joins the host. Start from the **Attach Service Project** preset.

**Abandon on destroy** -- an attachment whose project outlives the chart. Start from the **Attach Abandon On Destroy** preset.

## Works With

- [**GCP Shared VPC Host**](/cloud-catalog/gcp-shared-vpc-host) -- the host this project attaches to
- [**GCP Project**](/cloud-catalog/gcp-project) -- the service project
- [**GCP Project IAM Member**](/cloud-catalog/gcp-project-iam-member) -- the `networkUser` grants on the host
- [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- the host's subnetworks the service project uses
