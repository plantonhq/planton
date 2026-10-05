# GcpFilestoreInstance

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this kind: conventions, trade-offs, and what pairs well with it.

GcpFilestoreInstanceSpec defines the configuration for a Google Cloud
Filestore instance.

Filestore provides fully managed, high-performance NFS file storage for
applications that require a file system interface and shared access to data.
Common use cases include media rendering, EDA, genomics processing, web
serving, content management, and shared storage for GKE workloads.

Each instance contains exactly one file share that is mounted via NFS.
Instances are tied to a single VPC network and a single location (zone or
region depending on the tier).

Available tiers (in order of increasing capability):
  - BASIC_HDD / STANDARD: cost-effective HDD-backed storage
  - BASIC_SSD / PREMIUM: mid-tier SSD-backed storage
  - HIGH_SCALE_SSD: legacy high-performance SSD (being replaced by ZONAL)
  - ZONAL: modern single-zone SSD with performance tuning
  - REGIONAL: multi-zone SSD with high availability
  - ENTERPRISE: highest tier with regional HA and cross-region replication

Important behavioral notes:

  - The instance_name, location, tier, protocol, network configuration, and
    kms_key_name are immutable after creation. Changing them requires replacing
    the instance.

  - File share capacity can be increased after creation but not decreased.

  - NFS export options can be updated after creation.

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpFilestoreInstance
metadata:
  name: test-filestore
spec:
  projectId:
    value: "<gcp-project-id>"
  instanceName: my-test-nfs
  location: us-central1-a
  tier: BASIC_SSD
  fileShare:
    name: vol1
    capacityGb: 2560
  networkConfig:
    network:
      value: default
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.instanceName` | `string` |  |  |  |
| `spec.location` | `string` | yes |  |  |
| `spec.tier` | `string` | yes |  |  |
| `spec.description` | `string` |  |  |  |
| `spec.protocol` | `string` |  |  |  |
| `spec.kmsKeyName` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`) |
| `spec.deletionProtectionEnabled` | `bool` |  |  |  |
| `spec.deletionProtectionReason` | `string` |  |  |  |
| `spec.fileShare` | `GcpFilestoreInstanceFileShare` | yes |  |  |
| `spec.fileShare.name` | `string` | yes |  |  |
| `spec.fileShare.capacityGb` | `int32` | yes |  |  |
| `spec.fileShare.nfsExportOptions` | `[]GcpFilestoreInstanceNfsExportOption` |  |  |  |
| `spec.fileShare.nfsExportOptions[].ipRanges` | `[]string` |  |  |  |
| `spec.fileShare.nfsExportOptions[].accessMode` | `string` |  |  |  |
| `spec.fileShare.nfsExportOptions[].squashMode` | `string` |  |  |  |
| `spec.fileShare.nfsExportOptions[].anonUid` | `int32` |  |  |  |
| `spec.fileShare.nfsExportOptions[].anonGid` | `int32` |  |  |  |
| `spec.fileShare.nfsExportOptions[].network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_name`) |
| `spec.fileShare.sourceBackup` | `string` |  |  |  |
| `spec.fileShare.sourceBackupdrBackup` | `string` |  |  |  |
| `spec.networkConfig` | `GcpFilestoreInstanceNetworkConfig` | yes |  |  |
| `spec.networkConfig.network` | `string \| valueFrom` | yes |  | GcpVpcNetwork (`status.outputs.network_name`) |
| `spec.networkConfig.connectMode` | `string` |  |  |  |
| `spec.networkConfig.reservedIpRange` | `string` |  |  |  |
| `spec.networkConfig.modes` | `[]string` |  |  |  |
| `spec.networkConfig.pscEndpointProject` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.performanceConfig` | `GcpFilestoreInstancePerformanceConfig` |  |  |  |
| `spec.performanceConfig.fixedIops` | `GcpFilestoreInstanceFixedIops` |  |  |  |
| `spec.performanceConfig.fixedIops.maxIops` | `int32` | yes |  |  |
| `spec.performanceConfig.iopsPerTb` | `GcpFilestoreInstanceIopsPerTb` |  |  |  |
| `spec.performanceConfig.iopsPerTb.maxIopsPerTb` | `int32` | yes |  |  |
| `spec.initialReplication` | `GcpFilestoreInstanceInitialReplication` |  |  |  |
| `spec.initialReplication.role` | `string` |  |  |  |
| `spec.initialReplication.peerInstances` | `[]string \| valueFrom` | yes |  | GcpFilestoreInstance (`status.outputs.instance_id`) |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.tags` | `map<string, string>` |  |  |  |
| `spec.ldap` | `GcpFilestoreInstanceLdapConfig` |  |  |  |
| `spec.ldap.domain` | `string` | yes |  |  |
| `spec.ldap.servers` | `[]string` | yes |  |  |
| `spec.ldap.groupsOu` | `string` |  |  |  |
| `spec.ldap.usersOu` | `string` |  |  |  |
| `spec.desiredReplicaState` | `string` |  | `READY` |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

