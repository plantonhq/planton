# Pulumi Module for GCP Subnetwork

This directory contains the Pulumi implementation for deploying GCP custom-mode subnetworks.

## Architecture

The module follows a modular structure:

```
iac/pulumi/
├── main.go           # Entry point: loads stack input and invokes module
├── Pulumi.yaml       # Pulumi project configuration
├── Makefile          # Build and deployment helpers
├── debug.sh          # Local debugging script
└── module/
    ├── main.go       # Module entry point: orchestrates resource creation
    ├── locals.go     # Helper bundle (spec, metadata, labels)
    ├── outputs.go    # Output constants and helpers
    └── subnetwork.go # Core implementation: google_compute_subnetwork resource
```

## Key Components

### main.go (Root)

The entry point that:
1. Loads `GcpSubnetworkStackInput` from Pulumi config
2. Unmarshals protobuf JSON into Go structs
3. Calls `module.Resources()` to create infrastructure
4. Exports stack outputs

### module/main.go

Orchestrates resource creation:
- Creates `Locals` bundle (spec, metadata, labels)
- Configures GCP provider with credentials from `GcpProviderConfig`
- Calls `subnetwork()` to provision the actual subnet
- Returns error if any step fails

### module/subnetwork.go

Core implementation that:
1. **Enables APIs**: Activates `compute.googleapis.com` via `google_project_service`
2. **Prepares Secondary Ranges**: Converts spec to Pulumi `SubnetworkSecondaryIpRangeArray`
3. **Creates Subnet**: Provisions `google_compute_subnetwork` with all spec fields
4. **Exports Outputs**: Registers outputs (self-link, region, CIDR, secondary ranges)

### module/locals.go

Helper bundle containing:
- `GcpSubnetwork`: Full spec from stack input
- Resource labels (resource_id, resource_kind, organization, environment)
- Utility functions for label merging and resource naming

### module/outputs.go

Defines output constants:
- `OpSubnetworkSelfLink`: Self-link URL of created subnet
- `OpRegion`: Region where subnet resides
- `OpIpCidrRange`: Primary IPv4 CIDR
- `OpSecondaryRanges`: List of secondary ranges

## Local Development Workflow

### Prerequisites

