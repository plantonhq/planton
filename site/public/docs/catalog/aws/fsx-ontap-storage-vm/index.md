---
title: "FSx ONTAP Storage VM"
description: "FSx ONTAP Storage VM deployment documentation"
icon: "package"
order: 100
componentName: "awsfsxontapstoragevirtualmachine"
---

# AWS FSx ONTAP Storage VM

Deploys a Storage Virtual Machine (SVM) on an existing FSx for NetApp ONTAP file system, providing multi-protocol data access endpoints for NFS, iSCSI, and optionally SMB via Active Directory integration. The SVM serves as the data access layer and parent container for ONTAP volumes.

## What Gets Created

When you deploy an AwsFsxOntapStorageVirtualMachine resource, OpenMCF provisions:

- **ONTAP Storage Virtual Machine** — an `aws_fsx_ontap_storage_virtual_machine` resource within the specified FSx ONTAP file system, with the configured security style and optional admin password
- **NFS Endpoint** — automatically provisioned for NFS client access to volumes on this SVM
- **iSCSI Endpoint** — automatically provisioned for block-level storage access via iSCSI initiators
- **Management Endpoint** — automatically provisioned for ONTAP CLI (SSH) and REST API access scoped to this SVM
- **SMB Endpoint** — created only when Active Directory is configured, enables Windows SMB/CIFS file share access with identity-based permissions
- **AD Computer Object** — created only when Active Directory is configured, registers the SVM in the specified AD domain and organizational unit

## Prerequisites

- **An existing AwsFsxOntapFileSystem** — the SVM's parent file system must be provisioned first
- **AWS credentials** configured via environment variables or OpenMCF provider config
- **A self-managed Active Directory domain** reachable from the file system's VPC if enabling SMB access (AWS Managed Microsoft AD is not supported for ONTAP SVMs)
- **AD service account credentials** with permission to create computer objects in the target OU if enabling Active Directory

## Quick Start

Create a file `svm.yaml`:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: my-svm
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsFsxOntapStorageVirtualMachine.my-svm
spec:
  region: us-east-1
  fileSystemId: fs-0123456789abcdef0
  name: svm_default
```

Deploy:

```shell
openmcf apply -f svm.yaml
```

This creates an NFS/iSCSI-only SVM with UNIX security style (the default) on the specified ONTAP file system.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `region` | `string` | AWS region where the SVM will be created (e.g., `us-east-1`). | Required; non-empty |
| `fileSystemId` | `StringValueOrRef` | ID of the parent FSx ONTAP file system. ForceNew. | Required. Can reference AwsFsxOntapFileSystem via `valueFrom`. |
| `name` | `string` | ONTAP SVM name. ForceNew. This is the ONTAP identity, not the OpenMCF metadata name. | 1-47 characters, alphanumeric and underscore only. |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `rootVolumeSecurityStyle` | `string` | `UNIX` | Security style for the root volume: `UNIX`, `NTFS`, or `MIXED`. ForceNew. Sets the default for all volumes created under this SVM. |
| `svmAdminPassword` | `string` | — | Password for the vsadmin user (SVM-scoped SSH/REST access). 8-50 characters. Sensitive. |
| `activeDirectoryConfiguration.netbiosName` | `string` | auto-generated | NetBIOS name for the SVM's AD computer object. 1-15 characters. |
| `activeDirectoryConfiguration.domainName` | `string` | — | Fully qualified AD domain name (e.g., `corp.example.com`). Required when AD is configured. |
| `activeDirectoryConfiguration.dnsIps` | `string[]` | — | AD DNS server IP addresses. Required when AD is configured. 1-3 addresses. |
| `activeDirectoryConfiguration.username` | `string` | — | AD service account username for domain join. Required when AD is configured. |
| `activeDirectoryConfiguration.password` | `string` | — | AD service account password. Required when AD is configured. Sensitive. |
| `activeDirectoryConfiguration.fileSystemAdministratorsGroup` | `string` | `Domain Admins` | AD group with administrative privileges on the SVM. |
| `activeDirectoryConfiguration.organizationalUnitDistinguishedName` | `string` | `Computers` | OU DN where the SVM's computer object is created (e.g., `OU=FSx,DC=corp,DC=example,DC=com`). |

## Examples

### NFS-Only SVM

The simplest configuration for Linux/NFS workloads:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: nfs-svm
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.AwsFsxOntapStorageVirtualMachine.nfs-svm
spec:
  region: us-east-1
  fileSystemId: fs-0123456789abcdef0
  name: svm_nfs
  rootVolumeSecurityStyle: UNIX
```

