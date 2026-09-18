# GcpComputeInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpComputeInstanceSpec defines the configuration for a Google Compute
Engine virtual machine — the general-purpose compute node behind
databases-on-VM, stateful application servers, GPU workers, bastion
hosts, and any workload that needs a full OS rather than a managed
runtime.

Important behavioral notes:

  - zone, machine confidential mode, boot-disk source/encryption, NIC
    count and their networks, scratch disks, hostname, and reservation
    affinity are create-time decisions — changing them replaces the VM.
    Many other properties (machine_type, labels, metadata, service
    account, shielded config) are updated by stopping the VM, which the
    provider only does when allow_stopping_for_update is true.
  - Spot VMs (scheduling.provisioning_model = "SPOT") can be preempted
    at any time. Both engines set the API's legacy preemptible flag
    automatically for Spot; automatic restart is forced off.
  - deletion_protection guards against accidental instance deletion but
    does NOT protect data: the boot disk's auto_delete and each attached
    disk's own lifecycle are the data levers. Attached data disks are
    first-class GcpComputeDisk resources referenced here — they survive
    the VM unless their own configuration says otherwise.
  - desired_status stops ("TERMINATED"), suspends ("SUSPENDED"), or
    starts ("RUNNING") the VM without destroying it — compute billing
    stops while terminated; disks keep billing.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpComputeInstance
metadata:
  name: test-vm
