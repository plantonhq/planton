# HetznerCloudPrimaryIp Pulumi Module

Pulumi (Go) IaC module for allocating persistent public IP addresses in Hetzner Cloud with optional reverse DNS.

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Data extraction from stack input, label computation
│   ├── primary_ip.go # Primary IP and conditional rDNS creation; output exports
│   └── outputs.go    # Output name constants
├── Pulumi.yaml       # Project configuration
└── BUILD.bazel       # Bazel build definition
```

## Resources Created

- `hcloud.PrimaryIp` — Allocates an IPv4 address or IPv6 /64 block with labels and protection settings
- `hcloud.Rdns` (conditional) — Reverse DNS pointer record, created only when `dnsPtr` is non-empty in the spec

## Outputs

| Name | Description |
|------|-------------|
| `primary_ip_id` | Hetzner Cloud numeric ID of the created Primary IP |
| `ip_address` | The allocated IP address |
| `ip_network` | The allocated IPv6 /64 CIDR (empty for IPv4) |

## Usage

```bash
# Build
bazel build //apis/org/openmcf/provider/hetznercloud/hetznercloudprimaryip/v1/iac/pulumi:pulumi

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
