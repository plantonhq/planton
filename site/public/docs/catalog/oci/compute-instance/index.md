---
title: "Compute Instance"
description: "Compute Instance deployment documentation"
icon: "package"
order: 100
componentName: "ocicomputeinstance"
---

# OCI Compute Instance

Deploys an Oracle Cloud Infrastructure compute instance — a virtual machine or bare metal host — with flexible shape sizing, primary VNIC networking, and cloud-init metadata support. Flex shapes allow precise allocation of OCPUs and memory, while optional configurations cover preemptible pricing, platform-level security (Secure Boot, TPM), and Oracle Cloud Agent management.

## What Gets Created

When you deploy an OciComputeInstance resource, Planton provisions:

- **Compute Instance** — an `oci_core_instance` resource in the specified compartment and availability domain. The instance is created with the chosen shape, boots from the specified image or boot volume, and is attached to a subnet via its primary VNIC. Standard Planton freeform tags are applied for resource tracking.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the instance will be created — literal value or reference to an OciCompartment resource
- **A subnet OCID** for the primary VNIC — literal value or reference to an OciSubnet resource
- **An availability domain name** in the target region (e.g., `Ixxj:US-ASHBURN-AD-1`)
- **A compute shape name** (e.g., `VM.Standard.E4.Flex`, `VM.Standard.A1.Flex`, `BM.Standard3.64`)
- **An image OCID** for the boot source, or an existing boot volume OCID to clone

## Quick Start

Create a file `instance.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciComputeInstance
metadata:
  name: my-instance
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciComputeInstance.my-instance
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Ixxj:US-ASHBURN-AD-1"
  shape: "VM.Standard.E4.Flex"
  shapeConfig:
    ocpus: 1
    memoryInGbs: 16
  sourceDetails:
    sourceType: image
    sourceId: "ocid1.image.oc1.iad.example"
  createVnicDetails:
    subnetId:
      value: "ocid1.subnet.oc1.iad.example"
```

Deploy:

```shell
planton apply -f instance.yaml
```

