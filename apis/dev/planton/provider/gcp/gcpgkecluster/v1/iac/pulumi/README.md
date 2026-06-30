# Pulumi Implementation - GCP GKE Cluster

This directory contains the Pulumi implementation (in Go) for the **GcpGkeCluster** resource.

## Overview

The Pulumi implementation provides a programmatic way to deploy production-ready GKE clusters using Go. It leverages the Pulumi GCP provider to create private, VPC-native clusters with Workload Identity, network policies, and auto-upgrades.

## Directory Structure

```
pulumi/
├── main.go              # Entry point for Pulumi program
├── module/              # Core implementation
│   ├── main.go          # Resources function and provider setup
│   ├── cluster.go       # GKE cluster creation logic
│   ├── locals.go        # Local variables and initialization
│   └── outputs.go       # Output constants
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Helper commands
├── debug.sh             # Debug script
└── README.md            # This file
```

## Module Components

### main.go (Entry Point)

The entry point loads the stack input and calls the module's `Resources` function:

```go
func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        stackInput := &gcpgkeclusterv1.GcpGkeClusterStackInput{}
        if err := stackinput.Load(stackInput); err != nil {
            return err
        }
        return module.Resources(ctx, stackInput)
    })
}
```

### module/main.go (Provider Setup)

Sets up the GCP provider with credentials and calls the cluster creation logic:

- Initializes local variables from stack input
- Configures GCP provider from credentials
- Calls `cluster` function to create GKE cluster
- Handles errors and wraps them with context

### module/cluster.go (Core Logic)

Implements GKE cluster creation with production-ready defaults:

- **Private Cluster Configuration**:
  - Private nodes (inverted from `enable_public_nodes` flag)
  - Private endpoint access from VPC
  - Master CIDR block for control plane

- **VPC-Native Networking**:
  - IP allocation policy with secondary ranges for pods and services
  - Subnetwork reference

- **Workload Identity**:
  - Conditional configuration based on `disable_workload_identity` flag
  - Workload pool format: `PROJECT_ID.svc.id.goog`

- **Network Policy**:
  - Add-ons configuration for Calico
  - Controlled by `disable_network_policy` flag

- **Release Channel**:
  - Maps proto enum to GKE channel string (RAPID, REGULAR, STABLE, UNSPECIFIED)

- **Cluster Configuration**:
  - Removes default node pool (best practice)
  - Initial node count of 1 (deleted after cluster creation)
  - Deletion protection disabled for IaC-managed lifecycle

### module/locals.go (Local Variables)

Initializes local variables from stack input:

```go
type Locals struct {
    GcpGkeCluster     *gcpgkeclusterv1.GcpGkeCluster
    ReleaseChannelStr string
}
```

Maps the proto enum to GKE release channel strings:
- `RAPID` (1) → "RAPID"
- `REGULAR` (2) → "REGULAR" (default)
- `STABLE` (3) → "STABLE"
- `NONE` (4) → "UNSPECIFIED" (manual upgrades)

### module/outputs.go (Outputs)

Defines output constants for exported values:

- `endpoint` - Cluster API server endpoint
- `cluster_ca_certificate` - CA certificate for cluster verification
- `workload_identity_pool` - Workload Identity pool (if enabled)

## Prerequisites

- Go 1.21 or later
- Pulumi CLI installed
- GCP account with appropriate permissions
- VPC network with subnets and secondary IP ranges
- Cloud NAT for private node egress

## Required GCP Permissions

The service account needs the following IAM roles:

- `roles/container.admin` (to create GKE clusters)
- `roles/compute.networkUser` (to use VPC subnets)
- `roles/iam.serviceAccountUser` (to use service accounts)

## Stack Input Structure

The Pulumi program expects a `GcpGkeClusterStackInput` with:

- `target`: The `GcpGkeCluster` resource definition
- `provider_config`: GCP provider configuration (credentials, project)

Example stack input:

```json
{
  "target": {
    "api_version": "gcp.planton.dev/v1",
    "kind": "GcpGkeCluster",
    "metadata": {
      "name": "prod-cluster",
      "id": "gke-abc123",
      "org": "acme-corp",
      "env": {
        "id": "production"
      }
    },
    "spec": {
      "project_id": { "value": "my-project" },
      "location": "us-central1",
      "subnetwork_self_link": { "value": "https://..." },
      "cluster_secondary_range_name": { "value": "pods" },
      "services_secondary_range_name": { "value": "services" },
      "master_ipv4_cidr_block": "172.16.0.0/28",
      "enable_public_nodes": false,
      "release_channel": 2,
      "disable_network_policy": false,
      "disable_workload_identity": false,
      "router_nat_name": { "value": "my-nat" }
    }
  },
  "provider_config": {
    "credential": "base64-encoded-service-account-key"
  }
}
```

## Outputs

After successful deployment, the following outputs are available:

- **endpoint**: The IP address of the cluster's Kubernetes API server
- **cluster_ca_certificate**: Base64 encoded certificate for verifying the cluster's CA
- **workload_identity_pool**: The Workload Identity pool (format: `PROJECT_ID.svc.id.goog`)

