# Pulumi Module Architecture: AwsFsxOntapStorageVirtualMachine

## Overview

This Pulumi module provisions a single `fsx.OntapStorageVirtualMachine` resource with optional Active Directory configuration. The module follows the standard OpenMCF Pulumi pattern: load stack input, initialize locals, create resources, export outputs.

## Data Flow

```
StackInput (manifest YAML)
  → initializeLocals() → Locals struct (target + AWS tags)
  → svm() → fsx.OntapStorageVirtualMachine
  → ctx.Export() → 12 stack outputs (4 endpoint types × 3 fields + 4 identity)
```

## Key Implementation Details

### Active Directory Mapping

The proto spec uses a flattened AD configuration (single message), but the Pulumi SDK requires the nested `ActiveDirectoryConfiguration` → `SelfManagedActiveDirectoryConfiguration` structure. The `svm.go` file handles this mapping transparently.

### Endpoint Extraction

SVM endpoints are computed (not configurable) and nested inside the `Endpoints` output array. The module uses `ApplyT()` to extract individual endpoint fields (dns_name, ip_addresses) for each of the 4 endpoint types: iSCSI, management, NFS, SMB.

### Tags

Standard AWS tags are applied: Resource, Organization, Environment, ResourceKind, ResourceId. These follow the `awstagkeys` package conventions.
