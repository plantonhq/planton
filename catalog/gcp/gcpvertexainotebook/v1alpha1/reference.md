# GcpVertexAiNotebook

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `gcp.planton.dev/v1alpha1`

**Guide**: [GUIDE.md](../GUIDE.md) -- authored operational judgment for this component: conventions, trade-offs, and what pairs well with it.

GcpVertexAiNotebookSpec defines the configuration for a GCP Vertex AI
Workbench instance (managed notebook).

Vertex AI Workbench provides managed JupyterLab notebook instances
for data science and machine learning workflows. Each instance is a
Compute Engine VM pre-configured with JupyterLab, ML frameworks, and
GPU support. Users access notebooks through a secure proxy URL.

Important behavioral notes:

  - The instance_name and location fields are immutable after creation.
    Changing them requires destroying and recreating the instance.

  - The location field specifies a zone (e.g., "us-central1-a"), not
    a region. Notebooks run on individual VMs in specific zones.

  - Exactly one of vm_image or container_image may be specified. If
    neither is specified, GCP uses the default deep learning VM image.

  - GPU accelerators require compatible machine types (e.g., n1-standard-*
    for NVIDIA_TESLA_T4). See GCP documentation for compatibility matrix.

  - Network configuration, service account, tags, confidential computing,
    reservation affinity, and disk encryption settings are immutable
    after creation (ForceNew).

## Example

```yaml
apiVersion: gcp.planton.dev/v1alpha1
kind: GcpVertexAiNotebook
metadata:
  name: test-ml-notebook
spec:
  projectId:
    value: my-gcp-project
  location: us-central1-a
  machineType: n1-standard-4
  instanceName: test-ml-notebook
  labels:
    team: ml-platform
    cost-center: research
  instanceOwners:
    - data-scientist@my-gcp-project.iam.gserviceaccount.com
  bootDisk:
    diskType: PD_SSD
    diskSizeGb: 200
  dataDisk:
    diskType: PD_BALANCED
    diskSizeGb: 500
  acceleratorConfig:
    type: NVIDIA_TESLA_T4
    coreCount: 1
  networkInterface:
    network:
      value: projects/my-gcp-project/global/networks/default
    subnet:
      value: projects/my-gcp-project/regions/us-central1/subnetworks/default
  disablePublicIp: true
  serviceAccount:
    value: notebook-sa@my-gcp-project.iam.gserviceaccount.com
  tags:
    - notebook
    - ml-team
  vmImage:
    project: cloud-notebooks-managed
    family: workbench-instances
  shieldedInstanceConfig:
    enableSecureBoot: true
    enableVtpm: true
    enableIntegrityMonitoring: true
  deletionPolicy: DELETE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.projectId` | `string \| valueFrom` |  |  | GcpProject (`status.outputs.project_id`) |
