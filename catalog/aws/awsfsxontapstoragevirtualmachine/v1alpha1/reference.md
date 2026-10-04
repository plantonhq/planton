# AwsFsxOntapStorageVirtualMachine

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsFsxOntapStorageVirtualMachineSpec defines the desired configuration for an
Amazon FSx for NetApp ONTAP Storage Virtual Machine (SVM). An SVM is a
logical data server within an FSx for ONTAP file system that provides
multi-protocol data access (NFS, SMB, iSCSI) and serves as the parent
container for ONTAP volumes.

SVMs are the primary data access layer in the ONTAP architecture:
- **File System** provides the physical infrastructure (storage, throughput,
  networking, HA)
- **SVM** provides the logical data server (protocols, endpoints, AD
  integration, security style)
- **Volume** provides the data containers (capacity, tiering, snapshots,
  SnapLock)

Each SVM has its own set of network endpoints (NFS, SMB, iSCSI, management),
enabling multi-tenancy on a single file system. NFS and iSCSI endpoints are
always available. SMB endpoints require Active Directory configuration.

Key design notes:
- `file_system_id` and `name` are ForceNew — changing them requires replacing
  the SVM. The file system must be an AwsFsxOntapFileSystem.
- `root_volume_security_style` is ForceNew — it sets the default security
  style for all volumes created under this SVM.
- Everything else updates in place: `svm_admin_password` and the entire
  `active_directory_configuration` block (domain join details included) can
  be changed after creation without replacing the SVM.
- Active Directory is OPTIONAL — only required for SMB access. NFS-only and
  iSCSI-only SVMs do not need AD. Unlike Windows FSx (where AD is mandatory),
  ONTAP SVMs support self-managed AD only (no AWS Managed Microsoft AD).
- The `svm_admin_password` provides SSH access to the SVM management endpoint
  for ONTAP CLI operations scoped to this SVM.
- Endpoints are computed outputs: iSCSI, management, NFS, and SMB (SMB only
  when AD is configured).
- Credentials, region, and deployment workflow live outside this spec in IaC
  inputs.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: test-svm
  id: awsfxosvm-test-svm
  org: test-org
  env: dev
spec:
  region: us-west-2
  file_system_id:
    value: fs-0123456789abcdef0
  name: svm_test
  root_volume_security_style: MIXED
  svm_admin_password: VsAdmin2024!
  # Full AD block so the offline plan proof exercises the SMB join arm.
  active_directory_configuration:
    netbios_name: SVMTEST
    domain_name: corp.example.com
    dns_ips:
      - 10.0.0.10
      - 10.0.1.10
    username: svc_fsx_join
    password: ADJoinP@ssw0rd!
    file_system_administrators_group: FSx Admins
    organizational_unit_distinguished_name: OU=FSx,DC=corp,DC=example,DC=com
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.fileSystemId` | `string \| valueFrom` | yes |  | AwsFsxOntapFileSystem (`status.outputs.file_system_id`) |
| `spec.name` | `string` | yes |  |  |
| `spec.rootVolumeSecurityStyle` | `string` |  | `UNIX` |  |
| `spec.svmAdminPassword` | `string` (sensitive) |  |  |  |
| `spec.activeDirectoryConfiguration` | `AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration` |  |  |  |
| `spec.activeDirectoryConfiguration.netbiosName` | `string` |  |  |  |
| `spec.activeDirectoryConfiguration.domainName` | `string` | yes |  |  |
| `spec.activeDirectoryConfiguration.dnsIps` | `[]string` | yes |  |  |
| `spec.activeDirectoryConfiguration.username` | `string` | yes |  |  |
| `spec.activeDirectoryConfiguration.password` | `string` (sensitive) | yes |  |  |
| `spec.activeDirectoryConfiguration.fileSystemAdministratorsGroup` | `string` |  | `Domain Admins` |  |
| `spec.activeDirectoryConfiguration.organizationalUnitDistinguishedName` | `string` |  |  |  |

## Field Details

### spec.region

`string` · required

The AWS region where the resource will be created.
Example: "us-west-2", "eu-west-1"

- rule: {"string":{"minLen":"1"}}

### spec.fileSystemId

`string | valueFrom` · required

The ID of the FSx for ONTAP file system that this SVM belongs to. Required.
ForceNew — the SVM cannot be moved to a different file system.

The file system provides the underlying storage, throughput, networking,
and HA infrastructure. Multiple SVMs can share a single file system for
multi-tenancy scenarios.

- references: AwsFsxOntapFileSystem (`status.outputs.file_system_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsFsxOntapFileSystem, name: <that resource's name>, fieldPath: status.outputs.file_system_id}} -- a bare string does not parse