## Usage

### Local Development

1. Set up stack input file (`stack-input.json`)
2. Run Pulumi:

```bash
# Preview changes
pulumi preview --stack dev

# Deploy
pulumi up --stack dev

# Destroy
pulumi destroy --stack dev
```

### Debug Mode

Use the debug script to run with verbose logging:

```bash
./debug.sh
```

This sets `PULUMI_LOG_LEVEL=debug` and runs `pulumi preview`.

### Makefile Commands

```bash
# Install dependencies
make install

# Preview changes
make preview

# Deploy
make up

# Destroy
make destroy

# Clean up
make clean
```

## Implementation Details

### Private Cluster Logic

The `enable_public_nodes` flag is **inverted** when configuring private nodes:

```go
EnablePrivateNodes: pulumi.Bool(!locals.GcpGkeCluster.Spec.EnablePublicNodes)
```

- If `enable_public_nodes = false` (default): Nodes are private (no public IPs)
- If `enable_public_nodes = true`: Nodes get public IPs (legacy pattern)

### Workload Identity Configuration

Workload Identity is conditionally enabled:

```go
if !locals.GcpGkeCluster.Spec.DisableWorkloadIdentity {
    workloadIdentityCfg = container.ClusterWorkloadIdentityConfigPtrInput(
        &container.ClusterWorkloadIdentityConfigArgs{
            WorkloadPool: pulumi.Sprintf("%s.svc.id.goog",
                locals.GcpGkeCluster.Spec.ProjectId.GetValue()),
        })
}
```

The workload pool format is: `PROJECT_ID.svc.id.goog`

### Release Channel Mapping

The proto enum is mapped to GKE channel strings in `locals.go`:

```go
switch l.GcpGkeCluster.Spec.GetReleaseChannel() {
case gcpgkeclusterv1.GkeReleaseChannel_RAPID:
    l.ReleaseChannelStr = "RAPID"
case gcpgkeclusterv1.GkeReleaseChannel_REGULAR:
    l.ReleaseChannelStr = "REGULAR"
case gcpgkeclusterv1.GkeReleaseChannel_STABLE:
    l.ReleaseChannelStr = "STABLE"
case gcpgkeclusterv1.GkeReleaseChannel_NONE:
    l.ReleaseChannelStr = "UNSPECIFIED"
default:
    l.ReleaseChannelStr = "REGULAR"
}
```

### Network Policy Configuration

Calico network policy enforcement is controlled by the `disable_network_policy` flag:

```go
AddonsConfig: addonsCfg := container.ClusterAddonsConfigPtrInput(&container.ClusterAddonsConfigArgs{
    NetworkPolicyConfig: container.ClusterAddonsConfigNetworkPolicyConfigPtrInput(
        &container.ClusterAddonsConfigNetworkPolicyConfigArgs{
            Disabled: pulumi.Bool(locals.GcpGkeCluster.Spec.DisableNetworkPolicy),
        }),
})
```

## Design Decisions

### Separate Node Pools

The cluster is created with `RemoveDefaultNodePool: true` and `InitialNodeCount: 1`. This follows GKE best practices:

- Default node pools have generic configurations
- Production workloads need customized node pools (separate `GcpGkeNodePool` resources)
- Prevents unintended node pool updates from triggering cluster-level operations

### Deletion Protection Disabled

`DeletionProtection: false` allows IaC tools to destroy clusters. For production:

- Use CI/CD approval gates
- Enable deletion protection manually via console/gcloud if needed
- Prefer automated, repeatable cluster creation over long-lived pet clusters

### Master IPv4 CIDR Validation

The `/28` CIDR validation happens in the proto spec (`spec.proto`), not in Pulumi code. This ensures:

- Early validation before deployment
- Consistent validation across Pulumi and Terraform
- Clear error messages at manifest definition time

## Troubleshooting

### Error: "Invalid value for master_ipv4_cidr_block"

- Must be a `/28` CIDR block (exactly 16 IPs)
- Must be RFC1918 private range
- Cannot overlap with VPC, pod, or service ranges

### Error: "Workload Identity Pool not created"

- Check that `disable_workload_identity` is `false`
- Verify the project ID is correct
- Ensure Workload Identity is enabled on the project

### Error: "Cluster cannot be created in network"

- Verify the subnetwork exists and is in the correct project
- Check that secondary IP ranges exist on the subnet
- Ensure the service account has `roles/compute.networkUser`

## Next Steps

After deploying a GKE cluster with Pulumi:

1. Configure `kubectl` to access the cluster:
   ```bash
   gcloud container clusters get-credentials CLUSTER_NAME \
     --region=REGION \
     --project=PROJECT_ID
   ```

2. Deploy node pools using `GcpGkeNodePool` resources

3. Set up Workload Identity bindings for your applications

4. Deploy Kubernetes workloads

For comprehensive design rationale and production best practices, see the [research documentation](../../docs/README.md).