This creates a 1-OCPU, 16 GB VM on the E4 Flex shape in the specified availability domain and subnet. The instance boots from the given image, receives a private IP from the subnet's CIDR, and inherits the subnet's public IP assignment policy. The instance ID, IP addresses, boot volume ID, and availability domain are exported as stack outputs.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the instance will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `availabilityDomain` | `string` | Availability domain for instance placement (e.g., `Ixxj:US-ASHBURN-AD-1`). Changing this forces recreation. | Minimum 1 character |
| `shape` | `string` | Compute shape determining the hardware profile (e.g., `VM.Standard.E4.Flex`, `VM.Standard.A1.Flex`, `BM.Standard3.64`). Flex shapes require `shapeConfig`. | Minimum 1 character |
| `sourceDetails` | `SourceDetails` | Boot source configuration. See [sourceDetails fields](#sourcedetails-fields). | Required |
| `createVnicDetails` | `CreateVnicDetails` | Primary VNIC configuration determining network placement. See [createVnicDetails fields](#createvnicdetails-fields). | Required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Human-readable name shown in the OCI Console. |
| `shapeConfig` | `ShapeConfig` | — | Resource allocation for flex shapes. Required when the shape name contains "Flex". See [shapeConfig fields](#shapeconfig-fields). |
| `metadata` | `map<string, string>` | — | Key-value pairs passed to the instance. Common keys: `ssh_authorized_keys` (newline-separated public keys), `user_data` (base64-encoded cloud-init script). |
| `faultDomain` | `string` | Auto-distributed | Fault domain within the availability domain (e.g., `FAULT-DOMAIN-1`). OCI auto-distributes across fault domains when unspecified. |
| `isPvEncryptionInTransitEnabled` | `bool` | — | Enables in-transit encryption for paravirtualized boot and data volume attachments. Changing this forces recreation. |
| `agentConfig` | `AgentConfig` | — | Oracle Cloud Agent plugin configuration. See [agentConfig fields](#agentconfig-fields). |
| `availabilityConfig` | `AvailabilityConfig` | — | Maintenance and recovery behavior. See [availabilityConfig fields](#availabilityconfig-fields). |
| `launchOptions` | `LaunchOptions` | — | Low-level boot volume, network, and firmware settings. Most users can omit this. See [launchOptions fields](#launchoptions-fields). |
| `instanceOptions` | `InstanceOptions` | — | Instance Metadata Service (IMDS) endpoint settings. See [instanceOptions fields](#instanceoptions-fields). |
| `preemptibleInstanceConfig` | `PreemptibleInstanceConfig` | — | Configures the instance as preemptible (spot-like) for cost savings. See [preemptibleInstanceConfig fields](#preemptibleinstanceconfig-fields). |
| `capacityReservationId` | `StringValueOrRef` | — | OCID of a capacity reservation to launch against. |
| `dedicatedVmHostId` | `StringValueOrRef` | — | OCID of a dedicated VM host for physical isolation (compliance or licensing requirements). |
| `platformConfig` | `PlatformConfig` | — | Platform-level security and hardware configuration. See [platformConfig fields](#platformconfig-fields). |

### sourceDetails Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `sourceType` | `enum` | Boot source type. Values: `image` (platform or custom image), `boot_volume` (clone an existing boot volume). | Required |
| `sourceId` | `string` | OCID of the image or boot volume to launch from. | Minimum 1 character |
| `bootVolumeSizeInGbs` | `int64` | Boot volume size in GiB. Defaults to the image's minimum size when launching from an image. Must be >= the image minimum. | Optional |
| `bootVolumeVpusPerGb` | `int64` | Volume performance units per GiB. `10` = Balanced, `20` = Higher Performance, `30`–`120` = Ultra High Performance. Defaults to `10`. | Optional |
| `kmsKeyId` | `StringValueOrRef` | OCID of a KMS key for boot volume encryption at rest. Can reference an OciKmsKey resource via `valueFrom`. | Optional |

### createVnicDetails Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `subnetId` | `StringValueOrRef` | OCID of the subnet for the primary VNIC. Can reference an OciSubnet resource via `valueFrom`. | Required |
| `nsgIds` | `StringValueOrRef[]` | OCIDs of network security groups to associate with the VNIC. Can reference OciSecurityGroup resources via `valueFrom`. | Maximum 5 items |
| `assignPublicIp` | `bool` | Whether to assign a public IP. When unset, uses the subnet default (public subnets assign; private subnets do not). | Optional |
| `displayName` | `string` | Display name for the VNIC in the OCI Console. | Optional |
| `hostnameLabel` | `string` | DNS hostname label within the subnet's DNS domain. Must be alphanumeric, start with a letter, max 63 characters. | Optional |
| `privateIp` | `string` | Specific private IP to assign within the subnet's CIDR. OCI auto-assigns when omitted. | Optional |
| `skipSourceDestCheck` | `bool` | Disables source/destination checking on the VNIC. Required for NAT instances or virtual routers. | Optional |
| `assignPrivateDnsRecord` | `bool` | Whether to register a private DNS record for the VNIC. | Optional |

### shapeConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `ocpus` | `float` | Number of OCPUs. Each OCPU maps to a physical core with simultaneous multi-threading. |
| `memoryInGbs` | `float` | Memory in GiB. Flex shapes allow a range per OCPU — consult the shape documentation for valid ratios. |
| `baselineOcpuUtilization` | `string` | Baseline utilization for burstable instances. Values: `BASELINE_1_8`, `BASELINE_1_2`, `BASELINE_1_1`. Only applicable to burstable shapes. |
| `nvmes` | `int32` | Number of NVMe drives. Only applicable to dense-IO shapes. |

### agentConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `areAllPluginsDisabled` | `bool` | Disables all Oracle Cloud Agent plugins. |
| `isManagementDisabled` | `bool` | Disables the management agent (OS Management, etc.). |
| `isMonitoringDisabled` | `bool` | Disables the monitoring agent (Compute Instance Monitoring). |
| `pluginsConfig` | `PluginConfig[]` | Per-plugin overrides. Each entry requires `name` (plugin name, e.g., `Vulnerability Scanning`) and `desiredState` (enum: `enabled`, `disabled`). |

### availabilityConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `isLiveMigrationPreferred` | `bool` | When `true`, OCI prefers live migration over reboot during infrastructure maintenance. |
| `recoveryAction` | `enum` | Action on unplanned host failure. Values: `restore_instance` (restart on a new host), `stop_instance` (remain stopped). |

### launchOptions Fields

| Field | Type | Description |
|-------|------|-------------|
| `bootVolumeType` | `string` | Boot volume attachment emulation type. Values: `ISCSI`, `SCSI`, `IDE`, `VFIO`, `PARAVIRTUALIZED`. |
| `networkType` | `string` | Network interface emulation type. Values: `E1000`, `VFIO`, `PARAVIRTUALIZED`. |
| `firmware` | `enum` | Instance firmware. Values: `bios`, `uefi_64`. |
| `isPvEncryptionInTransitEnabled` | `bool` | In-transit encryption for the boot volume attachment within launch options. |
| `isConsistentVolumeNamingEnabled` | `bool` | Consistent device naming for attached volumes (e.g., `/dev/oracleoci/...`). |

### instanceOptions Fields

| Field | Type | Description |
|-------|------|-------------|
| `areLegacyImdsEndpointsDisabled` | `bool` | When `true`, disables legacy IMDSv1 endpoints. Recommended for security — use only the IMDSv2 token-based endpoint. |

### preemptibleInstanceConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `preserveBootVolume` | `bool` | When `true`, the boot volume is preserved when the instance is preempted. When `false`, both instance and boot volume are terminated. |

### platformConfig Fields

| Field | Type | Description |
|-------|------|-------------|
| `type` | `enum` | Platform type matching the instance shape family. Required when `platformConfig` is set. Values: `amd_milan_bm`, `amd_milan_bm_gpu`, `amd_rome_bm`, `amd_rome_bm_gpu`, `amd_vm`, `generic_bm`, `intel_icelake_bm`, `intel_skylake_bm`, `intel_vm`. |
| `isSecureBootEnabled` | `bool` | Verifies boot software signatures. VM and bare metal shapes. |
| `isMeasuredBootEnabled` | `bool` | Records integrity measurements in the TPM. VM and bare metal shapes. |
| `isTrustedPlatformModuleEnabled` | `bool` | Enables the TPM for secure key storage. VM and bare metal shapes. |
| `isMemoryEncryptionEnabled` | `bool` | Enables AMD SEV or Intel TME memory encryption. VM and bare metal shapes. |
| `isSymmetricMultiThreadingEnabled` | `bool` | Enables SMT/Hyperthreading. Bare metal shapes only. |
| `areVirtualInstructionsEnabled` | `bool` | Enables nested virtualization (AMD-V or VT-x). Bare metal shapes only. |
| `isAccessControlServiceEnabled` | `bool` | Enables Access Control Service for PCI passthrough isolation. Bare metal shapes only. |
| `isInputOutputMemoryManagementUnitEnabled` | `bool` | Enables IOMMU for device memory protection. Bare metal shapes only. |
| `numaNodesPerSocket` | `string` | NUMA nodes per socket configuration. Values: `NPS0`, `NPS1`, `NPS2`, `NPS4`. Bare metal shapes only. |
| `percentageOfCoresEnabled` | `int32` | Percentage of cores to enable on the instance. Bare metal shapes only. |

## Examples

### Minimal Flex VM

A VM with the AMD E4 Flex shape, 1 OCPU, and 16 GB memory — the simplest production-capable configuration:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciComputeInstance
metadata:
  name: web-server
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciComputeInstance.web-server
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Ixxj:US-ASHBURN-AD-1"
  shape: "VM.Standard.E4.Flex"
  shapeConfig:
    ocpus: 1
    memoryInGbs: 16
  sourceDetails:
    sourceType: image
    sourceId: "ocid1.image.oc1.iad.example"
  createVnicDetails:
    subnetId:
      value: "ocid1.subnet.oc1.iad.example"
```

### Private Application Server with Cloud-Init

A private instance bootstrapped via cloud-init, secured by NSGs referenced from Planton-managed resources. No public IP; the instance is accessible only through the private subnet:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciComputeInstance
metadata:
  name: app-server
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciComputeInstance.app-server
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Ixxj:US-ASHBURN-AD-1"
  shape: "VM.Standard.E4.Flex"
  shapeConfig:
    ocpus: 2
    memoryInGbs: 32
  sourceDetails:
    sourceType: image
    sourceId: "ocid1.image.oc1.iad.example"
    bootVolumeSizeInGbs: 100
    bootVolumeVpusPerGb: 20
  createVnicDetails:
    subnetId:
      valueFrom:
        kind: OciSubnet
        name: private-subnet
        fieldPath: status.outputs.subnetId
    nsgIds:
      - valueFrom:
          kind: OciSecurityGroup
          name: app-nsg
          fieldPath: status.outputs.networkSecurityGroupId
    assignPublicIp: false
    hostnameLabel: "appserver"
  metadata:
    ssh_authorized_keys: "ssh-rsa AAAA... user@workstation"
    user_data: "IyEvYmluL2Jhc2gKZWNobyAiSGVsbG8gZnJvbSBjbG91ZC1pbml0Ig=="
  faultDomain: "FAULT-DOMAIN-1"
  instanceOptions:
    areLegacyImdsEndpointsDisabled: true
```

### Preemptible Batch Worker

A cost-optimized preemptible instance for fault-tolerant batch processing. OCI can reclaim the instance when capacity is needed; the boot volume is preserved for resuming work:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciComputeInstance
metadata:
  name: batch-worker
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciComputeInstance.batch-worker
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Ixxj:US-ASHBURN-AD-2"
  shape: "VM.Standard.E4.Flex"
  shapeConfig:
    ocpus: 4
    memoryInGbs: 64
  sourceDetails:
    sourceType: image
    sourceId: "ocid1.image.oc1.iad.example"
  createVnicDetails:
    subnetId:
      value: "ocid1.subnet.oc1.iad.example"
    assignPublicIp: false
  preemptibleInstanceConfig:
    preserveBootVolume: true
```

### Security-Hardened Production VM

A production instance with platform security enabled — Secure Boot, Measured Boot, TPM, and memory encryption — plus IMDSv2-only access, in-transit encryption, and live migration preference:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciComputeInstance
metadata:
  name: secure-vm
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciComputeInstance.secure-vm
  env: prod
  org: acme
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: secure-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Ixxj:US-ASHBURN-AD-1"
  shape: "VM.Standard.E4.Flex"
  shapeConfig:
    ocpus: 4
    memoryInGbs: 64
  sourceDetails:
    sourceType: image
    sourceId: "ocid1.image.oc1.iad.example"
    bootVolumeSizeInGbs: 200
    bootVolumeVpusPerGb: 20
    kmsKeyId:
      value: "ocid1.key.oc1.iad.example"
  createVnicDetails:
    subnetId:
      valueFrom:
        kind: OciSubnet
        name: private-subnet
        fieldPath: status.outputs.subnetId
    nsgIds:
      - valueFrom:
          kind: OciSecurityGroup
          name: secure-nsg
          fieldPath: status.outputs.networkSecurityGroupId
    assignPublicIp: false
    hostnameLabel: "securevm"
  isPvEncryptionInTransitEnabled: true
  platformConfig:
    type: amd_vm
    isSecureBootEnabled: true
    isMeasuredBootEnabled: true
    isTrustedPlatformModuleEnabled: true
    isMemoryEncryptionEnabled: true
  instanceOptions:
    areLegacyImdsEndpointsDisabled: true
  availabilityConfig:
    isLiveMigrationPreferred: true
    recoveryAction: restore_instance
  metadata:
    ssh_authorized_keys: "ssh-rsa AAAA... admin@workstation"
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `instance_id` | `string` | OCID of the compute instance. |
| `private_ip` | `string` | Private IP address of the primary VNIC. |
| `public_ip` | `string` | Public IP address of the primary VNIC. Empty when no public IP is assigned. |
| `boot_volume_id` | `string` | OCID of the boot volume attached to the instance. |
| `availability_domain` | `string` | Availability domain where the instance was placed. |

## Related Components

- [OciCompartment](/docs/catalog/oci/compartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
- [OciSubnet](/docs/catalog/oci/subnet) — provides the subnet referenced by `createVnicDetails.subnetId` via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/network-security-group) — manages network security rules referenced by `createVnicDetails.nsgIds` via `valueFrom`
- [OciVcn](/docs/catalog/oci/vcn) — creates the virtual cloud network that subnets and security groups belong to
- [OciBlockVolume](/docs/catalog/oci/block-volume) — attaches additional block storage to the instance