spec:
  projectId:
    value: planton-testing
  zone: us-central1-a
  machineType: e2-micro
  bootDisk:
    image: debian-cloud/debian-12
    sizeGb: 10
    type: pd-balanced
    # When declaring guest OS features, list the image's COMPLETE
    # feature set: the API merges the list with the image's own features
    # at create, and the provider then compares the stored (merged) set
    # against the config with replace-on-change semantics — a partial
    # list plans a VM replacement on every re-apply. This is debian-12's
    # full set.
    guestOsFeatures:
      - UEFI_COMPATIBLE
      - VIRTIO_SCSI_MULTIQUEUE
      - GVNIC
      - SEV_CAPABLE
      - SEV_LIVE_MIGRATABLE_V2
  networkInterfaces:
    - network:
        value: default
      accessConfigs:
        - ephemeral: true
          networkTier: PREMIUM
  scheduling:
    provisioningModel: SPOT
    onHostMaintenance: TERMINATE
    instanceTerminationAction: DELETE
  shieldedInstanceConfig:
    enableSecureBoot: true
  # DELETE is the provider default — armed explicitly to prove the
  # round-trip stays diff-free.
  deletionPolicy: DELETE
  labels:
    env: testing
  tags:
    - http-server
  metadata:
    enable-oslogin: "TRUE"
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instanceName` | `string` |  |  |  |
| `spec.zone` | `string` | yes |  |  |
| `spec.machineType` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.hostname` | `string` |  |  |  |
| `spec.bootDisk` | `GcpComputeInstanceBootDisk` | yes |  |  |
| `spec.bootDisk.image` | `string` |  |  |  |
| `spec.bootDisk.sourceSnapshot` | `string` |  |  |  |
| `spec.bootDisk.sourceDisk` | `string \| valueFrom` |  |  | GcpComputeDisk (`status.outputs.self_link`) |
| `spec.bootDisk.sizeGb` | `int32` |  |  |  |
| `spec.bootDisk.type` | `string` |  |  |  |
| `spec.bootDisk.autoDelete` | `bool` |  | `true` |  |
| `spec.bootDisk.deviceName` | `string` |  |  |  |
| `spec.bootDisk.kmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.bootDisk.diskLabels` | `map<string, string>` |  |  |  |
| `spec.bootDisk.provisionedIops` | `int64` |  |  |  |
| `spec.bootDisk.provisionedThroughput` | `int64` |  |  |  |
| `spec.bootDisk.architecture` | `string` |  |  |  |
| `spec.bootDisk.enableConfidentialCompute` | `bool` |  |  |  |
| `spec.bootDisk.resourcePolicies` | `[]string` |  |  |  |
| `spec.bootDisk.storagePool` | `string` |  |  |  |
| `spec.bootDisk.mode` | `string` |  |  |  |
| `spec.bootDisk.interface` | `string` |  |  |  |
| `spec.bootDisk.forceAttach` | `bool` |  |  |  |
| `spec.bootDisk.guestOsFeatures` | `[]string` |  |  |  |
| `spec.bootDisk.replicaZones` | `[]string` |  |  |  |
| `spec.bootDisk.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.bootDisk.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.bootDisk.sourceImageEncryption` | `GcpComputeInstanceSourceEncryption` |  |  |  |
| `spec.bootDisk.sourceImageEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.bootDisk.sourceImageEncryption.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.bootDisk.sourceSnapshotEncryption` | `GcpComputeInstanceSourceEncryption` |  |  |  |
| `spec.bootDisk.sourceSnapshotEncryption.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.bootDisk.sourceSnapshotEncryption.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.attachedDisks` | `[]GcpComputeInstanceAttachedDisk` |  |  |  |
| `spec.attachedDisks[].source` | `string \| valueFrom` | yes |  | GcpComputeDisk (`status.outputs.self_link`) |
| `spec.attachedDisks[].deviceName` | `string` |  |  |  |
| `spec.attachedDisks[].mode` | `string` |  |  |  |
| `spec.attachedDisks[].kmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.attachedDisks[].kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.attachedDisks[].forceAttach` | `bool` |  |  |  |
| `spec.scratchDisks` | `[]GcpComputeInstanceScratchDisk` |  |  |  |
| `spec.scratchDisks[].interface` | `string` | yes |  |  |
| `spec.scratchDisks[].sizeGb` | `int32` |  |  |  |
| `spec.scratchDisks[].deviceName` | `string` |  |  |  |
| `spec.networkInterfaces` | `[]GcpComputeInstanceNetworkInterface` | yes |  |  |
| `spec.networkInterfaces[].network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.networkInterfaces[].subnetwork` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.networkInterfaces[].subnetworkProject` | `string` |  |  |  |
| `spec.networkInterfaces[].networkIp` | `string \| valueFrom` |  |  | GcpAddress (`status.outputs.address`) |
| `spec.networkInterfaces[].accessConfigs` | `[]GcpComputeInstanceAccessConfig` |  |  |  |
| `spec.networkInterfaces[].accessConfigs[].natIp` | `string \| valueFrom` |  |  | GcpAddress (`status.outputs.address`) |
| `spec.networkInterfaces[].accessConfigs[].networkTier` | `string` |  |  |  |
| `spec.networkInterfaces[].accessConfigs[].publicPtrDomainName` | `string` |  |  |  |
| `spec.networkInterfaces[].accessConfigs[].ephemeral` | `bool` |  |  |  |
| `spec.networkInterfaces[].ipv6AccessConfigs` | `[]GcpComputeInstanceIpv6AccessConfig` |  |  |  |
| `spec.networkInterfaces[].ipv6AccessConfigs[].networkTier` | `string` | yes |  |  |
| `spec.networkInterfaces[].ipv6AccessConfigs[].publicPtrDomainName` | `string` |  |  |  |
| `spec.networkInterfaces[].ipv6AccessConfigs[].externalIpv6` | `string` |  |  |  |
| `spec.networkInterfaces[].ipv6AccessConfigs[].externalIpv6PrefixLength` | `string` |  |  |  |
| `spec.networkInterfaces[].ipv6AccessConfigs[].name` | `string` |  |  |  |
| `spec.networkInterfaces[].stackType` | `string` |  |  |  |
| `spec.networkInterfaces[].nicType` | `string` |  |  |  |
| `spec.networkInterfaces[].queueCount` | `int32` |  |  |  |
| `spec.networkInterfaces[].aliasIpRanges` | `[]GcpComputeInstanceAliasIpRange` |  |  |  |
| `spec.networkInterfaces[].aliasIpRanges[].ipCidrRange` | `string` | yes |  |  |
| `spec.networkInterfaces[].aliasIpRanges[].subnetworkRangeName` | `string` |  |  |  |
| `spec.networkInterfaces[].networkAttachment` | `string` |  |  |  |
| `spec.networkInterfaces[].vlan` | `int32` |  |  |  |
| `spec.networkInterfaces[].igmpQuery` | `string` |  |  |  |
| `spec.networkInterfaces[].ipv6Address` | `string` |  |  |  |
| `spec.networkInterfaces[].internalIpv6PrefixLength` | `int32` |  |  |  |
| `spec.serviceAccount` | `GcpComputeInstanceServiceAccount` |  |  |  |
| `spec.serviceAccount.email` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.serviceAccount.scopes` | `[]string` | yes |  |  |
| `spec.scheduling` | `GcpComputeInstanceScheduling` |  |  |  |
| `spec.scheduling.provisioningModel` | `string` |  |  |  |
| `spec.scheduling.automaticRestart` | `bool` |  | `true` |  |
| `spec.scheduling.onHostMaintenance` | `string` |  |  |  |
| `spec.scheduling.instanceTerminationAction` | `string` |  |  |  |
| `spec.scheduling.maxRunDurationSeconds` | `int64` |  |  |  |
| `spec.scheduling.terminationTime` | `string` |  |  |  |
| `spec.scheduling.discardLocalSsdsOnStop` | `bool` |  |  |  |
| `spec.scheduling.availabilityDomain` | `int32` |  |  |  |
| `spec.scheduling.minNodeCpus` | `int32` |  |  |  |
| `spec.scheduling.nodeAffinities` | `[]GcpComputeInstanceNodeAffinity` |  |  |  |
| `spec.scheduling.nodeAffinities[].key` | `string` | yes |  |  |
| `spec.scheduling.nodeAffinities[].operator` | `string` | yes |  |  |
| `spec.scheduling.nodeAffinities[].values` | `[]string` | yes |  |  |
| `spec.scheduling.localSsdRecoveryTimeoutSeconds` | `int64` |  |  |  |
| `spec.scheduling.hostErrorTimeoutSeconds` | `int32` |  |  |  |
| `spec.shieldedInstanceConfig` | `GcpComputeInstanceShieldedConfig` |  |  |  |
| `spec.shieldedInstanceConfig.enableSecureBoot` | `bool` |  |  |  |
| `spec.shieldedInstanceConfig.enableVtpm` | `bool` |  | `true` |  |
| `spec.shieldedInstanceConfig.enableIntegrityMonitoring` | `bool` |  | `true` |  |
| `spec.confidentialInstanceConfig` | `GcpComputeInstanceConfidentialConfig` |  |  |  |
| `spec.confidentialInstanceConfig.confidentialInstanceType` | `string` | yes |  |  |
| `spec.advancedMachineFeatures` | `GcpComputeInstanceAdvancedMachineFeatures` |  |  |  |
| `spec.advancedMachineFeatures.enableNestedVirtualization` | `bool` |  |  |  |
| `spec.advancedMachineFeatures.threadsPerCore` | `int32` |  |  |  |
| `spec.advancedMachineFeatures.visibleCoreCount` | `int32` |  |  |  |
| `spec.advancedMachineFeatures.enableUefiNetworking` | `bool` |  |  |  |
| `spec.advancedMachineFeatures.performanceMonitoringUnit` | `string` |  |  |  |
| `spec.advancedMachineFeatures.turboMode` | `string` |  |  |  |
| `spec.guestAccelerators` | `[]GcpComputeInstanceGuestAccelerator` |  |  |  |
| `spec.guestAccelerators[].type` | `string` | yes |  |  |
| `spec.guestAccelerators[].count` | `int32` | yes |  |  |
| `spec.reservationAffinity` | `GcpComputeInstanceReservationAffinity` |  |  |  |
| `spec.reservationAffinity.type` | `string` | yes |  |  |
| `spec.reservationAffinity.specificReservation` | `GcpComputeInstanceSpecificReservation` |  |  |  |
| `spec.reservationAffinity.specificReservation.key` | `string` | yes |  |  |
| `spec.reservationAffinity.specificReservation.values` | `[]string` | yes |  |  |
| `spec.totalEgressBandwidthTier` | `string` |  |  |  |
| `spec.metadata` | `map<string, string>` |  |  |  |
| `spec.startupScript` | `string` |  |  |  |
| `spec.sshKeys` | `[]string` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.tags` | `[]string` |  |  |  |
| `spec.resourceManagerTags` | `map<string, string>` |  |  |  |
| `spec.resourcePolicies` | `[]string` |  |  |  |
| `spec.minCpuPlatform` | `string` |  |  |  |
| `spec.canIpForward` | `bool` |  |  |  |
| `spec.enableDisplay` | `bool` |  |  |  |
| `spec.deletionProtection` | `bool` |  |  |  |
| `spec.desiredStatus` | `string` |  |  |  |
| `spec.allowStoppingForUpdate` | `bool` |  | `true` |  |
| `spec.keyRevocationActionType` | `string` |  |  |  |
| `spec.instanceEncryptionKey` | `GcpComputeInstanceEncryptionKey` |  |  |  |
| `spec.instanceEncryptionKey.kmsKey` | `string \| valueFrom` | yes |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.instanceEncryptionKey.kmsKeyServiceAccount` | `string` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |
| `spec.workloadIdentityConfig` | `GcpComputeInstanceWorkloadIdentityConfig` |  |  |  |
| `spec.workloadIdentityConfig.identity` | `string` | yes |  |  |
| `spec.workloadIdentityConfig.identityCertificateEnabled` | `bool` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project that owns the instance.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable after creation.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instanceName

