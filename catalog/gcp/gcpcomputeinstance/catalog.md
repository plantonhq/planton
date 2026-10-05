# GCP Compute Instance

Deploys a Compute Engine virtual machine with configurable machine type, boot disk, network interfaces, attached disks, scheduling options (including Spot VMs), Shielded/confidential hardening, GPU accelerators, service account bindings, and startup scripts. Every composition seam — the GCP project, VPCs and subnets, disks, reserved addresses, KMS keys, and the runtime service account — accepts ValueFromRef wiring, so the VM drops into an InfraChart without hand-copied identifiers.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Compute Engine Instance** -- a virtual machine in the specified GCP project and zone, configured with the chosen machine type, boot source, and network interfaces
- **Boot Disk** -- initialized from exactly one source: an OS image (fresh install), a snapshot (restore), or an existing bootable GcpComputeDisk; with configurable size, type (pd-standard, pd-balanced, pd-ssd, hyperdisk types), CMEK encryption, and auto-delete behavior
- **Network Interfaces** -- one or more vNICs with optional external IPv4/IPv6 access configs, static internal/external IPs (reserved GcpAddress references), and alias IP ranges
- **Disk Attachments** -- created only when `attachedDisks` is configured; first-class GcpComputeDisk resources attached in READ_WRITE or READ_ONLY mode, plus optional ephemeral local-SSD `scratchDisks`
- **Service Account Binding** -- created only when `serviceAccount` is configured; binds a GCP service account with specified OAuth scopes to the instance
- **Scheduling Configuration** -- created only when `scheduling` is configured; the provisioning model (STANDARD, SPOT, FLEX_START, RESERVATION_BOUND), automatic restart, host maintenance policy, lifetime limits, and sole-tenant node affinities
- **Hardening** -- Shielded VM boot-chain options, confidential computing (`confidentialInstanceConfig`), and GPU accelerators when configured
- **GCP Labels** -- resource metadata labels (resource name, kind, organization, environment) applied automatically for tracking and governance

## Before You Deploy

### Planton Setup

- **GCP Provider Connection** -- an active connection in the Connect module with credentials for the target GCP project. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or browser OAuth authentication modes.

### GCP Project

- **A GCP project** where the Compute Engine instance will be created. Provide the project ID directly or reference a GcpProject Infra Component via ValueFromRef.
- **Compute Engine API** enabled in the target project.
- **VPC network and subnet** in the target project and region. Provide the network or subnet self-link directly or reference GcpVpcNetwork and GcpSubnetwork Infra Components via ValueFromRef.
- **GCP service account** (recommended) -- a dedicated service account following least-privilege principles, referenced directly or via ValueFromRef to a GcpServiceAccount Infra Component.
- **For CMEK disks** -- the Compute Engine service agent must hold `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the referenced GcpKmsKey.

## Deploy

### Console

Open the deployment store, find **GCP Compute Instance**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Dev Spot VM** preset in the [Presets](#presets) tab for the cheapest full Linux machine, or **Hardened Web Server** for the production security posture.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpComputeInstance
metadata:
  name: app-server
  org: acme-corp
  env: prod
spec:
  projectId:
    value: "acme-prod-12345"
  zone: us-central1-a
  machineType: e2-standard-2
  bootDisk:
    image:
      value: debian-cloud/debian-12
    sizeGb: 50
    type: pd-ssd
  networkInterfaces:
    - subnetwork:
        value: "projects/acme-prod-12345/regions/us-central1/subnetworks/default"
```

```shell
planton apply -f compute-instance.yaml
```

This creates a Debian 12 VM with 2 vCPU, 8 GB RAM, 50 GB SSD boot disk, and no external IP; Spot VM scheduling, service account, and deletion protection are not configured. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the instance to a GCP project, subnet, and service account deployed in the same InfraPipeline:

```yaml
spec:
  projectId:
    valueFrom:
      kind: GcpProject
      name: production-project
      fieldPath: status.outputs.project_id
  networkInterfaces:
    - subnetwork:
        valueFrom:
          kind: GcpSubnetwork
          name: production-subnet
          fieldPath: status.outputs.subnetwork_self_link
  serviceAccount:
    email:
      valueFrom:
        kind: GcpServiceAccount
        name: app-server-sa
        fieldPath: status.outputs.email
```

The InfraPipeline resolves the dependency graph, deploys the project, subnet, and service account first, then provisions the Compute Engine instance with the resolved values.

## Key Configuration

These are the most important decisions when configuring a Compute Engine instance. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Machine type and zone** -- The `machineType` field sets CPU and memory (e.g., `e2-standard-2` for 2 vCPU / 8 GB, `n2-standard-4` for 4 vCPU / 16 GB, or custom shapes like `custom-6-20480`). The `zone` determines physical placement and must match the subnet's region; it is immutable, and attached disks must live in the same zone.

**Boot disk** -- Exactly one source: `bootDisk.image` (a `GcpComputeImage` reference to pin a golden build, or a literal OS image family like `debian-cloud/debian-12`), `bootDisk.sourceSnapshot`, or `bootDisk.sourceDisk` (an existing bootable GcpComputeDisk). Choose `bootDisk.type` based on I/O requirements: `pd-balanced` for the sensible default, `pd-ssd` for high IOPS, hyperdisk types for tunable performance.

