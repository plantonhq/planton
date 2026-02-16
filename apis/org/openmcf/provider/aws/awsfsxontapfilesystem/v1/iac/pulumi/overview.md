# Pulumi Module Architecture: AWS FSx for NetApp ONTAP

## Overview

This Pulumi module provisions Amazon FSx for NetApp ONTAP file systems through a declarative, protobuf-defined specification. The architecture follows OpenMCF's standard pattern: input transformation → resource provisioning → output extraction.

The module exposes FSx for ONTAP's core features—deployment types (single-AZ and multi-AZ), storage types (SSD and HDD), HA pair scale-out, disk IOPS configuration, backup policies, and encryption—while maintaining simplicity through careful abstraction of AWS API complexity.

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint
├── Pulumi.yaml          # Pulumi project metadata
├── Makefile             # Build/test helpers
├── debug.sh             # Debug script for stack output inspection
└── module/
    ├── main.go          # Orchestration logic (provider setup, output export)
    ├── locals.go        # Input transformation and tag construction
    ├── file_system.go   # FSx ONTAP File System resource implementation
    └── outputs.go       # Output key constants
```

## Data Flow

```
AwsFsxOntapFileSystemStackInput (protobuf)
  ├─ target: AwsFsxOntapFileSystem
  │   ├─ metadata (org, env, name, id)
  │   └─ spec: AwsFsxOntapFileSystemSpec
  └─ provider_config (optional)
        │
        ▼
  initializeLocals → Locals + AwsTags
        │
        ▼
  AWS Provider (ambient or explicit credentials)
        │
        ▼
  fileSystem() → fsx.NewOntapFileSystem()
        │
        ▼
  Output Exports (file_system_id, management_dns_name, etc.)
```

## Key Implementation Details

### Deployment Types

| Type | Subnets | HA Pairs | Use Case |
|------|---------|----------|----------|
| SINGLE_AZ_1 | 1 | 1–12 | Legacy single-AZ |
| SINGLE_AZ_2 | 1 | 1–12 | Recommended single-AZ, in-place HA scale-out |
| MULTI_AZ_1 | 2 | 1 | Legacy multi-AZ |
| MULTI_AZ_2 | 2 | 1 | Recommended multi-AZ |

### ONTAP-Specific Outputs

Unlike FSx for Windows, ONTAP exposes two endpoint types:

- **Management:** SSH and REST API access. Used for ONTAP CLI, LIF management, SnapMirror.
- **Intercluster:** SnapMirror replication between file systems.

These are extracted from the `Endpoints` array on the Pulumi resource and exported as `management_dns_name`, `management_ip_addresses`, `intercluster_dns_name`, and `intercluster_ip_addresses`.

### Resource Hierarchy

This component provisions only the **file system**. Downstream resources (`AwsFsxOntapStorageVirtualMachine`, `AwsFsxOntapVolume`) consume `file_system_id` to create SVMs and volumes.

## File System Creation Time

FSx for ONTAP file system creation is a long-running operation:

- **Single-AZ:** Typically 20–40 minutes
- **Multi-AZ:** Typically 40–60 minutes
- **Scale-out (multiple HA pairs):** Additional time per HA pair

## ForceNew Fields

Changing these fields requires resource replacement (destroy + recreate):

- `deployment_type`
- `storage_type`
- `subnet_ids`
- `preferred_subnet_id`
- `security_group_ids`
- `kms_key_id`
- `endpoint_ip_address_range`

Use `pulumi preview` before `pulumi up` to understand replacement impact.