`string`

Name of the Compute Engine instance. 1-63 characters, lowercase
letters, numbers, and hyphens; must start with a letter and cannot end
with a hyphen. When omitted, metadata.name is used.
Immutable after creation.

- rule: instance_name must be 1-63 characters of lowercase letters, numbers, and hyphens, starting with a letter and not ending with a hyphen

### spec.zone

`string` · required

Zone where the instance runs, e.g. "us-central1-a".
Immutable after creation (moving zones replaces the VM).

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+-[a-z]$"}}

### spec.machineType

`string` · required

Machine type, e.g. "e2-medium", "n2-standard-4", "c3-highcpu-8", or a
custom shape like "custom-6-20480". Mutable — changing it stops and
restarts the VM (requires allow_stopping_for_update).

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.description

`string`

Human-readable description of the instance.

### spec.hostname

`string`

Custom fully-qualified DNS hostname, e.g. "db-1.prod.internal".
RFC-1035 labels separated by dots; when unset GCP derives
"<name>.c.<project>.internal". Immutable after creation.

- rule: hostname must be a fully qualified DNS name of RFC-1035 labels, e.g. host.example.internal

### spec.bootDisk

`GcpComputeInstanceBootDisk` · required

Boot disk configuration — the disk the OS boots from.

- rule: {"required":true}
- rule: exactly one boot source is required: image (fresh install), source_snapshot (restore), or source_disk (pre-created bootable disk)
- rule: source_image_encryption is only valid together with image
- rule: source_snapshot_encryption is only valid together with source_snapshot
- rule: replica_zones requires a source_snapshot boot source — GCP cannot create a regional boot disk from an image (API 400: "Creating a regional disk from a source image is not supported yet"), and a pre-created source_disk already carries its own zones

### spec.bootDisk.image

`string`

Source image for a fresh boot disk. Accepts an image family
("debian-cloud/debian-12", "ubuntu-os-cloud/ubuntu-2404-lts-amd64") or
a specific image self link. Families resolve to the newest image at
create time. Create-time only.

### spec.bootDisk.sourceSnapshot

`string`

Source snapshot to restore the boot disk from (name or self link).
Create-time only.

### spec.bootDisk.sourceDisk

`string | valueFrom`

Existing bootable disk to boot from, referenced as a GcpComputeDisk.
The disk must live in the instance's zone. When booting from an
existing disk, size/type/encryption below are ignored — the disk
already owns them.

