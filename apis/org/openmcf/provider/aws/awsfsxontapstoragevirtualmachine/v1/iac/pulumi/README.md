# Pulumi Module: AwsFsxOntapStorageVirtualMachine

Pulumi Go module that provisions an Amazon FSx for NetApp ONTAP Storage Virtual Machine (SVM) with optional Active Directory integration.

## Module Structure

```
module/
├── main.go      — Entry point: provider setup, SVM creation, output exports
├── locals.go    — Local variables: target resource, AWS tags
├── outputs.go   — Output key constants for stack exports
└── svm.go       — SVM resource creation with optional AD configuration
```

## Resources Created

| Resource | Pulumi Type | Description |
|----------|-------------|-------------|
| SVM | `fsx.OntapStorageVirtualMachine` | The ONTAP Storage Virtual Machine |

## Outputs Exported

| Key | Source | Description |
|-----|--------|-------------|
| `svm_id` | SVM ID | Primary identifier |
| `arn` | SVM ARN | For IAM policies |
| `uuid` | SVM UUID | ONTAP identifier |
| `subtype` | SVM subtype | Functional role |
| `iscsi_dns_name` | Endpoint | iSCSI DNS |
| `iscsi_ip_addresses` | Endpoint | iSCSI IPs |
| `management_dns_name` | Endpoint | Management DNS |
| `management_ip_addresses` | Endpoint | Management IPs |
| `nfs_dns_name` | Endpoint | NFS DNS |
| `nfs_ip_addresses` | Endpoint | NFS IPs |
| `smb_dns_name` | Endpoint | SMB DNS (AD only) |
| `smb_ip_addresses` | Endpoint | SMB IPs (AD only) |

## Local Development

```bash
cd module && go build ./...
```
