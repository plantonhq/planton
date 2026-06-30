# HetznerCloudSnapshot Pulumi Module

Pulumi (Go) IaC module for creating Hetzner Cloud server snapshots stored as Images.

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Data extraction from stack input, label computation
│   ├── snapshot.go   # Snapshot creation, server ID conversion, output export
│   └── outputs.go    # Output name constants
└── BUILD.bazel       # Bazel build configuration
```

## Resources Created

- `hcloud.Snapshot` — Creates a server snapshot stored as a Hetzner Cloud Image. Captures the full disk of the source server at the moment of creation.

## Outputs

| Name | Description |
|------|-------------|
| `snapshot_id` | Hetzner Cloud image ID of the created snapshot |

## Usage

```bash
# Build
bazel build //apis/dev/planton/provider/hetznercloud/hetznercloudsnapshot/v1/iac/pulumi:pulumi

# Test with local manifest
export STACK_INPUT=$(cat ../hack/manifest.yaml | base64)
pulumi up
```

## Debug

```bash
# Run locally against the hack manifest
export STACK_INPUT=$(cat ../hack/manifest.yaml | base64)
export HCLOUD_TOKEN="your-api-token"
pulumi up --stack dev
```
