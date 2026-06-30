# GCP Compute Instance Pulumi Module

This Pulumi module deploys Google Compute Engine instances using the Planton framework.

## Overview

The module creates a Compute Engine VM instance with configurable:
- Machine type and zone
- Boot disk (image, size, type)
- Network interfaces with optional external IPs
- Service account and OAuth scopes
- Scheduling options (Spot/Preemptible VMs)
- Labels, tags, and metadata
- Startup scripts
- Additional attached disks

## Usage

### With Planton CLI

```bash
# Deploy using Pulumi
planton pulumi up --manifest manifest.yaml --stack org/project/env

# Preview changes
planton pulumi preview --manifest manifest.yaml --stack org/project/env

# Destroy resources
planton pulumi destroy --manifest manifest.yaml --stack org/project/env
```

### Standalone Usage

1. Set the stack input as an environment variable:

```bash
export STACK_INPUT=$(cat <<'EOF'
{
  "target": {
    "apiVersion": "gcp.planton.dev/v1",
    "kind": "GcpComputeInstance",
    "metadata": {
      "name": "my-vm"
    },
    "spec": {
      "projectId": {"value": "my-gcp-project"},
      "zone": "us-central1-a",
      "machineType": "e2-medium",
      "bootDisk": {
        "image": "debian-cloud/debian-11"
      },
      "networkInterfaces": [
        {"network": {"value": "default"}}
      ]
    }
  },
  "providerConfig": {
    "gcpCredentialJson": "..."
  }
}
EOF
)
```

2. Run Pulumi:

```bash
pulumi up --stack my-stack
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `STACK_INPUT` | JSON-encoded GcpComputeInstanceStackInput |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to GCP service account key (alternative to providerConfig) |

## Module Structure

```
.
├── main.go          # Pulumi program entry point
├── Pulumi.yaml      # Pulumi project configuration
├── Makefile         # Build automation
├── debug.sh         # Debugging helper script
├── README.md        # This file
├── overview.md      # Architecture overview
└── module/
    ├── main.go      # Module entry point, provider setup
    ├── locals.go    # Data transformations and labels
    ├── outputs.go   # Export constants
    └── instance.go  # Instance resource creation
```

## Outputs

| Output | Description |
|--------|-------------|
| `instance_name` | Name of the created instance |
| `instance_id` | Unique instance ID |
| `self_link` | Full self-link URL |
| `internal_ip` | Internal (private) IP address |
| `external_ip` | External (public) IP address (if configured) |
| `zone` | Zone where instance is deployed |
| `machine_type` | Machine type of the instance |
| `cpu_platform` | CPU platform of the instance |

## Debugging

For debugging with Delve:

1. Uncomment the `binary` option in `Pulumi.yaml`
2. Run `pulumi up`
3. Attach your debugger to port 2345

## Dependencies

- Pulumi GCP Provider v9+
- Go 1.21+
- Planton SDK

## Development

```bash
# Build the module
make build

# Update dependencies
make update-deps

# Format code
make fmt

# Run vet
make vet
```
