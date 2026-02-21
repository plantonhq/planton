# HetznerCloudFirewall Pulumi Module

Pulumi (Go) IaC module for creating firewalls with inline rules in Hetzner Cloud.

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Data extraction from stack input, label computation
│   ├── firewall.go   # Rule mapping, firewall resource creation, output exports
│   └── outputs.go    # Output name constants
├── Pulumi.yaml       # Project configuration
└── BUILD.bazel       # Bazel build definition
```

## Resources Created

- `hcloud.Firewall` — Firewall with inline rules mapped from `spec.rules`

## Outputs

| Name | Description |
|------|-------------|
| `firewall_id` | Hetzner Cloud numeric ID of the created firewall |

## Usage

```bash
# Build
bazel build //apis/org/openmcf/provider/hetznercloud/hetznercloudfirewall/v1/iac/pulumi:pulumi

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