| `spec.location` | `string` | yes |  |  |
| `spec.machineType` | `string` | yes |  |  |
| `spec.instanceName` | `string` |  |  |  |
| `spec.instanceOwners` | `[]string` |  |  |  |
| `spec.desiredState` | `string` |  |  |  |
| `spec.disableProxyAccess` | `bool` |  |  |  |
| `spec.metadata` | `map<string, string>` |  |  |  |
| `spec.labels` | `map<string, string>` |  |  |  |
| `spec.bootDisk` | `GcpVertexAiNotebookBootDisk` |  |  |  |
| `spec.bootDisk.diskType` | `string` |  |  |  |
| `spec.bootDisk.diskSizeGb` | `int32` |  |  |  |
| `spec.bootDisk.kmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.dataDisk` | `GcpVertexAiNotebookDataDisk` |  |  |  |
| `spec.dataDisk.diskType` | `string` |  |  |  |
| `spec.dataDisk.diskSizeGb` | `int32` |  |  |  |
| `spec.dataDisk.kmsKey` | `string \| valueFrom` |  |  | GcpKmsKey (`status.outputs.key_id`) |
| `spec.acceleratorConfig` | `GcpVertexAiNotebookAcceleratorConfig` |  |  |  |
| `spec.acceleratorConfig.type` | `string` |  |  |  |
| `spec.acceleratorConfig.coreCount` | `int32` |  |  |  |
| `spec.networkInterface` | `GcpVertexAiNotebookNetworkInterface` |  |  |  |
| `spec.networkInterface.network` | `string \| valueFrom` |  |  | GcpVpcNetwork (`status.outputs.network_self_link`) |
| `spec.networkInterface.subnet` | `string \| valueFrom` |  |  | GcpSubnetwork (`status.outputs.subnetwork_self_link`) |
| `spec.networkInterface.nicType` | `string` |  |  |  |
| `spec.networkInterface.externalIp` | `string \| valueFrom` |  |  | GcpAddress (`status.outputs.address`) |
| `spec.disablePublicIp` | `bool` |  |  |  |
| `spec.enableIpForwarding` | `bool` |  |  |  |
| `spec.serviceAccount` | `string \| valueFrom` |  |  | GcpServiceAccount (`status.outputs.email`) |
| `spec.tags` | `[]string` |  |  |  |
| `spec.vmImage` | `GcpVertexAiNotebookVmImage` |  |  |  |
| `spec.vmImage.project` | `string` |  |  |  |
| `spec.vmImage.family` | `string` |  |  |  |
| `spec.vmImage.name` | `string` |  |  |  |
| `spec.containerImage` | `GcpVertexAiNotebookContainerImage` |  |  |  |
| `spec.containerImage.repository` | `string` | yes |  |  |
| `spec.containerImage.tag` | `string` |  |  |  |
| `spec.shieldedInstanceConfig` | `GcpVertexAiNotebookShieldedInstanceConfig` |  |  |  |
| `spec.shieldedInstanceConfig.enableSecureBoot` | `bool` |  |  |  |
| `spec.shieldedInstanceConfig.enableVtpm` | `bool` |  |  |  |
| `spec.shieldedInstanceConfig.enableIntegrityMonitoring` | `bool` |  |  |  |
| `spec.confidentialInstanceConfig` | `GcpVertexAiNotebookConfidentialInstanceConfig` |  |  |  |
| `spec.confidentialInstanceConfig.confidentialInstanceType` | `string` |  |  |  |
| `spec.confidentialInstanceConfig.enabled` | `bool` |  | `true` |  |
| `spec.reservationAffinity` | `GcpVertexAiNotebookReservationAffinity` |  |  |  |
| `spec.reservationAffinity.consumeReservationType` | `string` |  |  |  |
| `spec.reservationAffinity.key` | `string` |  |  |  |
| `spec.reservationAffinity.values` | `[]string` |  |  |  |
| `spec.enableManagedEuc` | `bool` |  |  |  |
| `spec.enableThirdPartyIdentity` | `bool` |  |  |  |
| `spec.deletionPolicy` | `string` |  |  |  |

## Field Details

### spec.projectId

`string | valueFrom`

GCP project where the notebook instance will be created.
If omitted, the instance is created in the provider's default
project (from the credential or ambient configuration).

