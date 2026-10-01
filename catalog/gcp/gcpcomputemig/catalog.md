# GCP Compute MIG

Creates a Compute Engine Managed Instance Group — a self-healing, optionally auto-scaling fleet of identical VMs. One resource manages the instance template, the group manager, an optional autoscaler, stateful per-instance configs, and queued resize requests, zonal or regional. The backbone of classic production serving on GCP: put the group's `instance_group` output behind a backend service and the fleet serves the modeled HTTPS load-balancing family.

## What Gets Created

When you deploy this Cloud Resource, the IaC module provisions:

- **Instance template** -- the immutable VM shape (machine type, disks, networking, identity, scheduling); every template change rotates to a fresh template natively
- **Instance group manager** -- the fleet controller: size, canary versions, rolling updates, auto-healing, stateful disks/IPs, standby pools
- **Autoscaler** (when `autoscaler` is set) -- CPU / LB-capacity / custom-metric / schedule-driven scaling between min and max replicas
- **Per-instance configs** (per `perInstanceConfigs` entry) -- stateful name/disk/IP overrides pinned to individual instances
- **Resize requests** (per `resizeRequests` entry) -- queued one-shot capacity asks (Dynamic Workload Scheduler)
- **Compute Engine API enablement** -- `compute.googleapis.com` enabled in the target project (never disabled on destroy)

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project.
- **Planton Runner** -- required when using Runner-based credential delivery.

### GCP Project

- **A GCP project** to host the group (directly or via a GcpProject reference) and a **VPC network** the fleet attaches to.
- **IAM**: the deploying identity needs `roles/compute.admin` (or narrower instance-group/template roles), plus `iam.serviceAccounts.actAs` on the template's service account when one is set.

## Deploy

### Console

Open the deployment store, find **GCP Compute MIG**, and click **Deploy**. The creation wizard walks you through environment and connection configuration, the VM template, placement (zonal or regional), scaling, and the rollout policy. Start from the **Autoscaled Web Tier** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpComputeMig
metadata:
  name: web-tier
  org: acme-corp
  env: prod
spec:
  region: us-central1
  template:
    machineType: e2-small
    disks:
      - boot: true
        sourceImage:
          value: debian-cloud/debian-12
    networkInterfaces:
      - subnetwork:
          valueFrom:
            kind: GcpSubnetwork
            name: prod-subnet
            fieldPath: status.outputs.subnetwork_self_link
    startupScript: |
      #!/bin/bash
      systemctl start my-app
  namedPorts:
    - name: http
      port: 8080
  autoscaler:
    minReplicas: 2
    maxReplicas: 10
    cpuTarget: 0.6
  updatePolicy:
    minimalAction: REPLACE
    type: PROACTIVE
    maxSurgeFixed: 3
```

```shell
planton apply -f mig.yaml
```

This creates a regional fleet of e2-small Debian 12 VMs scaling between 2 and 10 replicas on a 60% CPU target, rolled proactively with a three-instance surge on every template change. A Stack Job tracks the provisioning in real time.

### InfraChart

The serving backbone in one chart: this group, a GcpHealthCheck wired into `autoHealing`, and a GcpBackendService consuming the `instance_group` output — the whole HTTPS-LB family composes behind it:

```yaml
spec:
  autoHealing:
    healthCheck:
      valueFrom:
        kind: GcpHealthCheck
        name: web-hc
        fieldPath: status.outputs.self_link
    initialDelaySec: 120
  template:
    networkInterfaces:
      - subnetwork:
          valueFrom:
            kind: GcpSubnetwork
            name: prod-subnet
            fieldPath: status.outputs.subnetwork_self_link
