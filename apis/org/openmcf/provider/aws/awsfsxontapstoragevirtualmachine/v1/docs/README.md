# AwsFsxOntapStorageVirtualMachine — Technical Reference

## Architecture

Amazon FSx for NetApp ONTAP uses a three-tier architecture:

1. **File System** — The physical storage infrastructure. Provides SSD/HDD
   capacity, throughput bandwidth (via HA pairs), network interfaces, and
   encryption. Managed by `AwsFsxOntapFileSystem`.

2. **Storage Virtual Machine (SVM)** — A logical data server within the file
   system. Each SVM has its own network endpoints (NFS, SMB, iSCSI,
   management), security style, and optional Active Directory integration.
   Multiple SVMs can coexist on a single file system for multi-tenancy.

3. **Volume** — A data container within an SVM. Provides storage capacity,
   tiering policies, snapshot management, and optionally SnapLock for
   WORM compliance. Managed by `AwsFsxOntapVolume`.

This component manages tier 2 — the SVM.

## SVM Endpoints

Every SVM automatically provisions the following endpoints:

| Endpoint | Protocol | Port(s) | Always Available | Purpose |
|----------|----------|---------|------------------|---------|
| iSCSI | iSCSI | 3260 | Yes | Block-level storage (LUNs) |
| Management | SSH/REST | 22, 443 | Yes (if password set) | ONTAP CLI/API for SVM admin |
| NFS | NFS | 2049, 111, 635 | Yes | File-level NFS access |
| SMB | SMB/CIFS | 445 | Only with AD | File-level SMB access |

The SMB endpoint is only provisioned when the SVM joins an Active Directory
domain. Without AD configuration, only iSCSI, management, and NFS endpoints
are available.

## Security Styles

The `root_volume_security_style` determines the permission model for the SVM's
root volume and serves as the default for all volumes created under the SVM.

### UNIX

- Permissions: POSIX mode bits (rwxrwxrwx) + uid/gid ownership
- Best for: Linux clients, NFS mounts, Kubernetes PVs
- AD required: No
- Most common choice for pure NFS workloads

### NTFS

- Permissions: Windows ACLs (DACLs/SACLs) with SID-based identity
- Best for: Windows clients, SMB shares, .NET applications
- AD required: Yes (for meaningful permission management)
- Supports Windows features: inheritance, deny ACEs, audit

### MIXED

- Permissions: Both UNIX and NTFS, determined by last protocol to modify
- Best for: Dual-protocol environments (NFS + SMB on same data)
- AD required: Yes (for SMB access)
- Caution: Permission model is complex — the last protocol to set permissions
  on a file or directory determines the effective style. This can lead to
  surprising behavior in mixed environments.

## Active Directory Integration

ONTAP SVMs support self-managed Active Directory only. AWS Managed Microsoft AD
(Directory Service) is not supported for ONTAP SVMs. This differs from FSx for
Windows File Server, which supports both.

### Domain Join Process

When `active_directory_configuration` is specified:

1. SVM creates a computer object in the AD domain (in the specified OU or
   default "Computers" container)
2. SVM registers DNS records for its SMB endpoint
3. SMB endpoint becomes available for client connections
4. Windows ACLs and identity-based access become functional

### Credential Management

The AD service account credentials (`username`/`password`) are used only during
the domain join operation and subsequent AD updates. For production deployments:

- Use a dedicated service account with minimal privileges (only "Create Computer
  Objects" permission in the target OU)
- Rotate credentials periodically using the update operation
- Inject credentials via CI/CD secrets management rather than storing in manifests

### NetBIOS Name

The `netbios_name` (up to 15 characters) identifies the SVM's computer object in
AD. If not specified, AWS generates a name automatically. For predictable DNS
names and easier AD management, always specify the NetBIOS name explicitly.

## Multi-Tenancy

A single FSx ONTAP file system can host multiple SVMs, each with:

- Independent protocol endpoints
- Separate security configurations
- Different AD domain memberships (or no AD)
- Isolated volume namespaces

This enables scenarios like:

- **Team isolation**: Each team gets its own SVM with separate permissions
- **Protocol separation**: One UNIX/NFS SVM + one NTFS/SMB SVM on the same
  file system
- **Environment isolation**: Dev, staging, prod SVMs on shared infrastructure

## ForceNew Fields

The following fields cannot be changed after creation — modifying them requires
replacing the SVM:

| Field | Why ForceNew |
|-------|-------------|
| `file_system_id` | SVM is bound to a specific file system |
| `name` | ONTAP SVM identity is immutable |
| `root_volume_security_style` | Changing would break existing volume permissions |

## Cost Considerations

SVMs themselves have no direct cost — they are logical constructs within the
file system. Costs are determined by:

- **File system** — storage capacity, throughput tier, HA pairs
- **Volumes** — storage consumed within the file system
- **Data transfer** — cross-AZ and internet egress

Creating multiple SVMs on a single file system is a cost-neutral way to achieve
multi-tenancy and workload isolation.

## Relationship to Other Components

| Component | Relationship |
|-----------|-------------|
| AwsFsxOntapFileSystem | Parent (required). Provides infrastructure |
| AwsFsxOntapVolume | Child. Created within this SVM |
| AwsVpc | Indirect. SVM inherits networking from file system |
| AwsSecurityGroup | Indirect. File system's security groups control access |

## Deliberately Omitted (v1)

- **SVM peering** — Cross-file-system SVM peering for SnapMirror. Independent
  lifecycle, managed via ONTAP CLI.
- **Export policies** — NFS export policies are volume-level configuration,
  managed via ONTAP CLI or as part of volume provisioning.
- **CIFS shares** — SMB share definitions are created within volumes, not at the
  SVM level.
- **Kerberos configuration** — Advanced NFS Kerberos is configured via ONTAP CLI
  after SVM creation.