- Go 1.21 or later
- Pulumi CLI installed (`brew install pulumi` or see [pulumi.com](https://www.pulumi.com/docs/get-started/install/))
- GCP credentials configured (service account JSON or `gcloud auth application-default login`)
- `planton` CLI installed

### Setup

1. **Navigate to module directory**:

```shell
cd apis/dev/planton/provider/gcp/gcpsubnetwork/v1/iac/pulumi
```

2. **Initialize Pulumi stack** (if not already done):

```shell
pulumi login --local  # Or use Pulumi Service: pulumi login
pulumi stack init dev
```

3. **Prepare a test manifest**:

Create `../../hack/manifest.yaml`:

```yaml
apiVersion: gcp.planton.dev/v1
kind: GcpSubnetwork
metadata:
  name: test-subnet
spec:
  project_id: my-gcp-project
  vpc_self_link: projects/my-gcp-project/global/networks/my-vpc
  region: us-central1
  ip_cidr_range: 10.100.0.0/20
  private_ip_google_access: true
```

### Deploy

#### Method 1: Using Makefile (Recommended)

```shell
# Preview changes
make deploy MANIFEST=../../hack/manifest.yaml

# Apply changes (auto-approve)
make deploy MANIFEST=../../hack/manifest.yaml AUTO_APPROVE=true

# Destroy resources
make destroy MANIFEST=../../hack/manifest.yaml
```

#### Method 2: Direct Pulumi Commands

```shell
# Preview
pulumi preview --config-file <(planton pulumi stack-input --manifest ../../hack/manifest.yaml)

# Deploy
pulumi up --config-file <(planton pulumi stack-input --manifest ../../hack/manifest.yaml) --yes

# Destroy
pulumi destroy --config-file <(planton pulumi stack-input --manifest ../../hack/manifest.yaml) --yes
```

#### Method 3: Using planton CLI (Simplest)

```shell
# From v1/ directory
planton pulumi up --manifest hack/manifest.yaml

# Destroy
planton pulumi destroy --manifest hack/manifest.yaml
```

### Debugging

#### Debug Script

Use the included `debug.sh` for rapid iteration:

```shell
# Set environment variables
export GCP_PROJECT_ID=my-project
export GCP_CREDENTIALS_PATH=/path/to/service-account.json
export MANIFEST_PATH=../../hack/manifest.yaml

# Run debug script
./debug.sh
```

The script:
- Loads manifest using `planton` CLI
- Generates Pulumi config JSON
- Runs `pulumi up` with verbose logging
- Saves outputs to `debug-outputs.json`

#### Manual Debugging

1. **Generate stack input**:

```shell
planton pulumi stack-input --manifest ../../hack/manifest.yaml > stack-input.json
```

2. **Inspect generated config**:

```shell
cat stack-input.json | jq .
```

3. **Run Pulumi with logging**:

```shell
pulumi up --config-file stack-input.json --logtostderr --verbose=9
```

## Testing Changes

### Unit Tests (Go)

```shell
# Run all tests in module
cd module/
go test -v ./...
```

### Integration Tests

1. **Deploy to test environment**:

```shell
make deploy MANIFEST=../../hack/manifest.yaml AUTO_APPROVE=true
```

2. **Verify outputs**:

```shell
pulumi stack output subnetwork_self_link
pulumi stack output region
pulumi stack output ip_cidr_range
```

3. **Validate in GCP Console**:

Navigate to: [VPC Networks → Subnets](https://console.cloud.google.com/networking/subnetworks/list)

4. **Clean up**:

```shell
make destroy MANIFEST=../../hack/manifest.yaml
```

## Common Issues

### "compute API not enabled"

**Cause**: `compute.googleapis.com` API not enabled in project.

**Solution**: The module automatically enables it. Wait 30-60 seconds for propagation, then retry.

### "VPC network not found"

**Cause**: `vpc_self_link` references a non-existent VPC.

**Solution**: Ensure VPC exists before creating subnet. Use `GcpVpc` resource to create it.

### "CIDR range overlaps"

**Cause**: `ip_cidr_range` conflicts with existing subnet in the VPC.

**Solution**: Choose a non-overlapping CIDR range. Check existing subnets:

```shell
gcloud compute networks subnets list --network=my-vpc
```

### "Invalid secondary range name"

**Cause**: Secondary range name violates RFC1035 (must start with letter, lowercase only).

**Solution**: Use names like "pods" or "gke-services" (not "Pods" or "123-range").

## Pulumi State Management

### Local State (Development)

```shell
pulumi login --local
```

State stored in `~/.pulumi/` (not recommended for teams).

### Pulumi Service (Recommended for Teams)

```shell
pulumi login
pulumi stack init dev --secrets-provider=passphrase
```

State stored in Pulumi Cloud with encrypted secrets.

### Self-Hosted State (GCS Backend)

```shell
pulumi login gs://my-pulumi-state-bucket
```

State stored in Google Cloud Storage bucket.

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Deploy GCP Subnetwork
on:
  push:
    branches: [main]
    paths:
      - 'apis/dev/planton/provider/gcp/gcpsubnetwork/**'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Install Pulumi CLI
        run: curl -fsSL https://get.pulumi.com | sh
      - name: Deploy
        env:
          PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}
          GCP_CREDENTIALS: ${{ secrets.GCP_CREDENTIALS }}
        run: |
          cd apis/dev/planton/provider/gcp/gcpsubnetwork/v1/iac/pulumi
          pulumi stack select dev --create
          planton pulumi up --manifest ../../hack/manifest.yaml --yes
```

## Updating the Module

### Modify Spec

1. Update `spec.proto` with new fields
2. Regenerate Go stubs: `make protos` (from repo root)
3. Update `module/subnetwork.go` to handle new fields
4. Test locally with `make deploy`

### Add New Resource

1. Create new file in `module/` (e.g., `firewall.go`)
2. Implement resource creation function
3. Call from `module/main.go`
4. Export outputs in `module/outputs.go`

## Performance Tips

- **Parallel API Enablement**: Use `pulumi.DependsOn()` to parallelize API enablement across multiple resources
- **Conditional Secondaries**: Only create secondary ranges if spec defines them (avoid empty `dynamic` blocks)
- **Label Optimization**: Precompute labels in `locals.go` to avoid repeated merging

## Further Reading

- [Pulumi GCP Provider Docs](https://www.pulumi.com/registry/packages/gcp/)
- [google_compute_subnetwork Resource](https://www.pulumi.com/registry/packages/gcp/api-docs/compute/subnetwork/)
- [Pulumi Programming Model](https://www.pulumi.com/docs/intro/concepts/)
- [Planton CLI Documentation](../../../../../../../../../README.md)

## Related Files

- **Spec Definition**: `../../spec.proto`
- **Stack Input**: `../../stack_input.proto`
- **Stack Outputs**: `../../stack_outputs.proto`
- **Test Manifest**: `../../hack/manifest.yaml`
- **Module Overview**: `overview.md` (architectural deep dive)

---

**Need help?** Check the [module overview](overview.md) or consult the [component README](../../README.md).

