# GcpVpc: Private Services Access for Google Managed Services

**Date**: December 29, 2025
**Type**: Feature
**Components**: GCP Provider, GcpVpc, Pulumi Module, Terraform Module, API Definitions

## Summary

Added Private Services Access support to the GcpVpc component, enabling VPC peering with Google's service network. This allows Google managed services like Cloud SQL, Memorystore, and Filestore to use private IP addresses, providing secure connectivity without exposing databases to the public internet. The implementation spans proto schema, Pulumi module, Terraform module, and documentation.

## Problem Statement / Motivation

When deploying Cloud SQL with `privateIpEnabled: true`, deployments were failing with `SERVICE_NETWORKING_NOT_ENABLED` errors. Users coming from AWS expected a simple "put database in private subnet" model, but GCP's architecture for managed service networking is fundamentally different.

### Pain Points

- **Deployment failures**: Cloud SQL with private IP failed without clear guidance on prerequisites
- **Conceptual gap**: AWS-to-GCP migration users didn't understand the VPC peering model
- **Manual setup required**: Users had to manually configure Private Services Access via gcloud or console
- **No automation**: The GcpVpc module didn't support private service connectivity out of the box
- **Documentation gap**: The relationship between VPC configuration and managed service connectivity wasn't clear

### The AWS vs GCP Networking Difference

The root cause was a fundamental architectural difference between AWS and GCP that trips up many users:

```mermaid
flowchart TB
    subgraph aws [AWS Model]
        direction TB
        awsvpc[Your VPC]
        awssubnet[Private Subnet]
        awsrds[RDS Instance]
        awsvpc --> awssubnet
        awssubnet --> awsrds
        note1[RDS lives INSIDE your VPC]
    end
    
    subgraph gcp [GCP Model]
        direction TB
        gcpvpc[Your VPC]
        gcppeering{VPC Peering}
        googlenet[Google Service Network]
        cloudsql[Cloud SQL]
        gcpvpc <--> gcppeering
        gcppeering <--> googlenet
        googlenet --> cloudsql
        note2[Cloud SQL lives OUTSIDE your VPC]
    end
```

**AWS**: RDS instances live inside your VPC subnets. You specify a subnet group, and the database gets a private IP from your subnet's CIDR range.

**GCP**: Cloud SQL (and other managed services) live in Google's service network, which is external to your VPC. To connect privately, you need VPC peering between your VPC and Google's service network.

## Solution / What's New

Added an optional `privateServicesAccess` configuration to `GcpVpcSpec` that, when enabled, automatically creates:

1. **IP Range Allocation**: A reserved IP range for Google managed services
2. **VPC Peering**: A connection between your VPC and `servicenetworking.googleapis.com`

### Architecture Overview

```mermaid
flowchart LR
    subgraph project [GCP Project: pure-lantern-360309]
        subgraph vpc [Your VPC: odwen-prod-vpc]
            workloads[Your Workloads<br/>GKE, VMs, Cloud Run]
            iprange[Reserved IP Range<br/>e.g., 10.100.0.0/16]
        end
        
        subgraph googleservices [Google Service Network]
            cloudsql[Cloud SQL<br/>PostgreSQL<br/>IP: 10.100.5.12]
            memorystore[Memorystore<br/>Redis]
            filestore[Filestore<br/>NFS]
        end
    end
    
    vpc <-->|VPC Peering<br/>servicenetworking.googleapis.com| googleservices
    workloads -->|Private IP<br/>10.100.5.12:5432| cloudsql
    
    style iprange fill:#e1f5fe
    style cloudsql fill:#c8e6c9
```

### Configuration Model

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: prod-vpc
spec:
  networkName: prod-vpc
  projectId:
    value: my-project
  privateServicesAccess:
    enabled: true
    ipRangePrefixLength: 16  # /16 = 65,536 IPs (default)
```

### IP Range Prefix Length

The `ipRangePrefixLength` determines how many IPs are reserved for Google managed services:

```mermaid
graph TD
    subgraph prefix [IP Range Prefix Options]
        p24["/24 = 256 IPs<br/>Tiny deployments"]
        p20["/20 = 4,096 IPs<br/>Small deployments"]
        p16["/16 = 65,536 IPs<br/>Recommended default"]
        p8["/8 = 16.7M IPs<br/>Enterprise scale"]
    end
    
    p24 --> usecase24[1-2 Cloud SQL instances]
    p20 --> usecase20[Dozens of instances]
    p16 --> usecase16[Hundreds of instances<br/>+ Memorystore + Filestore]
    p8 --> usecase8[Massive multi-tenant platforms]
    
    style p16 fill:#c8e6c9
