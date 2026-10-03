# AWS FSx ONTAP Storage Virtual Machine

Deploys an ONTAP Storage Virtual Machine (SVM) on an existing FSx for NetApp ONTAP file system, providing multi-protocol data access endpoints for NFS, iSCSI, and optionally SMB via Active Directory. The SVM is the data-serving layer of the ONTAP family: the file system provides capacity and throughput, the SVM provides the endpoints clients mount, and volumes (AwsFsxOntapVolume) attach to the SVM by its `svm_id` output.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **ONTAP Storage Virtual Machine** -- a logical data server within the specified FSx ONTAP file system, with configurable root volume security style (UNIX, NTFS, or MIXED) and optional SVM admin password for vsadmin SSH access
- **NFS Endpoint** -- automatically provisioned for NFS client mounts to volumes on this SVM
- **iSCSI Endpoint** -- automatically provisioned for block-level storage access via iSCSI initiators
- **Management Endpoint** -- automatically provisioned for ONTAP CLI (SSH) and REST API access scoped to this SVM
- **SMB Endpoint** -- created only when Active Directory is configured; enables Windows SMB/CIFS file share access with identity-based permissions
- **AD Computer Object** -- created only when Active Directory is configured; registers the SVM in the specified AD domain and organizational unit
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied automatically

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **An FSx for ONTAP file system** -- the SVM's parent file system must be provisioned first. Provide the file system ID directly or reference an AwsFsxOntapFileSystem Infra Component via ValueFromRef.
- **A self-managed Active Directory domain** (optional) -- required only for SMB access. Must be reachable from the file system's VPC. AWS Managed Microsoft AD is not supported for ONTAP SVMs.
- **AD service account credentials** (optional) -- required only for Active Directory domain join. The account must have permissions to create computer objects in the target OU.

## Deploy

### Console

Open the deployment store, find **AWS FSx ONTAP Storage Virtual Machine**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **NFS-Only UNIX SVM** preset in the [Presets](#presets) tab to pre-populate the simplest configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsFsxOntapStorageVirtualMachine
metadata:
  name: app-svm
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  fileSystemId:
    value: "fs-0123456789abcdef0"
  name: svm_app
  rootVolumeSecurityStyle: UNIX
```

```shell
planton apply -f fsx-ontap-svm.yaml
```

This creates an NFS/iSCSI-only SVM with UNIX security style on the specified ONTAP file system. No Active Directory is configured, so SMB endpoints are not available. An Infra Job tracks the provisioning in real time.

### InfraChart

When deploying as part of a multi-resource environment, use ValueFromRef to wire the SVM to a file system deployed in the same InfraPipeline:

```yaml
spec:
  fileSystemId:
    valueFrom:
      kind: AwsFsxOntapFileSystem
      name: production-ontap
      fieldPath: status.outputs.file_system_id
```

The InfraPipeline resolves the dependency graph, deploys the FSx ONTAP file system first, then provisions the SVM with the resolved file system ID.

## Key Configuration

These are the most important decisions when configuring an ONTAP SVM. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Root volume security style** -- Defaults to UNIX. Choose NTFS for Windows/SMB-only workloads with Active Directory, or MIXED for environments where both NFS and SMB clients access the same volumes. This setting is ForceNew -- changing it requires replacing the SVM.

**Active Directory** -- Optional. Configure only when SMB access is needed. Requires `domainName`, `dnsIps`, `username`, and `password` for the AD service account. Without AD, the SVM provides NFS and iSCSI endpoints only.

**SVM name** -- The ONTAP identity for this SVM (not the Planton metadata name). Must be alphanumeric plus underscore only, 1-47 characters. ForceNew -- cannot be renamed after creation.

**SVM admin password** -- Optional. Enables SSH access to the vsadmin account for SVM-scoped ONTAP CLI operations. Omit if SVM-level CLI access is not needed; manage the file system through the fsxadmin account instead.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsFsxOntapFileSystem** | `fileSystemId` | `status.outputs.file_system_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `svm_id` | SVM identifier (e.g., svm-0123456789abcdef0) | AwsFsxOntapVolume parent reference |
| `arn` | Amazon Resource Name of the SVM | IAM policies for resource-level permissions |
| `uuid` | ONTAP UUID for SnapMirror and REST API operations | Cross-cluster volume identification |
| `subtype` | SVM subtype (e.g., DEFAULT) | Operational metadata |
| `iscsi_dns_name` | iSCSI endpoint DNS name | iSCSI initiator target discovery |
| `iscsi_ip_addresses` | iSCSI endpoint IP addresses | Direct IP-based iSCSI access |
| `management_dns_name` | Management endpoint DNS name | ONTAP CLI access via SSH |
| `management_ip_addresses` | Management endpoint IP addresses | Direct IP-based SVM management |
| `nfs_dns_name` | NFS endpoint DNS name | NFS mount commands for client volumes |
| `nfs_ip_addresses` | NFS endpoint IP addresses | Direct IP-based NFS mounts |
| `smb_dns_name` | SMB endpoint DNS name (AD required) | Windows UNC path file share access |
| `smb_ip_addresses` | SMB endpoint IP addresses (AD required) | Direct IP-based SMB access |

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**NFS-only UNIX SVM** -- UNIX security style, no Active Directory, no SMB. The simplest configuration for Linux/NFS workloads, Kubernetes persistent volumes, and data pipelines. Start from the **NFS-Only UNIX SVM** preset.

**SMB Windows SVM with Active Directory** -- NTFS security style with AD domain join for Windows file share access. Includes SVM admin password and OU placement. Start from the **SMB Windows SVM with Active Directory** preset.

**Multiprotocol SVM (NFS + SMB)** -- MIXED security style with Active Directory. Both NFS and SMB endpoints available for mixed Linux/Windows environments sharing the same data. Start from the **Multiprotocol SVM (NFS + SMB)** preset.

## Works With

- [**AWS FSx ONTAP File System**](/infra-catalog/aws-fsx-ontap-file-system) -- provides the parent file system infrastructure for this SVM
- [**AWS FSx ONTAP Volume**](/infra-catalog/aws-fsx-ontap-volume) -- the data volumes that attach to this SVM by `svm_id` and mount through its endpoints