**Data disks** -- Durable data belongs on `attachedDisks` entries referencing first-class GcpComputeDisk resources by `self_link` -- the data survives this VM. Ephemeral `scratchDisks` (local SSD) offer extreme IOPS but lose contents on stop or preemption.

**Networking** -- Configure `networkInterfaces` with a VPC network or subnetwork reference. Omit `accessConfigs` for private-only instances (use Cloud NAT for internet egress). Add an access config for an external IP -- reference a reserved EXTERNAL GcpAddress from `natIp` to keep the address across rebuilds, and a reserved INTERNAL GcpAddress from `networkIp` for a stable private address.

**Spot and FLEX_START VMs** -- Set `scheduling.provisioningModel: SPOT` for 60-91% cost savings with the trade-off of potential preemption (the legacy preemptible flag is derived automatically). Use `scheduling.instanceTerminationAction: STOP` to preserve the stopped VM on preemption, or `DELETE` for ephemeral workloads. `FLEX_START` trades a flexible start time for discounted capacity (bound the runtime with `maxRunDurationSeconds`); `RESERVATION_BOUND` pins the VM to one named reservation.

**Security** -- `shieldedInstanceConfig` hardens the boot chain (secure boot is the deliberate opt-in), `confidentialInstanceConfig` enables hardware memory encryption (supported machine families only; requires `onHostMaintenance: TERMINATE`), CMEK keys apply at every level (per-disk `kmsKey`, encrypted image/snapshot sources, and instance-level `instanceEncryptionKey`), and `deletionProtection` guards the VM object -- the data levers remain the disks' own lifecycles, and `deletionPolicy: PREVENT`/`ABANDON` controls what a destroy may do.

**Service account** -- Bind a dedicated `serviceAccount.email` with the single cloud-platform scope instead of the default Compute Engine service account. Reference a GcpServiceAccount Infra Component via ValueFromRef.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **GcpProject** | `projectId` | `status.outputs.project_id` |
| **GcpComputeImage** (optional) | `bootDisk.image` | `status.outputs.self_link` |
| **GcpComputeDisk** (optional) | `bootDisk.sourceDisk` | `status.outputs.self_link` |
| **GcpComputeDisk** (optional) | `attachedDisks[].source` | `status.outputs.self_link` |
| **GcpKmsKey** (optional) | `bootDisk.kmsKey`, `attachedDisks[].kmsKey`, `bootDisk.sourceImageEncryption.kmsKey`, `bootDisk.sourceSnapshotEncryption.kmsKey`, `instanceEncryptionKey.kmsKey` | `status.outputs.key_id` |
| **GcpVpcNetwork** (optional) | `networkInterfaces[].network` | `status.outputs.network_self_link` |
| **GcpSubnetwork** (optional) | `networkInterfaces[].subnetwork` | `status.outputs.subnetwork_self_link` |
| **GcpAddress** (optional) | `networkInterfaces[].networkIp` (INTERNAL), `networkInterfaces[].accessConfigs[].natIp` (EXTERNAL) | `status.outputs.address` |
| **GcpServiceAccount** (optional) | `serviceAccount.email` | `status.outputs.email` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `instance_name` | Name of the Compute Engine instance | Monitoring dashboards, SSH access |
| `instance_id` | Unique numeric instance identifier | API references, audit logs |
| `self_link` | GCP resource self link | IAM bindings, instance groups |
| `internal_ip` | Private IP address | Application connection strings, GcpDnsRecord targets |
| `external_ip` | Public IP address (empty for private VMs) | SSH access, external service endpoints |
| `status` | Current instance status (RUNNING, TERMINATED, ...) | Health monitoring, automation triggers |
| `zone` | Zone where the instance is located | Placement checks for zone-bound resources (disks, addresses) |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Dev Spot VM** -- The cheapest full Linux machine: an e2-medium Spot VM on the default network with an ephemeral external IP, deleted outright on preemption. Suitable for dev boxes, CI runners, and disposable experiments. Start from the **Dev Spot VM** preset.

**Hardened Web Server** -- The production security posture: a Shielded VM on a custom-mode subnetwork with no external IP, a dedicated least-privilege service account, OS Login instead of static SSH keys, and firewall targeting via network tags. Start from the **Hardened Web Server** preset.

**Stateful Data VM** -- The database-on-VM pattern where everything durable outlives the instance: data on a referenced GcpComputeDisk, a stable internal address from a referenced GcpAddress, deletion protection on the VM object, and live-migration for zero-downtime maintenance. Start from the **Stateful Data VM** preset.

## Works With

- [**GCP Project**](/infra-catalog/gcp-project) -- provides the GCP project where the instance is created
- [**GCP Compute Disk**](/infra-catalog/gcp-compute-disk) -- provides durable boot and data disks that survive the instance
- [**GCP VPC Network**](/infra-catalog/gcp-vpc-network) -- provides the VPC network for instance networking
- [**GCP Subnetwork**](/infra-catalog/gcp-subnetwork) -- provides the subnet for instance placement
- [**GCP Address**](/infra-catalog/gcp-address) -- provides reserved static internal and external IPs
- [**GCP Service Account**](/infra-catalog/gcp-service-account) -- provides the runtime identity for the instance
