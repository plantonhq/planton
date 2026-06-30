# GCP VPC - Pulumi Module

## Overview

This directory contains the Pulumi implementation for deploying GCP VPC networks using Planton's `GcpVpc` API. The module is written in Go and leverages the Pulumi GCP provider to create production-ready Virtual Private Cloud networks with best-practice defaults.

## Prerequisites

Before deploying, ensure you have:

1. **Pulumi CLI** installed (version 3.x or later)
   ```bash
   curl -fsSL https://get.pulumi.com | sh
   ```

2. **Go** installed (version 1.21 or later)
   ```bash
   go version  # Should show 1.21+
   ```

3. **GCP Project** with billing enabled
   - You need a GCP project where the VPC will be created
   - Ensure you have Owner or Editor role on the project

4. **GCP Credentials** configured
   ```bash
   gcloud auth application-default login
   gcloud config set project <your-project-id>
   ```

5. **Compute Engine API** enabled (the module will enable it automatically, but you can do it manually):
   ```bash
   gcloud services enable compute.googleapis.com
   ```

## Directory Structure

```
iac/pulumi/
├── main.go           # Pulumi program entrypoint
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment targets
├── debug.sh          # Debug helper script
├── README.md         # This file
├── overview.md       # Architecture overview
└── module/
    ├── main.go       # Module initialization and coordination
    ├── vpc.go        # VPC resource creation logic
    ├── locals.go     # Local values and labels
    └── outputs.go    # Stack outputs
```

## Quick Start

### 1. Initialize Pulumi Stack

```bash
cd iac/pulumi
pulumi stack init dev
```

### 2. Configure Stack

Set your GCP project and region:

```bash
pulumi config set gcp:project <your-gcp-project-id>
pulumi config set gcp:region us-west1  # Optional, for provider config
```

### 3. Create Input File

Create a `stack-input.yaml` file with your VPC specification:

```yaml
target:
  apiVersion: gcp.planton.dev/v1
  kind: GcpVpc
  metadata:
    name: my-vpc
    id: gcpvpc-12345
    org: my-org
    env: dev
  spec:
    project_id:
      value: my-gcp-project-123
    auto_create_subnetworks: false
    routing_mode: REGIONAL

providerConfig:
  gcpCredential:
    value: <base64-encoded-service-account-key>
```

### 4. Deploy

```bash
pulumi up
```

Review the preview and confirm the deployment.

### 5. View Outputs

After deployment, view the VPC self-link:

```bash
pulumi stack output network_self_link
```

Example output:
```
projects/my-gcp-project-123/global/networks/my-vpc
```

## Makefile Targets

The `Makefile` provides convenient commands for common operations:

```bash
# Download and tidy dependencies
make deps

# Run go vet (static analysis)
make vet

# Format code
make fmt

# Build (runs deps, vet, fmt)
make build

# Update Planton dependencies to latest
make update-deps
```

## Module Architecture

### Initialization Flow

1. **`main.go`**: Entry point that reads `stack-input.yaml` and calls the module
2. **`module/main.go`**: Initializes locals and orchestrates resource creation
3. **`module/locals.go`**: Computes labels, project ID, and routing mode from input
4. **`module/vpc.go`**: Creates the VPC network resource with proper dependencies
5. **`module/outputs.go`**: Exports stack outputs

### Key Components

#### Locals (`locals.go`)

Computes frequently-used values:
- **GCP Labels**: Converts metadata (name, org, env, resource kind) to GCP label format
- **Provider Config**: Extracts GCP credentials and project configuration

#### VPC Creation (`vpc.go`)

1. **Enable Compute API**: Ensures `compute.googleapis.com` is enabled in the target project
2. **Create VPC Network**: Provisions `google_compute_network` with:
   - Name from `metadata.name`
   - Project from `spec.project_id`
   - Auto-create-subnetworks setting from `spec.auto_create_subnetworks`
   - Routing mode (REGIONAL or GLOBAL) from `spec.routing_mode`
3. **Export Outputs**: Exports `network_self_link` for use by other resources

#### Outputs (`outputs.go`)

Defines constants for output keys:
- `network_self_link`: Full self-link URL of the VPC (format: `projects/<project>/global/networks/<name>`)

## Advanced Usage

### Using with Pulumi Automation API

You can embed this module in a larger Pulumi program:

```go
import (
    gcpvpcv1 "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpvpc/v1"
    "github.com/plantonhq/planton/apis/dev/planton/provider/gcp/gcpvpc/v1/iac/pulumi/module"
)

func deployVpc(ctx *pulumi.Context, stackInput *gcpvpcv1.GcpVpcStackInput) error {
    return module.Resources(ctx, stackInput)
}
```

### Debugging

Use the `debug.sh` script to run the Pulumi program locally with debug output:

```bash
./debug.sh
```

This sets `PULUMI_CONFIG_PASSPHRASE=password` and runs `pulumi up` with verbose logging.

### Customizing the Module

If you need to extend the module:

1. **Add new fields to `spec.proto`** at the API level
2. **Update `vpc.go`** to handle the new fields
3. **Regenerate Go stubs**: Run `make protos` in the workspace root
4. **Test locally**: Use `pulumi preview` to validate changes

## Deployment Scenarios

### Development Environment

```yaml
target:
  metadata:
    name: dev-vpc
    env: development
  spec:
    project_id:
      value: dev-project-123
    auto_create_subnetworks: false
```

### Production with Global Routing

```yaml
target:
  metadata:
    name: prod-vpc
    env: production
  spec:
    project_id:
      value: prod-project-456
    auto_create_subnetworks: false
    routing_mode: GLOBAL
```

### Referencing Another Resource

```yaml
target:
  spec:
    project_id:
      ref:
        kind: GcpProject
        name: my-project
```

Pulumi will automatically resolve the `project_id` from the referenced `GcpProject` resource.

## Outputs

After successful deployment, the stack exports:

| Output Key | Type | Description |
|------------|------|-------------|
| `network_self_link` | string | Full GCP self-link URL of the VPC network |

### Using Outputs in Other Pulumi Stacks

```go
vpcSelfLink := ctx.Export("network_self_link", createdNetwork.SelfLink)

// In another stack
vpcRef := pulumi.StackReference(...)
vpcSelfLink := vpcRef.GetOutput(pulumi.String("network_self_link"))
```

## Troubleshooting

### "API compute.googleapis.com has not been used"

**Cause**: Compute API is not enabled in the target project.

**Solution**: The module automatically enables it. If this fails, enable it manually:
```bash
gcloud services enable compute.googleapis.com --project=<your-project>
```

### "Resource already exists"

**Cause**: A VPC with the same name already exists in the project.

**Solution**:
1. Import existing VPC into Pulumi state:
   ```bash
   pulumi import google-native:compute/v1:Network vpc projects/<project>/global/networks/<name>
   ```
2. Or delete the existing VPC:
   ```bash
   gcloud compute networks delete <vpc-name> --project=<project>
   ```

### "Permission denied"

**Cause**: GCP credentials lack sufficient permissions.

**Solution**: Ensure your service account or user has `roles/compute.networkAdmin` or `roles/editor`:
```bash
gcloud projects add-iam-policy-binding <project-id> \
  --member=user:<your-email> \
  --role=roles/compute.networkAdmin
```

### State File Conflicts

**Cause**: Multiple Pulumi processes trying to update state simultaneously.

**Solution**: Pulumi uses state locking. If locked, wait or run:
```bash
pulumi cancel  # Cancel the operation holding the lock
```

## Best Practices

### 1. Use Remote State

Configure Pulumi to use remote state storage:

```bash
# Google Cloud Storage backend
pulumi login gs://my-pulumi-state-bucket
```

### 2. Separate Stacks Per Environment

```bash
pulumi stack init dev
pulumi stack init staging
pulumi stack init prod
```

Each stack maintains separate state and can have different configurations.

### 3. Version Pin Dependencies

In `go.mod`, pin Pulumi and GCP provider versions to avoid unexpected updates:

```go
require (
    github.com/pulumi/pulumi-gcp/sdk/v9 v8.0.0
    github.com/pulumi/pulumi/sdk/v3 v3.100.0
)
```

### 4. Label Resources

Always set `metadata.org` and `metadata.env` for cost tracking and organization:

```yaml
metadata:
  org: platform-team
  env: production
```

These are automatically converted to GCP labels by the module.

### 5. Plan IP Ranges

Before deploying, plan non-overlapping IP ranges for subnets across environments:
- Dev: `10.10.0.0/16`
- Staging: `10.20.0.0/16`
- Prod: `10.30.0.0/16`

## Related Resources

After deploying the VPC, you typically need:

1. **Subnets**: Create via `GcpSubnetwork` resources
2. **Firewall Rules**: Manage via `GcpFirewallRule` resources
3. **Cloud NAT**: For private instance outbound connectivity
4. **Private Google Access**: Enable on subnets for Google API access without public IPs

## Further Reading

- [Pulumi GCP Provider Documentation](https://www.pulumi.com/registry/packages/gcp/)
- [Planton Architecture Overview](../../docs/README.md)
- [Component API Reference](../../README.md)
- [Pulumi Best Practices](https://www.pulumi.com/docs/guides/best-practices/)

## Support

For issues or questions:
1. Check [troubleshooting section](#troubleshooting)
2. Review [examples](../../examples.md)
3. Consult [research documentation](../../docs/README.md)
4. Open an issue in the Planton repository

