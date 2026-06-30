# OCI File System

Deploys an Oracle Cloud Infrastructure File Storage file system with a dedicated mount target and one or more NFS exports. The mount target provides the network endpoint (IP address) that clients use to mount the file system via NFS. Export options control per-client access permissions, identity squashing, and privileged port requirements.

## What Gets Created

When you deploy an OciFileSystem resource, Planton provisions:

- **File System** — an `oci_file_storage_file_system` resource in the specified compartment and availability domain with optional KMS encryption and snapshot policy attachment.
- **Mount Target** — an `oci_file_storage_mount_target` resource in the specified subnet providing the NFS endpoint (private IP address). OCI automatically creates an export set on the mount target.
- **Export Set Configuration** — when `maxFsStatBytes` or `maxFsStatFiles` is set, an `oci_file_storage_export_set` resource is created to configure NFS capacity reporting via statfs on the auto-created export set.
- **NFS Exports** — one `oci_file_storage_export` per entry in `exports`. Each export connects the file system to the mount target at a specific path with optional per-source access control rules.

## Prerequisites

- **OCI credentials** configured via environment variables or Planton provider config (API Key, Instance Principal, Security Token, Resource Principal, or OKE Workload Identity)
- **A compartment OCID** where the file system and mount target will be created — either a literal value or a reference to an OciCompartment resource
- **An availability domain** — file system and mount target must be in the same AD
- **A subnet OCID** for the mount target — determines the VCN and network segment for NFS access
- **Mount target service limits** — OCI defaults to 2 mount targets per AD; request a limit increase if needed

## Quick Start

Create a file `filesystem.yaml`:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciFileSystem
metadata:
  name: my-fs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciFileSystem.my-fs
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  mountTarget:
    subnetId:
      value: "ocid1.subnet.oc1..example"
  exports:
    - path: "/shared"
```

Deploy:

```shell
planton apply -f filesystem.yaml
```

This creates a file system, a mount target in the specified subnet, and one NFS export at `/shared`. The mount target IP address is exported as a stack output for use in NFS mount commands:

```shell
mount -t nfs <mount_target_ip>:/shared /mnt/shared
```

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `compartmentId` | `StringValueOrRef` | OCID of the compartment where the file system and mount target will be created. Can reference an OciCompartment resource via `valueFrom`. | Required |
| `availabilityDomain` | `string` | Availability domain for the file system and mount target. Both must be in the same AD. Example: `"Uocm:US-ASHBURN-AD-1"`. Changing this forces recreation. | Min length 1 |
| `mountTarget` | `MountTarget` | Configuration for the dedicated NFS mount target. | Required |
| `mountTarget.subnetId` | `StringValueOrRef` | OCID of the subnet where the mount target will be created. Can reference an OciSubnet resource via `valueFrom`. Changing this forces recreation. | Required |
| `exports` | `Export[]` | NFS export paths. Each export makes the file system accessible at a specific path on the mount target. | Min 1 item |
| `exports[].path` | `string` | NFS export path. Must start with `/` and be unique within the mount target's export set. Changing this forces recreation. | Min length 1 |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `displayName` | `string` | `metadata.name` | Display name for the file system. When omitted, falls back to `metadata.name`. |
| `kmsKeyId` | `StringValueOrRef` | — | OCID of a KMS master encryption key for server-side encryption. When unset, Oracle-managed keys are used. |
| `filesystemSnapshotPolicyId` | `StringValueOrRef` | — | OCID of a filesystem snapshot policy for automated snapshots. Must be in the same availability domain. |

### MountTarget Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `mountTarget.displayName` | `string` | — | Display name for the mount target. When omitted, OCI generates one. |
| `mountTarget.hostnameLabel` | `string` | — | DNS hostname label within the VCN's DNS. Produces an FQDN like `<hostname>.<subnet>.<vcn>.oraclevcn.com`. Changing this forces recreation. |
| `mountTarget.ipAddress` | `string` | — | Specific private IP address to assign. Must be available in the subnet's CIDR. When omitted, OCI auto-assigns. Changing this forces recreation. |
| `mountTarget.nsgIds` | `StringValueOrRef[]` | — | OCIDs of network security groups for NFS traffic control (port 2049/TCP, 111/TCP). Can reference OciSecurityGroup resources via `valueFrom`. |
| `mountTarget.requestedThroughput` | `int64` | — | Requested throughput in Mbps. When omitted, OCI uses the default throughput tier. |
| `mountTarget.maxFsStatBytes` | `int64` | — | Maximum NFS capacity in bytes reported to clients via statfs. When omitted, the actual metered size is reported. |
| `mountTarget.maxFsStatFiles` | `int64` | — | Maximum file count reported to clients via statfs. When omitted, the actual count is reported. |

### Export Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `exports[].exportOptions` | `ExportOption[]` | — | NFS access control rules. When omitted, OCI applies default access. |

### ExportOption

| Field | Type | Description |
|-------|------|-------------|
| `source` | `string` | Source IP address or CIDR block allowed to access this export. Use `"0.0.0.0/0"` for unrestricted access. |
| `access` | `enum` | NFS access level. Values: `read_write`, `read_only`. |
| `identitySquash` | `enum` | Identity squashing mode. Values: `no_squash`, `root_squash`, `all_squash`. |
| `requirePrivilegedSourcePort` | `bool` | When `true`, only connections from privileged ports (< 1024) are allowed. |
| `isAnonymousAccessAllowed` | `bool` | When `true`, anonymous (unauthenticated) access is allowed. |
| `anonymousUid` | `int64` | UNIX UID for anonymous or squashed users. Typically 65534 (nobody). |
| `anonymousGid` | `int64` | UNIX GID for anonymous or squashed users. Typically 65534 (nogroup). |

## Examples

### Minimal File System

A file system with one export and default NFS access — suitable for development:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciFileSystem
metadata:
  name: dev-fs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OciFileSystem.dev-fs
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  mountTarget:
    subnetId:
      value: "ocid1.subnet.oc1..example"
  exports:
    - path: "/data"
```