### spec.name

`string` · required

The name of the SVM within the ONTAP file system. Required. ForceNew.

This is the ONTAP SVM name (not the Planton metadata name). It must be
unique within the file system and is used as the SVM identity in ONTAP CLI
operations, SnapMirror relationships, and DNS endpoint names.

Constraints: 1-47 characters, alphanumeric and underscore only (no hyphens,
no spaces, no special characters).

- rule: {"string":{"minLen":"1","maxLen":"47"}}

### spec.rootVolumeSecurityStyle

`string` · optional (explicit presence)

The security style for the root volume and the default for all volumes
created under this SVM. ForceNew — cannot be changed after creation.

- "UNIX": UNIX permissions (mode bits, uid/gid). Best for Linux/NFS
  workloads. This is the most common choice for NFS-only SVMs.
- "NTFS": Windows ACLs. Best for Windows/SMB workloads with Active
  Directory. Requires AD configuration for meaningful use.
- "MIXED": Both UNIX and NTFS permissions. The effective security style
  depends on which protocol last set permissions. Use for mixed NFS/SMB
  access patterns (advanced use case).

Default: UNIX

- default: `UNIX`

### spec.svmAdminPassword

`string` · sensitive

Password for the SVM administrative user ("vsadmin"). Enables SSH access
to the SVM management endpoint for ONTAP CLI operations scoped to this SVM.

The vsadmin account can manage volumes, LIFs, export policies, and other
SVM-scoped resources. Unlike the file system's fsxadmin account (which has
cluster-wide access), vsadmin is limited to this SVM.

Length: 8-50 characters. Optional — omit if SVM CLI access is not needed.
This value is sensitive and will not be returned in read operations.

### spec.activeDirectoryConfiguration

`AwsFsxOntapStorageVirtualMachineActiveDirectoryConfiguration`

Active Directory configuration for enabling SMB protocol access. Optional.

When configured, the SVM joins the specified AD domain and an SMB endpoint
is automatically created. Without AD, only NFS and iSCSI endpoints are
available.

ONTAP SVMs support only self-managed Active Directory (on-premises, on EC2,
or Azure AD DS). AWS Managed Microsoft AD is not supported for ONTAP SVMs.

- rule: netbios_name must be 1-15 characters when provided

### spec.activeDirectoryConfiguration.netbiosName

`string`

NetBIOS name for the SVM's computer object in Active Directory.

This is the short name (up to 15 characters) that identifies the SVM in
the AD domain. If omitted, AWS generates a name automatically.

Constraints: 1-15 characters, alphanumeric only.

### spec.activeDirectoryConfiguration.domainName

`string` · required

Fully qualified domain name of the Active Directory directory.
Example: "corp.example.com"

Constraints: 1-255 characters. Can be changed after creation (the SVM
re-joins the new domain).

- rule: {"string":{"minLen":"1","maxLen":"255"}}

### spec.activeDirectoryConfiguration.dnsIps

`[]string` · required

IP addresses of the DNS servers for the AD domain. Required. These must be
reachable from the file system's subnets (typically in the same VPC CIDR or
connected via VPN/peering).

Minimum 1, maximum 3 IPv4 addresses.

- rule: {"repeated":{"minItems":"1","maxItems":"3","items":{"cel":[{"id":"dns_ip_format","message":"each dns_ips entry must be an IPv4 address (e.g., '10.0.0.10')","expression":"this.matches('^([0-9]{1,3}\\\\.){3}[0-9]{1,3}$')"}]}}}

### spec.activeDirectoryConfiguration.username

`string` · required

Service account username for AD domain join operations. Required.