The GCP project where the Filestore instance is created.
Can be a literal project ID or a reference to a GcpProject resource.
If omitted, the provider's default project is used.
Immutable after creation.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.instanceName

`string`

Name of the Filestore instance. This becomes the GCP resource name.
Must start with a lowercase letter, contain only lowercase letters, numbers,
and hyphens, and be 2-63 characters long. When omitted, metadata.name
is used. Immutable after creation.

- rule: instance_name must be 2-63 characters of lowercase letters, numbers, and hyphens, starting with a letter and ending with a letter or number

### spec.location

`string` · required

Location where the instance will be deployed.
For BASIC_HDD, BASIC_SSD, STANDARD, PREMIUM, HIGH_SCALE_SSD, and ZONAL
tiers: specify a zone (e.g., "us-central1-a").
For ENTERPRISE and REGIONAL tiers: specify a region (e.g., "us-central1").
Immutable after creation.

- rule: {"required":true}

### spec.tier

`string` · required

Service tier controlling performance, availability, and pricing.
STANDARD / BASIC_HDD: cost-effective HDD-backed (1 TiB minimum).
PREMIUM / BASIC_SSD: mid-tier SSD-backed (2.5 TiB minimum).
HIGH_SCALE_SSD: legacy high-performance SSD (10 TiB minimum).
ZONAL: modern single-zone SSD with IOPS tuning (1 TiB minimum).
REGIONAL: multi-zone SSD with HA (1 TiB minimum).
ENTERPRISE: highest tier, regional HA (1 TiB minimum).
Immutable after creation.

- rule: {"required":true,"string":{"in":["STANDARD","PREMIUM","BASIC_HDD","BASIC_SSD","HIGH_SCALE_SSD","ZONAL","REGIONAL","ENTERPRISE"]}}

### spec.description

`string`

Human-readable description of the instance.

### spec.protocol

`string`

NFS protocol version.
NFS_V3 (default): NFSv3. Broad compatibility, no built-in auth.
NFS_V4_1: NFSv4.1. Supports Kerberos security. Available on
  HIGH_SCALE_SSD, ZONAL, REGIONAL, and ENTERPRISE tiers.
Immutable after creation.

- rule: protocol must be NFS_V3 or NFS_V4_1

### spec.kmsKeyName

`string | valueFrom`

Cloud KMS key for customer-managed encryption at rest (CMEK).
Format: projects/{project}/locations/{location}/keyRings/{keyRing}/cryptoKeys/{key}
If not specified, data is encrypted with Google-managed keys.
Immutable after creation.