### File System with DNS and Fixed IP

A file system with a predictable mount target address and DNS hostname:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciFileSystem
metadata:
  name: app-shared
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.OciFileSystem.app-shared
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  displayName: "App Shared Storage"
  mountTarget:
    subnetId:
      value: "ocid1.subnet.oc1..example"
    displayName: "app-shared-mt"
    hostnameLabel: "appshared"
    ipAddress: "10.0.1.100"
  exports:
    - path: "/app-data"

```

### Multiple Exports with Access Control

A production file system with separate exports for different teams, each with per-CIDR access rules:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciFileSystem
metadata:
  name: prod-shared
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciFileSystem.prod-shared
spec:
  compartmentId:
    value: "ocid1.compartment.oc1..example"
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  displayName: "Production Shared"
  kmsKeyId:
    value: "ocid1.key.oc1..example"
  mountTarget:
    subnetId:
      value: "ocid1.subnet.oc1..example"
    displayName: "prod-shared-mt"
    hostnameLabel: "prodshared"
    nsgIds:
      - value: "ocid1.networksecuritygroup.oc1..example"
    requestedThroughput: 1024
  exports:
    - path: "/team-a"
      exportOptions:
        - source: "10.0.1.0/24"
          access: read_write
          identitySquash: root_squash
          requirePrivilegedSourcePort: true
          anonymousUid: 65534
          anonymousGid: 65534
    - path: "/team-b"
      exportOptions:
        - source: "10.0.2.0/24"
          access: read_write
          identitySquash: no_squash
        - source: "10.0.3.0/24"
          access: read_only
          identitySquash: all_squash
          anonymousUid: 65534
          anonymousGid: 65534
```

### Using Foreign Key References

Reference Planton-managed compartment, subnet, and NSG resources instead of hardcoding OCIDs:

```yaml
apiVersion: oci.planton.dev/v1
kind: OciFileSystem
metadata:
  name: ref-fs
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OciFileSystem.ref-fs
spec:
  compartmentId:
    valueFrom:
      kind: OciCompartment
      name: prod-compartment
      fieldPath: status.outputs.compartmentId
  availabilityDomain: "Uocm:US-ASHBURN-AD-1"
  mountTarget:
    subnetId:
      valueFrom:
        kind: OciSubnet
        name: private-subnet
        fieldPath: status.outputs.subnetId
    nsgIds:
      - valueFrom:
          kind: OciSecurityGroup
          name: nfs-nsg
          fieldPath: status.outputs.networkSecurityGroupId
  exports:
    - path: "/shared"
      exportOptions:
        - source: "0.0.0.0/0"
          access: read_write
          identitySquash: root_squash
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `file_system_id` | `string` | OCID of the created file system |
| `mount_target_id` | `string` | OCID of the mount target |
| `mount_target_ip_address` | `string` | Private IP address of the mount target. Used in NFS mount commands. |
| `export_set_id` | `string` | OCID of the export set associated with the mount target |

## Related Components

- [OciSubnet](/docs/catalog/oci/ocisubnet) — provides the subnet for the mount target via `valueFrom`
- [OciSecurityGroup](/docs/catalog/oci/ocisecuritygroup) — controls NFS traffic to the mount target via `nsgIds`
- [OciCompartment](/docs/catalog/oci/ocicompartment) — provides the compartment referenced by `compartmentId` via `valueFrom`