```

Each managed service instance consumes IPs from this range:
- Cloud SQL (single): 1-2 IPs
- Cloud SQL (HA): 2-4 IPs
- Cloud SQL (read replicas): 1 IP per replica
- Memorystore Redis: 1-2 IPs
- Memorystore Memcached: Multiple (one per node)

## Implementation Details

### Proto Schema Changes

**File**: `apis/dev/planton/provider/gcp/gcpvpc/v1/spec.proto`

Added new message type and field:

```protobuf
// Private Services Access configuration for Google managed services.
message GcpVpcPrivateServicesAccess {
  // Enable Private Services Access (VPC peering with Google's service network).
  bool enabled = 1;

  // IP range prefix length for private services allocation.
  // Default: 16 (/16 = 65,536 IPs). Valid range: 8-24.
  int32 ip_range_prefix_length = 2 [
    (buf.validate.field).ignore = IGNORE_IF_ZERO_VALUE,
    (buf.validate.field).int32 = {
      gte: 8
      lte: 24
    },
    (dev.planton.shared.options.default) = "16"
  ];
}

message GcpVpcSpec {
  // ... existing fields ...
  
  // Private Services Access configuration
  GcpVpcPrivateServicesAccess private_services_access = 5;
}
```

**File**: `apis/dev/planton/provider/gcp/gcpvpc/v1/stack_outputs.proto`

Added new outputs:

```protobuf
message GcpVpcStackOutputs {
  string network_self_link = 1;
  string private_services_ip_range_name = 2;  // NEW
  string private_services_ip_range_cidr = 3;  // NEW
}
```

### Pulumi Module Changes

**File**: `apis/dev/planton/provider/gcp/gcpvpc/v1/iac/pulumi/module/vpc.go`

Added `privateServicesAccess()` function that creates:

1. **Global Address**: IP range allocation with `purpose: VPC_PEERING`
2. **Service Networking Connection**: VPC peering with `servicenetworking.googleapis.com`

```go
func privateServicesAccess(ctx *pulumi.Context, locals *Locals, 
    gcpProvider *gcp.Provider, network *compute.Network) error {
    
    // 1. Allocate IP range for private services
    privateIpAlloc, err := compute.NewGlobalAddress(ctx, "private-services-range",
        &compute.GlobalAddressArgs{
            Name:         pulumi.Sprintf("%s-private-svc", spec.NetworkName),
            Purpose:      pulumi.String("VPC_PEERING"),
            AddressType:  pulumi.String("INTERNAL"),
            PrefixLength: pulumi.Int(prefixLength),
            Network:      network.ID(),
        }, pulumi.Provider(gcpProvider))

    // 2. Create VPC peering with Google's service network
    _, err = servicenetworking.NewConnection(ctx, "private-services-connection",
        &servicenetworking.ConnectionArgs{
            Network:               network.ID(),
            Service:               pulumi.String("servicenetworking.googleapis.com"),
            ReservedPeeringRanges: pulumi.StringArray{privateIpAlloc.Name},
        }, pulumi.Provider(gcpProvider))

    return nil
}
```

### Terraform Module Changes

**File**: `apis/dev/planton/provider/gcp/gcpvpc/v1/iac/tf/main.tf`

Added equivalent Terraform resources:

```hcl
# Allocate IP range for private services
resource "google_compute_global_address" "private_services_range" {
  count         = local.enable_private_services ? 1 : 0
  name          = "${var.spec.network_name}-private-svc"
  purpose       = "VPC_PEERING"
  address_type  = "INTERNAL"
  prefix_length = local.private_services_prefix_length
  network       = google_compute_network.vpc.id
}

# Create VPC peering with Google's service network
resource "google_service_networking_connection" "private_services" {
  count                   = local.enable_private_services ? 1 : 0
  network                 = google_compute_network.vpc.id
  service                 = "servicenetworking.googleapis.com"
  reserved_peering_ranges = [google_compute_global_address.private_services_range[0].name]
}
```

### Key Design Decision: No API Enablement at VPC Level

We explicitly chose **not** to enable `servicenetworking.googleapis.com` at the VPC level. This is a critical safety decision:

```mermaid
flowchart TD
    subgraph danger [Dangerous Pattern - AVOIDED]
        vpc1[VPC A enables API]
        vpc2[VPC B enables API]
        destroy[VPC A destroyed]
        disabled[API disabled for project]
        broken[VPC B broken!]
        
        vpc1 --> destroy
        vpc2 --> destroy
        destroy --> disabled
        disabled --> broken
    end
    
    subgraph safe [Safe Pattern - IMPLEMENTED]
        project[GcpProject enables API]
        vpcA[VPC A uses API]
        vpcB[VPC B uses API]
        destroyA[VPC A destroyed]
        stillworks[VPC B still works]
        
        project --> vpcA
        project --> vpcB
        vpcA --> destroyA
        destroyA --> stillworks
    end
    
    style broken fill:#ffcdd2
    style stillworks fill:#c8e6c9
