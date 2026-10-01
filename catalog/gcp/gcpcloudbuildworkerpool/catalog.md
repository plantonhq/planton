# GCP Cloud Build Worker Pool

Gives your builds their own machines. A private worker pool runs Cloud Build jobs on dedicated workers in one region, so builds can reach private package indexes, internal artifact stores, databases they migrate, and GKE control planes on private endpoints -- and so a team can choose bigger machines or nested virtualization. Many triggers and Cloud Deploy targets can share one pool.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **API enablement** -- the Cloud Build API on the pool's project
- **Worker pool** -- the private pool, optionally peered into a VPC network or attached through Private Service Connect

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with permission to manage Cloud Build worker pools in the target project. Map it as the default for your environment, or specify it explicitly when creating the Cloud Resource.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Account

- **A VPC with private services access** -- only for a peered pool: the network needs a Service Networking connection before the pool can peer into it.

## Deploy

### Console

Open the deployment store, find **GCP Cloud Build Worker Pool**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Default Private Pool** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpCloudBuildWorkerPool
metadata:
  name: private-builds
  org: acme-corp
  env: prod
spec:
  projectId:
    value: acme-ci
  location: us-central1
  workerConfig:
    machineType: e2-standard-4
    diskSizeGb: 200
```

```shell
planton apply -f cloud-build-worker-pool.yaml
```

This creates a pool of e2-standard-4 build machines in us-central1. A Stack Job tracks the provisioning in real time.

### InfraChart

Reference a project's `status.outputs.project_id` from `projectId`, and a VPC network's `status.outputs.network_self_link` from `networkConfig.peeredNetwork`.

## Key Configuration

These are the most important decisions when configuring a worker pool. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Region** -- builds that use the pool run in its region, and a trigger naming the pool must be in the same region.

**Networking** -- peer the workers into your VPC, attach them through Private Service Connect, or neither (public egress on Google's network). The choice is fixed at creation.

**Machines** -- the machine type sets the per-minute build rate; the disk size and nested virtualization serve heavy image builds and emulator tests.

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** | `networkConfig.peeredNetwork` | `status.outputs.network_self_link` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `name` | The pool's full resource name | A trigger's build options, a Cloud Deploy target's execution environment |
| `worker_pool_id` | The pool's ID | Tooling |
| `state` | The pool's state | Readiness checks |
| `uid` | The pool's unique identifier | Audits |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Default private pool** -- dedicated machines with public egress. Start from the **Default Private Pool** preset.

**Pool inside your VPC** -- workers peered into a network, without public IPs. Start from the **Peered Private Pool** preset.

## Works With

- [**GCP Cloud Build Trigger**](/cloud-catalog/gcp-cloud-build-trigger) -- builds that run on the pool
- [**GCP Deploy Target**](/cloud-catalog/gcp-deploy-target) -- Cloud Deploy render and deploy jobs that run on the pool
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) -- the network a peered pool joins
- [**GCP Service Networking Connection**](/cloud-catalog/gcp-service-networking-connection) -- private services access the peering needs