- references: GcpComputeDisk (`status.outputs.self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpComputeDisk, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.bootDisk.sizeGb

`int32`

Size of the boot disk in GB. When omitted, the image or snapshot size
is used. Grows in place; never shrinks. Most OS images need at least
10 GB; the API floor itself is 1 GB.

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","int32":{"lte":65536,"gte":1}}

### spec.bootDisk.type

`string`

Disk type: "pd-standard" (HDD), "pd-balanced" (default in GCP for most
machine shapes and the sensible choice), "pd-ssd" (high IOPS), or a
hyperdisk type on supported machine families ("hyperdisk-balanced").

### spec.bootDisk.autoDelete

`bool` · optional (explicit presence)

Delete the boot disk automatically when the instance is deleted.
Defaults to true (matching GCP). Set false to keep the OS disk for
forensics or re-attachment.

- default: `true`

### spec.bootDisk.deviceName

`string`

Device name exposed under /dev/disk/by-id/google-*. When omitted GCP
assigns one.

### spec.bootDisk.kmsKey

`string | valueFrom`

Customer-managed encryption key (CMEK) for the boot disk, referenced
as a GcpKmsKey. The Compute Engine service agent
(service-<project-number>@compute-system.iam.gserviceaccount.com) must
hold roles/cloudkms.cryptoKeyEncrypterDecrypter on the key.
Create-time only.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.bootDisk.diskLabels

`map<string, string>`

Labels applied to the boot disk itself (distinct from instance
labels) — useful for disk-level cost attribution and snapshot
policies.

### spec.bootDisk.provisionedIops

`int64` · optional (explicit presence)

Provisioned IOPS for hyperdisk types that support tuning
(e.g. hyperdisk-extreme, hyperdisk-balanced). Leave unset for pd-*
types.

- rule: {"int64":{"gt":"0"}}

### spec.bootDisk.provisionedThroughput

`int64` · optional (explicit presence)

Provisioned throughput in MB/s for hyperdisk types that support
tuning (e.g. hyperdisk-throughput, hyperdisk-balanced). Leave unset
for pd-* types.

- rule: {"int64":{"gt":"0"}}

### spec.bootDisk.architecture

`string`

CPU architecture of the disk/image: "X86_64" or "ARM64" (e.g. for
Tau T2A/Axion machine types). Normally inferred from the image.

- rule: architecture must be X86_64 or ARM64

### spec.bootDisk.enableConfidentialCompute

`bool`

Create the boot disk in confidential-compute mode (hyperdisk SKUs
only; requires kms_key).

### spec.bootDisk.resourcePolicies

`[]string`

Self links of resource policies to attach to the boot disk at create
time (e.g. a snapshot schedule). GCP currently allows at most one.
Changing this replaces the instance.

- rule: {"repeated":{"maxItems":"1"}}

### spec.bootDisk.storagePool

`string`

URL of the storage pool to create the boot disk in (hyperdisk storage
pools).

### spec.bootDisk.mode

`string`

Attachment mode: "READ_WRITE" (default) or "READ_ONLY" (share one
boot disk read-only across many VMs).

- rule: mode must be READ_WRITE or READ_ONLY

### spec.bootDisk.interface

`string`

Disk attachment interface: "SCSI" or "NVME". GCP normally selects
the right interface from the machine type and disk type — the
provider's own guidance is "only used for specific cases, please
don't specify this field without advice from Google". Leave unset
unless you have such a case.

- rule: interface must be SCSI or NVME

### spec.bootDisk.forceAttach

`bool`

Force-attach a REGIONAL boot disk even if it is currently attached
to another instance (regional-disk failover takeover). Attempting to
force-attach a zonal disk fails. Changing this replaces the VM.

### spec.bootDisk.guestOsFeatures

`[]string`

Guest OS features to enable on the boot disk, e.g.
["UEFI_COMPATIBLE", "SECURE_BOOT", "GVNIC", "MULTI_IP_SUBNET",
"WINDOWS"]. The accepted set evolves with GCP — see "Enabling guest
operating system features" in the Compute Engine docs. Create-time
only. When set, list the image's COMPLETE feature set, never just
the additions: the API merges this list with the image's own
features at create and the stored disk echoes the merged set, which
the provider then compares authoritatively with replace-on-change
semantics — a partial list plans a VM REPLACEMENT on every re-apply
(live-verified: debian-12's ["UEFI_COMPATIBLE", "GVNIC"] echoed back
["UEFI_COMPATIBLE", "VIRTIO_SCSI_MULTIQUEUE", "GVNIC", "SEV_CAPABLE",
"SEV_LIVE_MIGRATABLE_V2"]). Leave unset to follow the image's own
features cleanly.

### spec.bootDisk.replicaZones

`[]string`

Zones for a REGIONAL boot disk (exactly two, one of which must be
the instance's own zone; short names or self links). Setting this
converts the boot disk to a regional disk replicated across both
zones. Only valid with a source_snapshot boot source (enforced
pre-deploy): GCP rejects creating a regional boot disk from an
image (live-verified API 400: "Creating a regional disk from a
source image is not supported yet" — the snapshot path was
live-verified to produce a true regional boot disk), and a
pre-created source_disk already carries its own zones.
Create-time only.

- rule: replica_zones takes exactly two zones (one must be the instance's zone)

### spec.bootDisk.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the boot disk at create time. Keys in
the form "tagKeys/{id}", values "tagValues/{id}". Create-time only —
changing them replaces the VM. Ignored when booting from an existing
source_disk.

### spec.bootDisk.kmsKeyServiceAccount

`string`

Service account used for the encryption request of kms_key (CMEK).
When omitted, the Compute Engine default service agent is used.
Only meaningful together with kms_key.

### spec.bootDisk.sourceImageEncryption

`GcpComputeInstanceSourceEncryption`

Decrypts the source image when it is itself CMEK-encrypted. Only
valid together with image.

### spec.bootDisk.sourceImageEncryption.kmsKey

`string | valueFrom` · required

The KMS key the source was encrypted with, referenced as a GcpKmsKey
or a literal self link. The service agent performing the read needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.bootDisk.sourceImageEncryption.kmsKeyServiceAccount

`string`

Service account used for the decryption request. When omitted, the
Compute Engine default service agent is used.

### spec.bootDisk.sourceSnapshotEncryption

`GcpComputeInstanceSourceEncryption`

Decrypts the source snapshot when it is itself CMEK-encrypted. Only
valid together with source_snapshot.

### spec.bootDisk.sourceSnapshotEncryption.kmsKey

`string | valueFrom` · required

The KMS key the source was encrypted with, referenced as a GcpKmsKey
or a literal self link. The service agent performing the read needs
roles/cloudkms.cryptoKeyEncrypterDecrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.bootDisk.sourceSnapshotEncryption.kmsKeyServiceAccount

`string`

Service account used for the decryption request. When omitted, the
Compute Engine default service agent is used.

### spec.attachedDisks

`[]GcpComputeInstanceAttachedDisk`

Additional persistent data disks attached to the instance. Each disk
is a first-class GcpComputeDisk resource referenced by self link — the
disk has its own lifecycle and survives this VM unless its own
configuration says otherwise.

### spec.attachedDisks[].source

`string | valueFrom` · required

The disk to attach, referenced as a GcpComputeDisk (or a literal disk
name/self link). The disk must live in the instance's zone.

- references: GcpComputeDisk (`status.outputs.self_link`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpComputeDisk, name: <that resource's name>, fieldPath: status.outputs.self_link}} -- a bare string does not parse

### spec.attachedDisks[].deviceName

`string`

Device name exposed under /dev/disk/by-id/google-*. When omitted GCP
assigns "persistent-disk-N".

### spec.attachedDisks[].mode

`string`

Attachment mode: "READ_WRITE" (default) or "READ_ONLY" (lets many VMs
attach the same disk simultaneously).

- rule: mode must be READ_WRITE or READ_ONLY

### spec.attachedDisks[].kmsKey

`string | valueFrom`

The CMEK key protecting the attached disk, referenced as a GcpKmsKey.
Required only when the disk is CMEK-encrypted — the attachment must
present the same key the disk was created with.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.attachedDisks[].kmsKeyServiceAccount

`string`

Service account used for the encryption request of kms_key (CMEK).
When omitted, the Compute Engine default service agent is used.
Only meaningful together with kms_key.

### spec.attachedDisks[].forceAttach

`bool`

Force-attach a REGIONAL disk even if it is currently attached to
another instance (regional-disk failover takeover). Attempting to
force-attach a zonal disk fails. Changing this replaces the VM.

### spec.scratchDisks

`[]GcpComputeInstanceScratchDisk`

Ephemeral local-SSD scratch disks physically attached to the host.
Contents are lost when the VM stops or is preempted — use only for
caches, temp space, and high-IOPS scratch data. Create-time only.

### spec.scratchDisks[].interface

`string` · required

Disk interface: "NVME" (recommended; highest performance) or "SCSI".

- rule: {"required":true,"string":{"in":["NVME","SCSI"]}}

### spec.scratchDisks[].sizeGb

`int32`

Size in GB. Local SSDs come in fixed 375 GB units (or 3000 GB on
supported Z3 shapes). When omitted, 375 is used.

- rule: scratch disk size_gb must be 375 (standard local SSD unit) or 3000 (Z3 shapes)

### spec.scratchDisks[].deviceName

`string`

Device name exposed under /dev/disk/by-id/google-*.

### spec.networkInterfaces

`[]GcpComputeInstanceNetworkInterface` · required

Network interfaces. At least one is required; multiple NICs must each
attach to a different VPC network. NIC count and their target
networks are immutable after creation.

- rule: {"required":true,"repeated":{"minItems":"1"}}
- rule: each network interface needs an attachment point: a network (auto-mode VPC), a subnetwork (custom-mode VPC), or a network_attachment (Private Service Connect) — set at least one

### spec.networkInterfaces[].network

`string | valueFrom`

VPC network for this interface, referenced as a GcpVpcNetwork.
Sufficient alone only for auto-mode VPCs; custom-mode VPCs need
subnetwork.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.networkInterfaces[].subnetwork

`string | valueFrom`

Subnetwork for this interface, referenced as a GcpSubnetwork. The
subnetwork's region must contain the instance's zone.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.networkInterfaces[].subnetworkProject

`string`

Project owning the subnetwork — set when attaching to a Shared VPC
host project's subnetwork from a service project.

### spec.networkInterfaces[].networkIp

`string | valueFrom`

Static internal IP for this interface. Accepts a literal IP or a
reference to a reserved INTERNAL GcpAddress. When omitted, GCP
assigns an ephemeral internal IP from the subnetwork range.

- references: GcpAddress (`status.outputs.address`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.networkInterfaces[].accessConfigs

`[]GcpComputeInstanceAccessConfig`

External IPv4 access configs. Empty means no external IP (private
VM — pair with Cloud NAT for egress). GCP supports at most one
access config per interface.

- rule: {"repeated":{"maxItems":"1"}}
- rule: an access config grants either a static nat_ip or an ephemeral address — set nat_ip, or set ephemeral: true

### spec.networkInterfaces[].accessConfigs[].natIp

`string | valueFrom`

Static external IP, as a literal or a reference to a reserved
EXTERNAL GcpAddress. Alternative to ephemeral: true.

- references: GcpAddress (`status.outputs.address`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.networkInterfaces[].accessConfigs[].networkTier

`string`

Network service tier for this IP: "PREMIUM" (default; Google's global
backbone) or "STANDARD" (regional, cheaper).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["PREMIUM","STANDARD"]}}

### spec.networkInterfaces[].accessConfigs[].publicPtrDomainName

`string`

Domain name for the public PTR (reverse DNS) record of this IP.

### spec.networkInterfaces[].accessConfigs[].ephemeral

`bool` · optional (explicit presence)

Let GCP assign an ephemeral external IPv4 to this interface. Alternative
to a static nat_ip. The manifest's word for it -- never reaches GCP: an
access config without a nat_ip IS the ephemeral request on the wire.

### spec.networkInterfaces[].ipv6AccessConfigs

`[]GcpComputeInstanceIpv6AccessConfig`

External IPv6 access configs. Requires stack_type "IPV4_IPV6" and a
subnetwork with an external IPv6 range. At most one per interface.

- rule: {"repeated":{"maxItems":"1"}}

### spec.networkInterfaces[].ipv6AccessConfigs[].networkTier

`string` · required

Network service tier for IPv6 traffic. Only "PREMIUM" is valid.

- rule: {"required":true,"string":{"in":["PREMIUM"]}}

### spec.networkInterfaces[].ipv6AccessConfigs[].publicPtrDomainName

`string`

Domain name for the public PTR (reverse DNS) record of the external
IPv6 range.

### spec.networkInterfaces[].ipv6AccessConfigs[].externalIpv6

`string`

Static EXTERNAL IPv6 address (the first address of the external
range) to pin to this interface. Must be unused and in the same
region as the instance's zone. When omitted, GCP assigns an external
IPv6 range from the subnetwork. Changing it replaces the VM.

### spec.networkInterfaces[].ipv6AccessConfigs[].externalIpv6PrefixLength

`string`

Prefix length of the external IPv6 range (the provider models this
as a string). Normally read back from GCP rather than set — only
meaningful together with external_ipv6. Changing it replaces the VM.

### spec.networkInterfaces[].ipv6AccessConfigs[].name

`string`

Name of this IPv6 access configuration; GCP's recommended value is
"External IPv6". When omitted, GCP assigns one. Changing it replaces
the VM.

### spec.networkInterfaces[].stackType

`string`

IP stack of the interface:
  ""            -- same as "IPV4_ONLY" (GCP default)
  "IPV4_ONLY"   -- IPv4 addresses only
  "IPV4_IPV6"   -- dual stack (subnetwork must have an IPv6 range)
  "IPV6_ONLY"   -- IPv6 only (supported on IPv6-enabled subnetworks)

- rule: stack_type must be IPV4_ONLY, IPV4_IPV6, or IPV6_ONLY

### spec.networkInterfaces[].nicType

`string`

vNIC type: "" (GCP picks), "GVNIC" (recommended on modern machine
families; required for TIER_1 bandwidth), "VIRTIO_NET" (legacy), or
the RDMA types "IDPF", "MRDMA", "IRDMA" on specialized shapes.

- rule: nic_type must be one of GVNIC, VIRTIO_NET, IDPF, MRDMA, IRDMA

### spec.networkInterfaces[].queueCount

`int32` · optional (explicit presence)

Networking queue count for Rx and Tx (1-32). When omitted GCP sizes
queues from vCPU count.

- rule: {"int32":{"lte":32,"gte":1}}

### spec.networkInterfaces[].aliasIpRanges

`[]GcpComputeInstanceAliasIpRange`

Alias IP ranges served by this interface — the mechanism behind
per-pod/per-container IPs and multi-IP VMs.

### spec.networkInterfaces[].aliasIpRanges[].ipCidrRange

`string` · required

The alias range: a CIDR ("10.1.2.0/24"), a single IP ("10.1.2.3"), or
a netmask ("/24") to auto-allocate from the range.

- rule: {"required":true}

### spec.networkInterfaces[].aliasIpRanges[].subnetworkRangeName

`string`

Secondary range name on the subnetwork to allocate from. When omitted
the primary range is used.

### spec.networkInterfaces[].networkAttachment

`string`

URL of a Private Service Connect NETWORK ATTACHMENT this interface
connects to, in the form
"projects/{projectNumber}/regions/{region}/networkAttachments/{name}"
— the consumer side of PSC interfaces, connecting this VM into a
producer's VPC. An attachment-only interface is legal (no network or
subnetwork). Immutable after creation.

### spec.networkInterfaces[].vlan

`int32` · optional (explicit presence)

VLAN tag (2-255) making this a DYNAMIC network interface — a
sub-interface multiplexed onto a parent NIC. Immutable after
creation.

- rule: {"int32":{"lte":255,"gte":2}}

### spec.networkInterfaces[].igmpQuery

`string`

IGMP multicast query support on this interface:
  ""                    -- GCP default (disabled)
  "IGMP_QUERY_V2"       -- IGMPv2 queries enabled (multicast)
  "IGMP_QUERY_DISABLED" -- explicitly disabled
Updatable in place.

- rule: igmp_query must be IGMP_QUERY_V2 or IGMP_QUERY_DISABLED

### spec.networkInterfaces[].ipv6Address

`string`

Static INTERNAL IPv6 address for this interface (requires an
IPv6-enabled stack_type and subnetwork). When omitted, GCP assigns
one from the subnetwork's internal IPv6 range. Long-form and
compressed spellings are equivalent.

### spec.networkInterfaces[].internalIpv6PrefixLength

`int32` · optional (explicit presence)

Prefix length of the primary internal IPv6 range assigned to this
interface. When omitted, GCP assigns its default.

- rule: {"int32":{"lte":128,"gte":1}}

### spec.serviceAccount

`GcpComputeInstanceServiceAccount`

Service account the VM's workloads authenticate as. When omitted, the
Compute Engine default service account is used with its default
scopes — prefer a dedicated least-privilege account for production.

### spec.serviceAccount.email

`string | valueFrom`

Service account email, referenced as a GcpServiceAccount or a literal
email. Changing it stops and restarts the VM (requires
allow_stopping_for_update).

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.serviceAccount.scopes

`[]string` · required

OAuth scopes for the attached account. The modern practice is a
single "https://www.googleapis.com/auth/cloud-platform" scope with
access controlled entirely by IAM roles; narrower legacy scopes
remain supported. Required when this block is set.

- rule: {"repeated":{"minItems":"1"}}

### spec.scheduling

`GcpComputeInstanceScheduling`

Scheduling policy: Spot vs standard provisioning, maintenance
behavior, run-duration limits, and sole-tenant node placement.

### spec.scheduling.provisioningModel

`string`

Provisioning model:
  ""                  -- same as "STANDARD"
  "STANDARD"          -- on-demand capacity
  "SPOT"              -- deeply discounted preemptible capacity; GCP
                         may reclaim the VM at any time
  "FLEX_START"        -- discounted capacity with a flexible start
                         time (Dynamic Workload Scheduler); requires
                         max_run_duration_seconds and is deleted when
                         reclaimed
  "RESERVATION_BOUND" -- runs only on capacity from one specific
                         reservation (pair with reservation_affinity
                         type SPECIFIC_RESERVATION)
Create-time only.

- rule: provisioning_model must be STANDARD, SPOT, FLEX_START, or RESERVATION_BOUND

### spec.scheduling.automaticRestart

`bool` · optional (explicit presence)

Restart the VM automatically when Compute Engine (not a user) stops
it. Defaults to true for standard VMs; must be false (or unset) for
Spot.

- default: `true`

### spec.scheduling.onHostMaintenance

`string`

Host maintenance behavior: "" (GCP default "MIGRATE"), "MIGRATE"
(live-migrate; zero downtime), or "TERMINATE" (stop during
maintenance — required for GPUs and confidential VMs).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["MIGRATE","TERMINATE"]}}

### spec.scheduling.instanceTerminationAction

`string`

What GCP does when the VM is reclaimed (Spot preemption, FLEX_START
expiry) or a run-duration limit fires: "STOP" (keep the stopped VM
and disks) or "DELETE" (remove the VM). Applies to SPOT and
FLEX_START models and to timed-run VMs
(max_run_duration_seconds/termination_time).

- rule: {"ignore":"IGNORE_IF_ZERO_VALUE","string":{"in":["STOP","DELETE"]}}

### spec.scheduling.maxRunDurationSeconds

`int64` · optional (explicit presence)

Maximum run duration in seconds, after which
instance_termination_action is executed. The duration clock starts at
every VM start. Mutually exclusive with termination_time.

- rule: {"int64":{"lte":"315576000000","gte":"1"}}

### spec.scheduling.terminationTime

`string`

Absolute timestamp (RFC 3339) at which the VM is terminated. Mutually
exclusive with max_run_duration_seconds.

### spec.scheduling.discardLocalSsdsOnStop

`bool` · optional (explicit presence)

Discard local-SSD contents when the VM is stopped by a lifetime limit
(max_run_duration/termination_time) instead of preserving them.

### spec.scheduling.availabilityDomain

`int32` · optional (explicit presence)

Availability domain for spread-placement within the zone (used with
spread placement policies). 0 means unset.

- rule: {"int32":{"gte":1}}

### spec.scheduling.minNodeCpus

`int32` · optional (explicit presence)

Minimum vCPUs on the sole-tenant node this VM can be scheduled onto.
Sole-tenancy only.

- rule: {"int32":{"gte":1}}

### spec.scheduling.nodeAffinities

`[]GcpComputeInstanceNodeAffinity`

Sole-tenant node affinities selecting which node groups this VM may
run on. Setting any affinity places the VM on sole-tenant hardware.

### spec.scheduling.nodeAffinities[].key

`string` · required

Node-group label key to match (e.g.
"compute.googleapis.com/node-group-name").

- rule: {"required":true}

### spec.scheduling.nodeAffinities[].operator

`string` · required

Match operator: "IN" or "NOT_IN".

- rule: {"required":true,"string":{"in":["IN","NOT_IN"]}}

### spec.scheduling.nodeAffinities[].values

`[]string` · required

Label values to match.

- rule: {"repeated":{"minItems":"1"}}

### spec.scheduling.localSsdRecoveryTimeoutSeconds

`int64` · optional (explicit presence)

How long Compute Engine waits for a local-SSD-preserving recovery
when the host fails, in seconds, before falling back to default
recovery.

- rule: {"int64":{"lte":"604800","gte":"0"}}

### spec.scheduling.hostErrorTimeoutSeconds

`int32` · optional (explicit presence)

How long Compute Engine waits, in seconds, before declaring the host
failed and starting host-error recovery (restart or termination per
automatic_restart). A lower value recovers faster from a hung host at
the cost of more false positives; leave unset for Compute Engine's
default recovery timing. Must be 90..330 in steps of 30 (90, 120, ...,
330).

- rule: host_error_timeout_seconds must be a multiple of 30 between 90 and 330
- rule: {"int32":{"lte":330,"gte":90}}

### spec.shieldedInstanceConfig

`GcpComputeInstanceShieldedConfig`

Shielded VM configuration (secure boot, vTPM, integrity monitoring).
Requires an image with Shielded VM support (all recent Google-provided
images qualify).

### spec.shieldedInstanceConfig.enableSecureBoot

`bool` · optional (explicit presence)

Verify the boot loader's signature chain; blocks boot on tampering.
GCP default is false because some third-party images are unsigned.

### spec.shieldedInstanceConfig.enableVtpm

`bool` · optional (explicit presence)

Virtual Trusted Platform Module. GCP default is true.

- default: `true`

### spec.shieldedInstanceConfig.enableIntegrityMonitoring

`bool` · optional (explicit presence)

Boot-integrity measurement and monitoring via the vTPM. GCP default
is true.

- default: `true`

### spec.confidentialInstanceConfig

`GcpComputeInstanceConfidentialConfig`

Confidential VM configuration — hardware memory encryption (AMD SEV /
SEV-SNP or Intel TDX). Requires a supported machine family (e.g. N2D,
C2D, C3) and on_host_maintenance = "TERMINATE". Create-time only.

### spec.confidentialInstanceConfig.confidentialInstanceType

`string` · required

Confidential computing technology:
  "SEV"     -- AMD Secure Encrypted Virtualization (N2D/C2D/C3D)
  "SEV_SNP" -- AMD SEV Secure Nested Paging (requires an AMD Milan+
               min_cpu_platform)
  "TDX"     -- Intel Trust Domain Extensions (C3)

- rule: {"required":true,"string":{"in":["SEV","SEV_SNP","TDX"]}}

### spec.advancedMachineFeatures

`GcpComputeInstanceAdvancedMachineFeatures`

Advanced machine features: nested virtualization, SMT control, visible
core count, UEFI networking, performance monitoring unit, and turbo
mode.

### spec.advancedMachineFeatures.enableNestedVirtualization

`bool` · optional (explicit presence)

Expose nested virtualization support (VMX) to the guest — run VMs
inside this VM.

### spec.advancedMachineFeatures.threadsPerCore

`int32` · optional (explicit presence)

Threads per physical core: 1 disables simultaneous multithreading
(SMT) — common for licensing and security isolation; 2 is the
hardware default.

- rule: {"int32":{"in":[1,2]}}

### spec.advancedMachineFeatures.visibleCoreCount

`int32` · optional (explicit presence)

Number of physical cores exposed to the guest (core visibility for
per-core licensing). When unset all cores are visible.

- rule: {"int32":{"gte":1}}

### spec.advancedMachineFeatures.enableUefiNetworking

`bool` · optional (explicit presence)

Enable UEFI networking in the guest firmware. Create-time only.

### spec.advancedMachineFeatures.performanceMonitoringUnit

`string`

Performance monitoring unit exposure level: "STANDARD", "ENHANCED",
or "ARCHITECTURAL".

- rule: performance_monitoring_unit must be STANDARD, ENHANCED, or ARCHITECTURAL

### spec.advancedMachineFeatures.turboMode

`string`

Turbo frequency mode. "ALL_CORE_MAX" runs all cores at maximum turbo
frequency (supported machine families only).

- rule: turbo_mode must be ALL_CORE_MAX

### spec.guestAccelerators

`[]GcpComputeInstanceGuestAccelerator`

GPU accelerator cards attached to the instance. Requires a
GPU-capable zone and on_host_maintenance = "TERMINATE". Each entry
attaches at least one card (count >= 1). Removing every entry from an
existing VM leaves its GPUs attached (the provider preserves the
current accelerators when the block is absent); detaching GPUs
replaces the VM, so plan it as a recreate rather than an edit.

### spec.guestAccelerators[].type

`string` · required

Accelerator type available in the instance's zone, e.g.
"nvidia-tesla-t4", "nvidia-l4", "nvidia-a100-80gb".

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.guestAccelerators[].count

`int32` · required

Number of cards of this type.

- rule: {"required":true,"int32":{"gte":1}}

### spec.reservationAffinity

`GcpComputeInstanceReservationAffinity`

Reservation affinity — whether this VM consumes capacity from any
matching reservation, a specific reservation, or none.

- rule: specific_reservation must be set when (and only when) type is SPECIFIC_RESERVATION

### spec.reservationAffinity.type

`string` · required

Reservation consumption mode:
  "ANY_RESERVATION"      -- consume any matching reservation (GCP
                            default)
  "SPECIFIC_RESERVATION" -- consume only the named reservation
  "NO_RESERVATION"       -- never consume reserved capacity

- rule: {"required":true,"string":{"in":["ANY_RESERVATION","SPECIFIC_RESERVATION","NO_RESERVATION"]}}

### spec.reservationAffinity.specificReservation

`GcpComputeInstanceSpecificReservation`

The specific reservation to consume (type SPECIFIC_RESERVATION).

### spec.reservationAffinity.specificReservation.key

`string` · required

Reservation label key — use
"compute.googleapis.com/reservation-name" to target by name.

- rule: {"required":true}

### spec.reservationAffinity.specificReservation.values

`[]string` · required

Reservation label values (the reservation name when using the
reservation-name key).

- rule: {"repeated":{"minItems":"1"}}

### spec.totalEgressBandwidthTier

`string`

Per-VM egress bandwidth tier. "TIER_1" raises the bandwidth cap on
supported machine shapes (N2/N2D/C2/C3 with >= 30 vCPUs and gVNIC);
"DEFAULT" is the standard cap.

- rule: total_egress_bandwidth_tier must be DEFAULT or TIER_1

### spec.metadata

`map<string, string>`

Custom metadata key/value pairs made available to the guest OS via the
metadata server. Well-known keys configure agents and features (e.g.
"enable-oslogin", "startup-script-url").

### spec.startupScript

`string`

Startup script executed by the guest agent on every boot. Maps to the
metadata_startup_script surface, which keeps it distinct from user
metadata and re-runs it on each start.

### spec.sshKeys

`[]string`

SSH public keys in "username:ssh-rsa AAAA... user" format. Folded into
the instance metadata "ssh-keys" key (newline-joined) identically on
both engines. Ignored by VMs using OS Login.

### spec.labels

`map<string, string>`

User labels merged with Planton attribution labels (which win on key
conflicts). Keys and values must match GCP label constraints:
lowercase letters, numbers, hyphens, underscores; keys start with a
letter, max 63 characters.

### spec.tags

`[]string`

Network tags used by firewall rules and network routes to select this
instance.

### spec.resourceManagerTags

`map<string, string>`

Resource Manager tags bound to the instance for org-policy and IAM
conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
Create-time only.

### spec.resourcePolicies

`[]string`

Self links of compute resource policies to attach to the instance
(e.g. an instance schedule that starts/stops the VM on a calendar).
GCP currently allows at most one policy per instance.

- rule: {"repeated":{"maxItems":"1"}}

### spec.minCpuPlatform

`string`

Minimum CPU platform for the VM, e.g. "Intel Ice Lake" or
"AMD Milan". Constrains scheduling to hosts with at least this
platform.

### spec.canIpForward

`bool`

Allow sending/receiving packets with source or destination IPs that do
not match the instance's own addresses — required for VMs acting as
routers, NAT gateways, or VPN endpoints. Create-time only.

### spec.enableDisplay

`bool`

Enable the virtual display device (needed by some remote-desktop and
screen-capture tooling on headless VMs).

### spec.deletionProtection

`bool`

Protect the instance against accidental deletion. Deleting a protected
instance fails until this is set back to false. Defaults to false: a
VM is a compute node whose data levers are its disks — the boot disk's
auto_delete and each GcpComputeDisk's own lifecycle guard the data.

### spec.desiredStatus

`string`

Desired lifecycle status of the VM:
  ""           -- same as "RUNNING"
  "RUNNING"    -- started
  "SUSPENDED"  -- suspended to disk (fast resume; memory persisted)
  "TERMINATED" -- stopped (compute billing stops; disks keep billing)
Changing this starts/suspends/stops the VM in place.

- rule: desired_status must be RUNNING, SUSPENDED, or TERMINATED

### spec.allowStoppingForUpdate

`bool` · optional (explicit presence)

Allow the provider to stop and restart the VM when an update requires
it (machine type, service account, network interface changes, ...).
Without it those updates fail instead of causing downtime. Recommended
true for instances whose brief restart is acceptable.

- default: `true`

### spec.keyRevocationActionType

`string`

Action GCP takes on the VM when a Cloud KMS key protecting it is
revoked: "NONE" (default) or "STOP".

- rule: key_revocation_action_type must be NONE or STOP

### spec.instanceEncryptionKey

`GcpComputeInstanceEncryptionKey`

Customer-managed encryption key (CMEK) for INSTANCE-LEVEL data —
memory contents and other instance state, distinct from the per-disk
keys on boot_disk/attached_disks. Create-time only.

### spec.instanceEncryptionKey.kmsKey

`string | valueFrom` · required

The KMS key encrypting instance-level data, referenced as a
GcpKmsKey or a literal self link. The Compute Engine service agent
needs roles/cloudkms.cryptoKeyEncrypterDecrypter on it.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.instanceEncryptionKey.kmsKeyServiceAccount

`string`

Service account used for the encryption request. When omitted, the
Compute Engine default service agent is used.

### spec.deletionPolicy

`string`

Deletion policy — what happens when this resource is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the instance is deleted (disks follow their own
               lifecycle: boot auto_delete, GcpComputeDisk configs)
  "PREVENT" -- destroy FAILS; a guard rail beyond deletion_protection
               because it blocks the IaC destroy itself
  "ABANDON" -- the instance is removed from management but left
               running in GCP

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

### spec.workloadIdentityConfig

`GcpComputeInstanceWorkloadIdentityConfig`

Managed workload identity for the VM: a SPIFFE identity issued to the
instance (and, optionally, X.509 identity certificates) so workloads
on it authenticate to each other by identity instead of shared
secrets or network position. Create-time only: both fields are
immutable, so changing them replaces the VM.

### spec.workloadIdentityConfig.identity

`string` · required

The SPIFFE ID Compute Engine issues to the instance, e.g.
"spiffe://PROJECT.svc.id.goog/ns/NAMESPACE/sa/SERVICE_ACCOUNT" or a
workload-identity-pool identity of the form
"spiffe://POOL.global.PROJECT_NUMBER.workload.id.goog/ns/NS/sa/SA".
Immutable.

- rule: {"required":true,"string":{"minLen":"1","prefix":"spiffe://"}}

### spec.workloadIdentityConfig.identityCertificateEnabled

`bool`

Whether Compute Engine also issues and rotates X.509 certificates
bound to the identity, made available on the VM for mutual TLS.
Immutable.

## Validation Rules

- `reclaimable_models_require_no_automatic_restart`: SPOT and FLEX_START VMs cannot automatically restart after reclamation — leave scheduling.automatic_restart unset or set it to false
- `termination_action_requires_reclaimable_or_timed_vm`: scheduling.instance_termination_action applies to SPOT/FLEX_START VMs and to timed-run VMs (max_run_duration_seconds or termination_time) — configure one of those or remove the termination action
- `flex_start_requires_max_run_duration`: FLEX_START VMs must bound their runtime — set scheduling.max_run_duration_seconds
- `flex_start_terminates_by_delete`: FLEX_START VMs are deleted when reclaimed — leave scheduling.instance_termination_action unset or set it to DELETE
- `reservation_bound_requires_specific_reservation`: RESERVATION_BOUND VMs consume one named reservation — set reservation_affinity.type to SPECIFIC_RESERVATION and name the reservation
- `confidential_requires_terminate_maintenance`: confidential VMs cannot live-migrate — set scheduling.on_host_maintenance to TERMINATE when confidential_instance_config is configured
- `accelerators_require_terminate_maintenance`: VMs with guest accelerators (GPUs) cannot live-migrate — set scheduling.on_host_maintenance to TERMINATE when guest_accelerators are attached
- `max_run_duration_conflicts_with_termination_time`: max_run_duration_seconds and termination_time both bound the VM's lifetime — set at most one

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpComputeInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.instance_name` | `string` | Name of the Compute Engine instance. |
| `status.outputs.instance_id` | `string` | Instance ID (unique numeric identifier). |
| `status.outputs.self_link` | `string` | Self link URL of the instance. |
| `status.outputs.internal_ip` | `string` | Internal (private) IP address of the instance. |
| `status.outputs.external_ip` | `string` | External (public) IP address of the instance (if configured). |
| `status.outputs.status` | `string` | Current status of the instance (RUNNING, STOPPED, etc.). |
| `status.outputs.zone` | `string` | Zone where the instance is located. |
| `status.outputs.machine_type` | `string` | Machine type of the instance. |
| `status.outputs.cpu_platform` | `string` | CPU platform of the instance. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.bootDisk.sourceDisk` | GcpComputeDisk | `status.outputs.self_link` |
| `spec.bootDisk.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.bootDisk.sourceImageEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.bootDisk.sourceSnapshotEncryption.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.attachedDisks[].source` | GcpComputeDisk | `status.outputs.self_link` |
| `spec.attachedDisks[].kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.networkInterfaces[].network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.networkInterfaces[].subnetwork` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.networkInterfaces[].networkIp` | GcpAddress | `status.outputs.address` |
| `spec.networkInterfaces[].accessConfigs[].natIp` | GcpAddress | `status.outputs.address` |
| `spec.serviceAccount.email` | GcpServiceAccount | `status.outputs.email` |
| `spec.instanceEncryptionKey.kmsKey` | GcpKmsKey | `status.outputs.key_id` |

## See Also

- [Overview](../README.md)