This account must have permissions to create computer objects in the
specified OU (or the default Computers container).

Constraints: 1-256 characters.

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.activeDirectoryConfiguration.password

`string` · required · sensitive

Service account password for AD domain join operations. Required.

This value is sensitive and will not be returned in read operations.
For production workloads, consider injecting this value via CI/CD secrets
management rather than storing it in the resource manifest.

Constraints: 1-256 characters.

- rule: {"string":{"minLen":"1","maxLen":"256"}}

### spec.activeDirectoryConfiguration.fileSystemAdministratorsGroup

`string` · optional (explicit presence)

Name of the AD domain group whose members are granted administrative
privileges on the SVM for file share management.

Constraints: 1-256 characters.
Default: Domain Admins

- default: `Domain Admins`
- rule: {"string":{"maxLen":"256"}}

### spec.activeDirectoryConfiguration.organizationalUnitDistinguishedName

`string`

Organizational Unit (OU) distinguished name within the AD directory where
the SVM's computer object is created.

Example: "OU=FSx,DC=corp,DC=example,DC=com"

Only the OU immediately above the computer object can be specified.
If not provided, the computer object is created in the default "Computers"
container in the AD domain.

Constraints: up to 2000 characters.

- rule: {"string":{"maxLen":"2000"}}

## Validation Rules

- `security_style_valid`: root_volume_security_style must be 'UNIX', 'NTFS', or 'MIXED'
- `admin_password_length`: svm_admin_password must be 8-50 characters when provided
- `name_format`: name must contain only alphanumeric characters and underscores (no hyphens, spaces, or special characters)

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsFsxOntapStorageVirtualMachine, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.svm_id` | `string` | The ID of the SVM (e.g., "svm-0123456789abcdef0"). Primary identifier used by ONTAP volumes and other AWS services referencing this SVM. |
| `status.outputs.arn` | `string` | The Amazon Resource Name of the SVM. Used in IAM policies for resource-level permissions. |
| `status.outputs.uuid` | `string` | The universally unique identifier of the SVM in ONTAP. Used for SnapMirror relationships and ONTAP REST API operations. |
| `status.outputs.subtype` | `string` | The SVM subtype (e.g., "DEFAULT"). Indicates the SVM's functional role within the file system. |
| `status.outputs.iscsi_dns_name` | `string` | The iSCSI endpoint DNS name. Used by iSCSI initiators for block-level storage access to volumes on this SVM. |
| `status.outputs.iscsi_ip_addresses` | `[]string` | The iSCSI endpoint IP addresses. Alternative to DNS for direct IP access to the iSCSI target portal. |
| `status.outputs.management_dns_name` | `string` | The management endpoint DNS name. Used for SSH (ONTAP CLI) and REST API access scoped to this SVM. Connect via: ssh vsadmin@<management_dns_name> |
| `status.outputs.management_ip_addresses` | `[]string` | The management endpoint IP addresses. Alternative to DNS for direct IP access to the SVM management interface. |
| `status.outputs.nfs_dns_name` | `string` | The NFS endpoint DNS name. Used by NFS clients to mount volumes on this SVM. Mount command: mount -t nfs <nfs_dns_name>:/vol1 /mnt/vol1 |
| `status.outputs.nfs_ip_addresses` | `[]string` | The NFS endpoint IP addresses. Alternative to DNS for direct IP-based NFS mounts. |
| `status.outputs.smb_dns_name` | `string` | The SMB endpoint DNS name. Used by Windows clients for SMB/CIFS file share access. UNC path: \\<smb_dns_name>\share_name. Only available when the SVM is joined to an Active Directory domain. |
| `status.outputs.smb_ip_addresses` | `[]string` | The SMB endpoint IP addresses. Alternative to DNS for direct IP-based SMB access. Only populated when Active Directory is configured. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.fileSystemId` | AwsFsxOntapFileSystem | `status.outputs.file_system_id` |

## Referenced By

Fields on other kinds that can point at this resource:

| Kind | Field | Reads |
|---|---|---|
| AwsFsxOntapVolume | `spec.storageVirtualMachineId` | `status.outputs.svm_id` |

## See Also

- [Overview](../README.md)