- references: GcpKmsKey (`status.outputs.key_id`), GcpKmsKeyHandle (`status.outputs.kms_key`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.deletionProtectionEnabled

`bool`

Whether deletion protection is enabled. When true, the instance cannot
be deleted until this flag is set to false.

### spec.deletionProtectionReason

`string`

Reason for enabling deletion protection. Informational only.

### spec.fileShare

`GcpFilestoreInstanceFileShare` · required

File share configuration. Each Filestore instance has exactly one file share.

- rule: {"required":true}
- rule: source_backup and source_backupdr_backup are mutually exclusive; restore from a Filestore backup or a Backup and DR backup, not both

### spec.fileShare.name

`string` · required

Name of the file share. Becomes the NFS export path.
Must start with a letter, followed by letters, numbers, or underscores.
Maximum 16 characters.
Immutable after creation.

- rule: {"required":true,"string":{"maxLen":"16","pattern":"^[a-zA-Z][a-zA-Z0-9_]{0,15}$"}}

### spec.fileShare.capacityGb

`int32` · required

Capacity of the file share in GiB.
Minimum 1024 GiB (1 TiB) for most tiers. BASIC_SSD/PREMIUM requires
2560 GiB minimum. HIGH_SCALE_SSD requires 10240 GiB minimum.
The GCP API enforces tier-specific minimums.

- rule: {"required":true,"int32":{"gte":1024}}

### spec.fileShare.nfsExportOptions

`[]GcpFilestoreInstanceNfsExportOption`

NFS export options controlling client access to the file share.
Maximum 10 export options per file share.
If empty, all clients are allowed with READ_WRITE access and NO_ROOT_SQUASH.

- rule: {"repeated":{"maxItems":"10"}}

### spec.fileShare.nfsExportOptions[].ipRanges

`[]string`

List of IPv4 addresses or CIDR ranges that are allowed to mount
the file share. If empty, all clients are allowed.
Maximum 64 IP ranges/addresses across all export options per file share.

### spec.fileShare.nfsExportOptions[].accessMode

`string`

Access mode for the export.
READ_WRITE (default): clients can read and write.
READ_ONLY: clients can only read.

- rule: access_mode must be READ_ONLY or READ_WRITE

### spec.fileShare.nfsExportOptions[].squashMode

`string`

Root squash mode for the export.
NO_ROOT_SQUASH (default): root users on clients have root access on the file share.
ROOT_SQUASH: root users on clients are mapped to anon_uid/anon_gid.

- rule: squash_mode must be NO_ROOT_SQUASH or ROOT_SQUASH

### spec.fileShare.nfsExportOptions[].anonUid

`int32` · optional (explicit presence)

Anonymous user ID used when squash_mode is ROOT_SQUASH.
Defaults to 65534 (nobody) if not specified.
Only valid when squash_mode is ROOT_SQUASH.

### spec.fileShare.nfsExportOptions[].anonGid

`int32` · optional (explicit presence)

Anonymous group ID used when squash_mode is ROOT_SQUASH.
Defaults to 65534 (nogroup) if not specified.
Only valid when squash_mode is ROOT_SQUASH.

### spec.fileShare.nfsExportOptions[].network

`string | valueFrom`

Source VPC network for ip_ranges, as the network NAME — a
GcpVpcNetwork reference resolves to it. Required by GCP for
instances using Private Service Connect (where client IPs are not
otherwise attributable to a network), optional for other connect
modes.

- references: GcpVpcNetwork (`status.outputs.network_name`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_name}} -- a bare string does not parse

### spec.fileShare.sourceBackup

`string`

Restore this file share from an existing Filestore backup, in the
format projects/{project}/locations/{location}/backups/{backup}.
The share's capacity must be at least the backup's source capacity.
Create-time only.

### spec.fileShare.sourceBackupdrBackup

`string`

Restore this file share from a Backup and DR Service backup, in
the format projects/{project}/locations/{location}/
backupVaults/{vault}/dataSources/{source}/backups/{backup}.
The vault-based alternative to source_backup (which restores from
Filestore's own backups); set at most one restore source.
Create-time only.

### spec.networkConfig

`GcpFilestoreInstanceNetworkConfig` · required

VPC network configuration. Each Filestore instance connects to exactly one network.

- rule: {"required":true}
- rule: psc_endpoint_project is only meaningful when connect_mode is PRIVATE_SERVICE_CONNECT

### spec.networkConfig.network

`string | valueFrom` · required

VPC network to which the Filestore instance is connected, as the
network NAME (e.g. "prod-vpc") — the Filestore API rejects self-link
URLs for same-project networks, so the reference resolves the
GcpVpcNetwork's plain name output.
Immutable after creation.

- references: GcpVpcNetwork (`status.outputs.network_name`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_name}} -- a bare string does not parse