### SMB SVM with Active Directory

Windows-focused SVM with AD domain join for SMB file share access:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: smb-svm
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsFsxOntapStorageVirtualMachine.smb-svm
spec:
  region: us-east-1
  fileSystemId: fs-0123456789abcdef0
  name: svm_windows
  rootVolumeSecurityStyle: NTFS
  svmAdminPassword: VsAdmin2024!
  activeDirectoryConfiguration:
    netbiosName: SVMWIN
    domainName: corp.example.com
    dnsIps:
      - "10.0.0.1"
      - "10.0.0.2"
    username: svc_fsx_join
    password: ADJoinP@ssw0rd!
    organizationalUnitDistinguishedName: "OU=FSx,DC=corp,DC=example,DC=com"
```

### Multiprotocol SVM (NFS + SMB)

Dual-protocol SVM with MIXED security style for environments where both Linux and Windows clients access the same data:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: multi-svm
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsFsxOntapStorageVirtualMachine.multi-svm
spec:
  region: us-east-1
  fileSystemId: fs-0123456789abcdef0
  name: svm_shared
  rootVolumeSecurityStyle: MIXED
  svmAdminPassword: SharedAdmin2024!
  activeDirectoryConfiguration:
    netbiosName: SVMSHARED
    domainName: corp.example.com
    dnsIps:
      - "10.0.0.1"
      - "10.0.0.2"
    username: svc_fsx_join
    password: ADJoinP@ssw0rd!
    fileSystemAdministratorsGroup: FSx Admins
    organizationalUnitDistinguishedName: "OU=FSx,DC=corp,DC=example,DC=com"
```

### Using Foreign Key References

Reference an OpenMCF-managed FSx ONTAP file system instead of hardcoding the ID:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: linked-svm
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: prod.AwsFsxOntapStorageVirtualMachine.linked-svm
spec:
  region: us-east-1
  fileSystemId:
    valueFrom:
      kind: AwsFsxOntapFileSystem
      name: my-ontap-fs
      field: status.outputs.file_system_id
  name: svm_linked
  rootVolumeSecurityStyle: UNIX
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `svm_id` | `string` | The SVM identifier (e.g., `svm-0123456789abcdef0`) |
| `arn` | `string` | Amazon Resource Name of the SVM for IAM policies |
| `uuid` | `string` | The SVM's UUID in ONTAP, used for SnapMirror and REST API operations |
| `subtype` | `string` | The SVM subtype (e.g., `DEFAULT`) |
| `iscsi_dns_name` | `string` | DNS name for the iSCSI endpoint |
| `iscsi_ip_addresses` | `string[]` | IP addresses for the iSCSI endpoint |
| `management_dns_name` | `string` | DNS name for the SVM management endpoint (SSH/REST) |
| `management_ip_addresses` | `string[]` | IP addresses for the management endpoint |
| `nfs_dns_name` | `string` | DNS name for the NFS endpoint |
| `nfs_ip_addresses` | `string[]` | IP addresses for the NFS endpoint |
| `smb_dns_name` | `string` | DNS name for the SMB endpoint. Only populated when Active Directory is configured. |
| `smb_ip_addresses` | `string[]` | IP addresses for the SMB endpoint. Only populated when Active Directory is configured. |

## Related Components

- [AwsFsxOntapFileSystem](/docs/catalog/aws/fsx-for-ontap) — parent file system that provides the storage infrastructure for this SVM
- [AwsFsxOntapVolume](/docs/catalog/aws/fsx-ontap-volume) — data volumes created within this SVM
- [AwsVpc](/docs/catalog/aws/vpc) — provides the network subnets for the parent file system
- [AwsSecurityGroup](/docs/catalog/aws/security-group) — controls network access to the parent file system's endpoints