- references: GcpProject (`status.outputs.project_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpProject, name: <that resource's name>, fieldPath: status.outputs.project_id}} -- a bare string does not parse

### spec.location

`string` · required

GCP zone where the notebook instance will be created (e.g., "us-central1-a").
Immutable after creation.

- rule: {"required":true,"string":{"pattern":"^[a-z]+-[a-z]+[0-9]+-[a-z]$"}}

### spec.machineType

`string` · required

Compute Engine machine type for the instance.
Examples: "e2-standard-4", "n1-standard-8", "a2-highgpu-1g".
Choose based on workload requirements -- CPU-only for data processing,
N1/A2 for GPU-accelerated ML training.

- rule: {"required":true,"string":{"minLen":"1"}}

### spec.instanceName

`string`

Name of the Workbench instance in GCP.
If not specified, defaults to metadata.name.
Immutable after creation. Must be a valid RFC1035 hostname.

- rule: instance_name must be a valid RFC1035 hostname: lowercase letters, digits, and hyphens, starting with a letter

### spec.instanceOwners

`[]string`

Email addresses of users who own the instance.
Format: alias@example.com. Currently GCP supports one owner only.
If set, access mode is Single User (only this owner can access).
Immutable after creation.

### spec.desiredState

`string`

Desired state of the instance: ACTIVE (running) or STOPPED.
Use STOPPED to suspend the instance and stop billing for compute
(storage charges still apply). Defaults to ACTIVE.

- rule: desired_state must be one of: ACTIVE, STOPPED

### spec.disableProxyAccess

`bool`

If true, the notebook instance will not register with the proxy
and no JupyterLab proxy URL will be generated. Use this for
instances accessed only via SSH or other direct methods.
Immutable after creation.

### spec.metadata

`map<string, string>`

Custom metadata key-value pairs for the instance.
Some keys trigger special behaviors (e.g., install-monitoring-agent).

### spec.labels

`map<string, string>`

User-defined labels to organize the instance (cost attribution,
team ownership, environment tagging). Keys and values must follow
GCP label rules: lowercase letters, digits, underscores, and dashes,
at most 63 characters. Merged with the platform's attribution labels;
on key conflicts the platform labels win. Mutable in place.

### spec.bootDisk

`GcpVertexAiNotebookBootDisk`

Boot disk configuration for the notebook instance.
If not specified, GCP provisions a 150 GB PD_SSD boot disk with
Google-managed encryption.

### spec.bootDisk.diskType

`string`

Disk type for the boot disk.
Persistent Disk: PD_STANDARD (HDD), PD_SSD, PD_BALANCED, PD_EXTREME.
Hyperdisk (newer-generation machine series only): HYPERDISK_BALANCED,
HYPERDISK_BALANCED_HIGH_AVAILABILITY, HYPERDISK_ML.
If not specified, defaults to PD_SSD.

- rule: disk_type must be one of: PD_STANDARD, PD_SSD, PD_BALANCED, PD_EXTREME, HYPERDISK_BALANCED, HYPERDISK_BALANCED_HIGH_AVAILABILITY, HYPERDISK_ML

### spec.bootDisk.diskSizeGb

`int32`

Size of the boot disk in GB.
Minimum 10 GB, maximum 64000 GB (64 TB).
If not specified, defaults to 150 GB.

- rule: disk_size_gb must be between 10 and 64000 (or omitted for default 150 GB)

### spec.bootDisk.kmsKey

`string | valueFrom`

KMS key for CMEK encryption of the boot disk.
Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
If not specified, Google-managed encryption (GMEK) is used.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.dataDisk

`GcpVertexAiNotebookDataDisk`

Data disk configuration for the notebook instance.
GCP supports exactly one data disk per Workbench instance.
If not specified, GCP provisions a 100 GB data disk with
Google-managed encryption.

### spec.dataDisk.diskType

`string`

Disk type for the data disk.
Persistent Disk: PD_STANDARD (HDD), PD_SSD, PD_BALANCED, PD_EXTREME.
Hyperdisk (newer-generation machine series only): HYPERDISK_BALANCED,
HYPERDISK_EXTREME, HYPERDISK_THROUGHPUT,
HYPERDISK_BALANCED_HIGH_AVAILABILITY, HYPERDISK_ML.
If not specified, defaults to PD_STANDARD.

- rule: disk_type must be one of: PD_STANDARD, PD_SSD, PD_BALANCED, PD_EXTREME, HYPERDISK_BALANCED, HYPERDISK_EXTREME, HYPERDISK_THROUGHPUT, HYPERDISK_BALANCED_HIGH_AVAILABILITY, HYPERDISK_ML

### spec.dataDisk.diskSizeGb

`int32`

Size of the data disk in GB.
Minimum 10 GB, maximum 64000 GB (64 TB).
If not specified, defaults to 100 GB.

- rule: disk_size_gb must be between 10 and 64000 (or omitted for default 100 GB)

### spec.dataDisk.kmsKey

`string | valueFrom`

KMS key for CMEK encryption of the data disk.
Format: projects/{project}/locations/{location}/keyRings/{ring}/cryptoKeys/{key}
If not specified, Google-managed encryption (GMEK) is used.

- references: GcpKmsKey (`status.outputs.key_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpKmsKey, name: <that resource's name>, fieldPath: status.outputs.key_id}} -- a bare string does not parse

### spec.acceleratorConfig

`GcpVertexAiNotebookAcceleratorConfig`

GPU accelerator configuration for the notebook instance.
GCP supports one accelerator configuration per instance.
Requires compatible machine types (e.g., n1-standard-* for Tesla GPUs).

### spec.acceleratorConfig.type

`string`

GPU accelerator type. Availability varies by zone and machine type:
current-generation training/inference GPUs (NVIDIA_H100_80GB,
NVIDIA_H100_MEGA_80GB, NVIDIA_H200_141GB, NVIDIA_B200) require
their matching A3/A4 machine series; NVIDIA_L4 and the Tesla
family run on G2/N1; the _VWS variants are virtual-workstation
(graphics) licenses of the same silicon.
See https://cloud.google.com/vertex-ai/docs/workbench/instances/create#accelerator
for supported types per zone.

- rule: type must be a valid accelerator type (e.g., NVIDIA_TESLA_T4, NVIDIA_L4, NVIDIA_TESLA_A100, NVIDIA_H100_80GB, NVIDIA_H200_141GB, NVIDIA_B200)

### spec.acceleratorConfig.coreCount

`int32`

Number of accelerator cores.
Valid values depend on the accelerator type (typically 1, 2, 4, or 8).

- rule: core_count must be a positive integer

### spec.networkInterface

`GcpVertexAiNotebookNetworkInterface`

Network interface configuration for the notebook instance.
GCP supports one network interface per instance.
If not specified, the instance uses the default VPC network.
Immutable after creation.

### spec.networkInterface.network

`string | valueFrom`

VPC network for the instance.
Can be a literal value (VPC name or self_link) or a reference to
a GcpVpcNetwork resource.

- references: GcpVpcNetwork (`status.outputs.network_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpVpcNetwork, name: <that resource's name>, fieldPath: status.outputs.network_self_link}} -- a bare string does not parse

### spec.networkInterface.subnet

`string | valueFrom`

Subnetwork for the instance.
Can be a literal value (subnet name or self_link) or a reference to
a GcpSubnetwork resource.

- references: GcpSubnetwork (`status.outputs.subnetwork_self_link`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpSubnetwork, name: <that resource's name>, fieldPath: status.outputs.subnetwork_self_link}} -- a bare string does not parse

### spec.networkInterface.nicType

`string`

NIC type for the network interface.
Valid values: VIRTIO_NET, GVNIC.
GVNIC provides higher bandwidth and lower latency.
If not specified, defaults to VIRTIO_NET.

- rule: nic_type must be one of: VIRTIO_NET, GVNIC

### spec.networkInterface.externalIp

`string | valueFrom`

Static external IP address for the instance (ONE_TO_ONE_NAT access
config). The address must be an unused static external IP in the
same region as the instance's zone. Reference a GcpAddress resource
to reserve and pin the IP as a first-class node, or pass a literal
IP address.

If omitted (and public IP is not disabled), GCP assigns an ephemeral
external IP from a shared pool -- the IP changes across stop/start
cycles. Pin a static address when firewall allowlists or DNS records
depend on the instance's IP.

Cannot be combined with disable_public_ip. Immutable after creation.

- references: GcpAddress (`status.outputs.address`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpAddress, name: <that resource's name>, fieldPath: status.outputs.address}} -- a bare string does not parse

### spec.disablePublicIp

`bool`

If true, no external IP is assigned to the instance.
Use this for instances that should only be accessible through
the Vertex AI proxy or a VPN/Cloud IAP tunnel.
Immutable after creation.

### spec.enableIpForwarding

`bool`

If true, enable IP forwarding on the instance. Useful for
instances acting as network gateways. Default false.
Immutable after creation.

### spec.serviceAccount

`string | valueFrom`

Service account email for the notebook VM identity.
The VM uses this service account to access GCP resources
(BigQuery, GCS, Vertex AI, etc.). Scopes are fixed to
"https://www.googleapis.com/auth/cloud-platform".
Immutable after creation.

- references: GcpServiceAccount (`status.outputs.email`)
- rule: write as {value: <literal>} or {valueFrom: {kind: GcpServiceAccount, name: <that resource's name>, fieldPath: status.outputs.email}} -- a bare string does not parse

### spec.tags

`[]string`

Compute Engine network tags for firewall rules.
Immutable after creation.

### spec.vmImage

`GcpVertexAiNotebookVmImage`

VM image configuration for the notebook environment.
Uses pre-built deep learning VM images from GCP with popular
ML frameworks (TensorFlow, PyTorch, JAX) and JupyterLab.
Mutually exclusive with container_image.
Immutable after creation.

### spec.vmImage.project

`string`

Google Cloud project that the VM image belongs to.
GCP's own Workbench images live in "cloud-notebooks-managed".
(The legacy deep-learning-VM notebook families under
"deeplearning-platform-release" have been retired by GCP and no
longer resolve.)

### spec.vmImage.family

`string`

VM image family. The newest image in this family will be used.
GCP's maintained family for Workbench instances is
"workbench-instances" (in the "cloud-notebooks-managed" project).
Mutually exclusive with name (within this message).

### spec.vmImage.name

`string`

Specific VM image name. Use this instead of family when you need
to pin to an exact image version.
Mutually exclusive with family (within this message).

### spec.containerImage

`GcpVertexAiNotebookContainerImage`

Container image configuration for a custom notebook environment.
Use this when pre-built VM images don't meet your needs.
Mutually exclusive with vm_image.

### spec.containerImage.repository

`string` · required

Container image repository path.
Example: "gcr.io/deeplearning-platform-release/base-cu113.py310"

- rule: {"required":true}

### spec.containerImage.tag

`string`

Container image tag. If not specified, the latest tag is used.

### spec.shieldedInstanceConfig

`GcpVertexAiNotebookShieldedInstanceConfig`

Shielded VM configuration for enhanced security.
Shielded VMs protect against rootkits and bootkits with
Secure Boot, vTPM, and integrity monitoring.

### spec.shieldedInstanceConfig.enableSecureBoot

`bool` · optional (explicit presence)

Enable Secure Boot. Ensures only verified boot software runs.
Disabled by default because some ML libraries may not have signed
boot loaders.

### spec.shieldedInstanceConfig.enableVtpm

`bool` · optional (explicit presence)

Enable vTPM (Virtual Trusted Platform Module).
Provides measured boot integrity and key generation.
Enabled by default; an explicit false turns it off.

### spec.shieldedInstanceConfig.enableIntegrityMonitoring

`bool` · optional (explicit presence)

Enable integrity monitoring. Compares boot measurements against
a trusted baseline.
Enabled by default; an explicit false turns it off.

### spec.confidentialInstanceConfig

`GcpVertexAiNotebookConfidentialInstanceConfig`

Confidential Computing configuration. When set, guest memory is
encrypted in use with AMD SEV -- data stays protected even from the
host hypervisor. Requires an SEV-capable AMD machine type (e.g., the
n2d family). Immutable after creation.

### spec.confidentialInstanceConfig.confidentialInstanceType

`string`

Confidential computing technology for the instance.
The only supported value is SEV (AMD Secure Encrypted Virtualization).
If not specified, defaults to SEV.

- rule: confidential_instance_type must be SEV (the only supported confidential technology)

### spec.confidentialInstanceConfig.enabled

`bool` · optional (explicit presence)

Whether Confidential Computing is on. Unset means on: declaring the
block has always meant enabling it, and this switch lets a manifest say
the opposite out loud.

- default: `true`

### spec.reservationAffinity

`GcpVertexAiNotebookReservationAffinity`

Compute Engine reservation affinity. Points the notebook VM at
pre-purchased capacity reservations -- the way organizations
guarantee GPU availability for ML workloads. Immutable after creation.

- rule: key and values may only be set when consume_reservation_type is RESERVATION_SPECIFIC

### spec.reservationAffinity.consumeReservationType

`string`

How the instance consumes reservations:
  - RESERVATION_ANY (default): consume any matching open reservation.
  - RESERVATION_SPECIFIC: consume only the reservation named in
    key/values.
  - RESERVATION_NONE: never consume reserved capacity (on-demand only).

- rule: consume_reservation_type must be one of: RESERVATION_NONE, RESERVATION_ANY, RESERVATION_SPECIFIC

### spec.reservationAffinity.key

`string`

Corresponds to the label key of a reservation resource. To target a
SPECIFIC_RESERVATION by name, use "compute.googleapis.com/reservation-name"
as the key. Only valid with RESERVATION_SPECIFIC.

### spec.reservationAffinity.values

`[]string`

Corresponds to the label values of a reservation resource -- for the
reservation-name key, the reservation's name. Only valid with
RESERVATION_SPECIFIC.

### spec.enableManagedEuc

`bool`

Enable managed end-user credentials (EUC) for the instance. With
managed EUC, JupyterLab runs as the accessing user's own Google
identity rather than the VM's service account, so notebook code sees
the user's IAM permissions (single-user auditability). Mutable in place.

### spec.enableThirdPartyIdentity

`bool`

Allow access to the notebook through a third-party identity provider
configured on the organization (workforce identity federation).
Leave false when all users authenticate with Google identities.
Mutable in place.

### spec.deletionPolicy

`string`

Deletion policy for the notebook — what happens when this resource
is destroyed:
  ""        -- same as "DELETE" (provider default)
  "DELETE"  -- the instance is deleted, including its boot and data
               disks; unsynced notebook work on those disks is lost
  "PREVENT" -- destroy FAILS; a guard for a workstation whose local
               disks hold work not yet pushed anywhere else
  "ABANDON" -- the instance is removed from management but left
               running (and billing) in GCP with its disks intact

- rule: deletion_policy must be one of: DELETE, PREVENT, ABANDON

## Validation Rules

- `vm_image_container_image_mutual_exclusion`: only one of vm_image or container_image can be set, not both
- `external_ip_conflicts_with_disable_public_ip`: network_interface.external_ip cannot be set when disable_public_ip is true

## Outputs

Reference an output from another manifest as `valueFrom: {kind: GcpVertexAiNotebook, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.instance_id` | `string` | The fully qualified instance ID. Format: projects/{project}/locations/{location}/instances/{instance_id} |
| `status.outputs.instance_name` | `string` | The short instance name (same as the spec's instance_name input, or metadata.name if instance_name was not specified). |
| `status.outputs.proxy_uri` | `string` | The proxy URI for accessing JupyterLab. This is the primary endpoint users use to access their notebook. Empty if disable_proxy_access is true. |
| `status.outputs.state` | `string` | The current state of the instance. Possible values: ACTIVE, STOPPED, DELETED, UPGRADING, INITIALIZING, SUSPENDING, SUSPENDED, STARTING, STOPPING. |
| `status.outputs.creator` | `string` | Email address of the entity that sent the original CreateInstance request. |
| `status.outputs.create_time` | `string` | RFC3339 timestamp of when the instance was created. |
| `status.outputs.health_state` | `string` | Instance health as reported by the Workbench health service. Possible values: HEALTHY, UNHEALTHY, AGENT_NOT_INSTALLED, AGENT_NOT_RUNNING, HEALTH_STATE_UNSPECIFIED. |
| `status.outputs.update_time` | `string` | RFC3339 timestamp of the most recent instance update. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.projectId` | GcpProject | `status.outputs.project_id` |
| `spec.bootDisk.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.dataDisk.kmsKey` | GcpKmsKey | `status.outputs.key_id` |
| `spec.networkInterface.network` | GcpVpcNetwork | `status.outputs.network_self_link` |
| `spec.networkInterface.subnet` | GcpSubnetwork | `status.outputs.subnetwork_self_link` |
| `spec.networkInterface.externalIp` | GcpAddress | `status.outputs.address` |
| `spec.serviceAccount` | GcpServiceAccount | `status.outputs.email` |

## See Also

- [Overview](../README.md)