### spec.networkConfig.connectMode

`string`

Network connection mode.
DIRECT_PEERING (default): VPC peering. Simplest setup.
PRIVATE_SERVICE_ACCESS: uses a private services connection. Required for
  Shared VPC and some enterprise network configurations.
PRIVATE_SERVICE_CONNECT: uses Private Service Connect endpoints.
Immutable after creation.

- rule: connect_mode must be DIRECT_PEERING, PRIVATE_SERVICE_ACCESS, or PRIVATE_SERVICE_CONNECT

### spec.networkConfig.reservedIpRange

`string`

A /29 CIDR block for internal IP addresses reserved for this instance.
Must be unique and non-overlapping with existing subnets in the VPC.
If not specified, GCP automatically selects an unused range.
Immutable after creation.

### spec.networkConfig.modes

`[]string`

IP address versions the instance serves. Values: "MODE_IPV4",
"MODE_IPV6". When empty, ["MODE_IPV4"] is used — the standard NFS
posture. Immutable after creation.

- rule: {"repeated":{"maxItems":"2","unique":true,"items":{"string":{"in":["MODE_IPV4","MODE_IPV6"]}}}}

### spec.networkConfig.pscEndpointProject

`string | valueFrom`

Consumer project in which the Private Service Connect endpoint is
created — a project ID; a GcpProject reference resolves to it. If
omitted, the endpoint is created in the instance's own project.
Only meaningful when connect_mode is PRIVATE_SERVICE_CONNECT
(enforced pre-deploy). Immutable after creation.

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.performanceConfig

`GcpFilestoreInstancePerformanceConfig`

Performance configuration for IOPS tuning.
Available on ZONAL, REGIONAL, and ENTERPRISE tiers.
If not specified, the instance uses the default performance for its tier.

- rule: fixed_iops and iops_per_tb are mutually exclusive; set only one

### spec.performanceConfig.fixedIops

`GcpFilestoreInstanceFixedIops`

Fixed IOPS provisioning. IOPS remains constant regardless of capacity.
Mutually exclusive with iops_per_tb.

### spec.performanceConfig.fixedIops.maxIops

`int32` · required

The number of IOPS to provision. Must be a multiple of 1000.

- rule: {"required":true,"int32":{"gte":1000}}

### spec.performanceConfig.iopsPerTb

`GcpFilestoreInstanceIopsPerTb`

Dynamic IOPS provisioning. IOPS scales with instance capacity.
Mutually exclusive with fixed_iops.

### spec.performanceConfig.iopsPerTb.maxIopsPerTb

`int32` · required

Maximum IOPS per terabyte of capacity.

- rule: {"required":true,"int32":{"gte":1}}

### spec.initialReplication

`GcpFilestoreInstanceInitialReplication`

Cross-instance replication established at create time: this instance
becomes the STANDBY replica of an existing ACTIVE peer (the common
DR posture). Create-time only.

### spec.initialReplication.role

`string`

Replication role of THIS instance:
  ""        -- same as "STANDBY" (GCP default; this instance receives
               replication from the peer)
  "STANDBY" -- this instance is the read-only replica
  "ACTIVE"  -- this instance is the replication source

- rule: replication role must be ACTIVE or STANDBY

### spec.initialReplication.peerInstances

`[]string | valueFrom` · required

Peer Filestore instances in the replication relationship, each as a
reference to a GcpFilestoreInstance (or a literal full resource path
projects/{project}/locations/{location}/instances/{instance}).

- references: GcpFilestoreInstance (`status.outputs.instance_id`)
- rule: {"repeated":{"minItems":"1"}}
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpFilestoreInstance, name: <that resource's name>, fieldPath: status.outputs.instance_id}} -- a bare string does not parse

