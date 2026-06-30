# HetznerCloudSshKey Pulumi Module

Pulumi (Go) IaC module for registering SSH public keys in Hetzner Cloud.

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Data extraction from stack input, label computation
│   ├── ssh_key.go    # SSH key resource creation and output exports
│   └── outputs.go    # Output name constants
├── Pulumi.yaml       # Project configuration
└── BUILD.bazel       # Bazel build definition
```

## Resources Created

- `hcloud.SshKey` — SSH public key registered in Hetzner Cloud

## Outputs

| Name | Description |
|------|-------------|
| `ssh_key_id` | Hetzner Cloud numeric ID of the created SSH key |
| `fingerprint` | MD5 fingerprint of the SSH public key |

## Usage

```bash
# Build
bazel build //apis/dev/planton/provider/hetznercloud/hetznercloudsshkey/v1/iac/pulumi:pulumi

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