```

The InfraPipeline deploys the network and health check first, then the group; the backend service downstream picks up `instance_group` the same way.

## Key Configuration

These are the most important decisions when configuring a managed instance group. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**zone / region** -- exactly one. Zonal is simplest; regional spreads VMs across the region's zones so a zone outage takes down only part of the fleet (add `distributionPolicy` to shape the spread).

**template** -- the VM shape. IMMUTABLE in GCP: every change here rotates the template and rolls the group per `updatePolicy` (labels are the one in-place exception). Keep fleets private (no `accessConfigs`) behind Cloud NAT.

**autoscaler vs targetSize** -- one or the other, never both (they fight over the size on every apply). The autoscaler scales on CPU, load-balancer serving capacity, custom Cloud Monitoring metrics, or calendar schedules; `scaleInControl` caps how fast it shrinks.

**updatePolicy** -- `PROACTIVE` + `REPLACE` with a surge budget is the standard production rollout; `OPPORTUNISTIC` for fleets where changes should wait for natural churn. `RECREATE` preserves instance names (stateful) but needs an unavailability budget above zero.

**autoHealing** -- recreate VMs that fail an application health check (a GcpHealthCheck reference). Size `initialDelaySec` to the app's cold start or the group repair-loops slow-booting instances.

**statefulDisks / perInstanceConfigs** -- the stateful story: preserved disks and per-instance identity that survive recreation. Stateful groups want `RECREATE` replacement and (regional) `instanceRedistributionType: NONE`.

**deletionPolicy** -- what a destroy does to the whole stack: `DELETE` (default), `PREVENT` (destroy fails), or `ABANDON` (leave everything running unmanaged).

## Outputs and Dependencies

### What This Component Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** (optional) | `projectId` | `status.outputs.project_id` |
| **GcpVpcNetwork** (per NIC) | `template.networkInterfaces[].network` | `status.outputs.network_self_link` |
| **GcpSubnetwork** (per NIC) | `template.networkInterfaces[].subnetwork` | `status.outputs.subnetwork_self_link` |
| **GcpHealthCheck** (optional) | `autoHealing.healthCheck` | `status.outputs.self_link` |
| **GcpServiceAccount** (optional) | `template.serviceAccount.email` | `status.outputs.email` |
| **GcpKmsKey** (optional) | `template.disks[].diskEncryption.kmsKey` | `status.outputs.key_id` |
| **GcpComputeImage** (optional) | `template.disks[].sourceImage` | `status.outputs.self_link` |
| **GcpComputeDisk** (optional) | `template.disks[].source`, `perInstanceConfigs[].preservedState.disks[].source` | `status.outputs.self_link` |
| **GcpAddress** (optional) | `perInstanceConfigs[].preservedState.externalIps[].address` | `status.outputs.address` |

### What This Component Provides

After provisioning, `status.outputs` contains values that downstream Cloud Resources can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `instance_group` | Full instance-group URL | A GcpBackendService backend's `group` — the LB edge |
| `self_link` | Group manager self link | gcloud commands, monitoring scopes |
| `current_template_self_link` | The active template's link | Deploy tracking — changes on every rotation |
| `mig_name` | The group name | Display, per-group metrics filters |
| `location` | Zone or region | Scope-compatibility checks downstream |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Autoscaled web tier** -- regional group, CPU-target autoscaler, proactive rolling updates, named port for the backend service. Start from the **Autoscaled Web Tier** preset.

**Regional HA with auto-healing** -- spread across zones, application-level health-check repairs, conservative scale-in. Start from the **Regional HA Group** preset.

**Stateful fleet** -- preserved disks and instance identity through recreation (brokers, databases-on-VM). Start from the **Stateful Group** preset.

## Works With

- [**GCP Backend Service**](/cloud-catalog/gcp-backend-service) -- consumes the `instance_group` output as a backend
- [**GCP Health Check**](/cloud-catalog/gcp-health-check) -- drives auto-healing and LB health
- [**GCP VPC Network**](/cloud-catalog/gcp-vpc-network) / [**GCP Subnetwork**](/cloud-catalog/gcp-subnetwork) -- the fleet's network home
- [**GCP Service Account**](/cloud-catalog/gcp-service-account) -- the VMs' workload identity
- [**GCP KMS Key**](/cloud-catalog/gcp-kms-key) -- CMEK for the fleet's disks
- [**GCP Compute Instance**](/cloud-catalog/gcp-compute-instance) -- the single-VM sibling for pets rather than cattle