### spec.labels

`map<string, string>`

User labels merged with Planton attribution labels (which win on key
conflicts). Keys and values must match GCP label constraints.

### spec.tags

`map<string, string>`

Resource Manager tags bound to the instance for org-policy and IAM
conditions. Keys in the form "tagKeys/{id}", values "tagValues/{id}".
Create-time only.

### spec.ldap

`GcpFilestoreInstanceLdapConfig`

LDAP directory integration for NFSv4.1 identity mapping. Requires
protocol NFS_V4_1 — with NFSv3, identity is numeric UID/GID
matching and no directory service applies.

### spec.ldap.domain

`string` · required

LDAP domain name, e.g. "my-domain.com".

- rule: {"required":true}

### spec.ldap.servers

`[]string` · required

LDAP server addresses — either all DNS names (e.g.
"ldap.example.com") or all IP addresses; GCP rejects a mix of the
two formats.

- rule: {"repeated":{"minItems":"1"}}

### spec.ldap.groupsOu

`string`

Groups Organizational Unit (OU) — an optional hint that narrows
LDAP lookups to one OU instead of querying the whole namespace
(faster lookups on large directories).

### spec.ldap.usersOu

`string`

Users Organizational Unit (OU) — the same lookup-narrowing hint
for user entries.

### spec.desiredReplicaState

`string` · optional (explicit presence)

Desired state of THIS instance's replica relationship, when the
instance is the STANDBY side of a replication pair:
  "READY"  (default) -- replication runs; the standby receives
                        changes from the active peer
  "PAUSED"           -- replication is paused (e.g. to freeze the
                        standby at a point in time); resume by
                        setting READY again
Updatable in place; has no effect on an instance without a
replica relationship.

- default: `READY`
- rule: desired_replica_state must be READY or PAUSED

### spec.deletionPolicy

`string`

Deletion policy for the instance — what happens when this resource
is destroyed (evaluated only after deletion_protection_enabled
allows the destroy at all):
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the instance and every file on its share are deleted
  "PREVENT" -- destroy FAILS; a second, independent guard for a
               file server whose data exists nowhere else
  "ABANDON" -- the instance is removed from management but left
               running (and billing) in GCP with its data intact

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `ldap_requires_nfs_v4_1`: ldap requires protocol NFS_V4_1 — LDAP identity mapping is not available on NFSv3

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpFilestoreInstance, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.instance_id` | `string` | Fully qualified resource ID of the Filestore instance. Format: projects/{project}/locations/{location}/instances/{instance} |
| `status.outputs.instance_name` | `string` | Short name of the Filestore instance. |
| `status.outputs.ip_addresses` | `[]string` | IP addresses assigned to the instance on its connected VPC network. Use the first address for NFS mounts: mount <ip>:/<share_name> /mnt/nfs |
| `status.outputs.file_share_name` | `string` | Name of the file share. Used to construct the NFS mount path. |
| `status.outputs.create_time` | `string` | Timestamp when the instance was created (RFC3339 format). |
| `status.outputs.reserved_ip_range` | `string` | The /29 CIDR block reserved for this instance on its VPC network, as resolved by GCP (also populated when the spec left it to GCP to pick). |
| `status.outputs.etag` | `string` | Server-specified ETag guarding against conflicting concurrent updates. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.kmsKeyName` | GcpKmsKey | `status.outputs.key_id` |
| `spec.kmsKeyName` | GcpKmsKeyHandle | `status.outputs.kms_key` |
| `spec.fileShare.nfsExportOptions[].network` | GcpVpcNetwork | `status.outputs.network_name` |
| `spec.networkConfig.network` | GcpVpcNetwork | `status.outputs.network_name` |
| `spec.networkConfig.pscEndpointProject` | GcpProject | `status.outputs.project_id` |
| `spec.initialReplication.peerInstances` | GcpFilestoreInstance | `status.outputs.instance_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| GcpFilestoreInstance | `spec.initialReplication.peerInstances` | `status.outputs.instance_id` |

## See Also

- [Overview](../README.md)
