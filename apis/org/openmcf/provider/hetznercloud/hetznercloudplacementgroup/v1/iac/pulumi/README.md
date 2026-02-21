# HetznerCloudPlacementGroup Pulumi Module

Pulumi (Go) IaC module for creating placement groups in Hetzner Cloud.

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go             # Provider setup and orchestration
│   ├── locals.go           # Data extraction from stack input, label computation
│   ├── placement_group.go  # Placement group resource creation and output export
│   └── outputs.go          # Output name constants
├── Pulumi.yaml       # Project configuration
└── BUILD.bazel       # Bazel build definition
```

## Resources Created

- `hcloud.PlacementGroup` — Placement group with the specified strategy (defaults to `spread`)

## Outputs

| Name | Description |
|------|-------------|
| `placement_group_id` | Hetzner Cloud numeric ID of the created placement group |

## Usage

```bash
# Build
bazel build //apis/org/openmcf/provider/hetznercloud/hetznercloudplacementgroup/v1/iac/pulumi:pulumi

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