```

**Reason**: In Terraform/Pulumi, when you enable an API via `google_project_service`, destroying that resource can disable the API. If multiple VPCs share the same project, destroying one VPC could break all others.

**Solution**: The VPC module assumes `servicenetworking.googleapis.com` is already enabled at the GcpProject level.

## Benefits

### For Developers
- **One-click private networking**: Enable `privateServicesAccess.enabled: true` and it just works
- **Sensible defaults**: `/16` prefix provides 65K IPs without requiring thought
- **Clear documentation**: Example 6 in `examples.md` shows the complete pattern

### For Operations
- **Consistent deployments**: Same configuration works across Pulumi and Terraform
- **Auditable outputs**: Stack outputs include IP range name and CIDR for reference
- **Safe design**: No risk of accidentally disabling APIs across VPCs

### For Security
- **Private by design**: Managed services never need public IP exposure
- **Network isolation**: All traffic stays within Google's network backbone
- **Reduced attack surface**: No public endpoints for databases

## Required IAM Roles

Deploying Private Services Access requires these roles on the service account:

| Role | Purpose |
|------|---------|
| `roles/compute.networkAdmin` | Create global addresses, use networks, set labels |
| `roles/servicenetworking.networksAdmin` | Create VPC peering with Google's service network |

## Usage Flow

```mermaid
sequenceDiagram
    participant User
    participant GcpProject
    participant GcpVpc
    participant GcpCloudSql
    participant GoogleServices
    
    User->>GcpProject: Enable servicenetworking.googleapis.com
    GcpProject-->>User: API enabled
    
    User->>GcpVpc: Create VPC with privateServicesAccess.enabled
    GcpVpc->>GoogleServices: Allocate IP range (VPC_PEERING)
    GcpVpc->>GoogleServices: Create service networking connection
    GoogleServices-->>GcpVpc: VPC peering established
    GcpVpc-->>User: VPC ready with private services access
    
    User->>GcpCloudSql: Create Cloud SQL with privateIpEnabled
    GcpCloudSql->>GoogleServices: Request private IP from reserved range
    GoogleServices-->>GcpCloudSql: Assigned 10.100.5.12
    GcpCloudSql-->>User: PostgreSQL ready at private IP
    
    Note over User,GoogleServices: App connects to 10.100.5.12:5432
```

## Impact

### Who's Affected
- **GcpVpc users**: New optional feature, no breaking changes
- **GcpCloudSql users**: Can now use `privateIpEnabled: true` without manual VPC peering setup
- **AWS migration users**: Clearer path to private database connectivity on GCP

### Files Changed

| File | Change Type |
|------|-------------|
| `apis/.../gcpvpc/v1/spec.proto` | Added `GcpVpcPrivateServicesAccess` message |
| `apis/.../gcpvpc/v1/stack_outputs.proto` | Added output fields |
| `apis/.../gcpvpc/v1/iac/pulumi/module/vpc.go` | Added `privateServicesAccess()` |
| `apis/.../gcpvpc/v1/iac/pulumi/module/outputs.go` | Added output constants |
| `apis/.../gcpvpc/v1/iac/tf/main.tf` | Added Private Services Access resources |
| `apis/.../gcpvpc/v1/iac/tf/locals.tf` | Added local variables |
| `apis/.../gcpvpc/v1/iac/tf/outputs.tf` | Added new outputs |
| `apis/.../gcpvpc/v1/examples.md` | Added Example 6 |

## Related Work

- **GcpCloudSql**: Uses VPC with Private Services Access for `privateIpEnabled: true`
- **GcpProject**: Should enable `servicenetworking.googleapis.com` as prerequisite
- **Future**: Consider adding automatic prerequisite validation

## Example: Complete Private Database Setup

### Step 1: GcpProject (Enable API)
```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpProject
spec:
  enabledApis:
    - compute.googleapis.com
    - servicenetworking.googleapis.com
```

### Step 2: GcpVpc (Enable Private Services Access)
```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpVpc
metadata:
  name: prod-vpc
spec:
  networkName: prod-vpc
  projectId:
    value: my-project
  privateServicesAccess:
    enabled: true
    ipRangePrefixLength: 16
```

### Step 3: GcpCloudSql (Use Private IP)
```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpCloudSql
spec:
  databaseEngine: POSTGRESQL
  databaseVersion: POSTGRES_15
  network:
    privateIpEnabled: true
    vpcId:
      ref:
        kind: GcpVpc
        name: prod-vpc
```

---

**Status**: ✅ Production Ready
**Timeline**: ~3 hours (research, implementation, testing, documentation)

