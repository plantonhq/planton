# HetznerCloudNetwork Pulumi Module

Pulumi (Go) IaC module for creating private networks with subnets and static routes in Hetzner Cloud.

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Data extraction from stack input, label computation
│   ├── network.go    # Network, subnet, and route creation; output exports
│   └── outputs.go    # Output name constants
├── Pulumi.yaml       # Project configuration
└── BUILD.bazel       # Bazel build definition
```

## Resources Created

- `hcloud.Network` — Network with top-level CIDR, labels, and protection settings
- `hcloud.NetworkSubnet` (1 per subnet) — Subnets within the network, keyed by CIDR
- `hcloud.NetworkRoute` (1 per route, optional) — Static routes for custom traffic paths

## Outputs

| Name | Description |
|------|-------------|
| `network_id` | Hetzner Cloud numeric ID of the created network |

## Usage

```bash
# Build
bazel build //apis/org/openmcf/provider/hetznercloud/hetznercloudnetwork/v1/iac/pulumi:pulumi

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
