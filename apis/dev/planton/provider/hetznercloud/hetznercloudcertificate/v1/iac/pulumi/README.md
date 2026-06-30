# HetznerCloudCertificate Pulumi Module

Pulumi (Go) IaC module for provisioning Hetzner Cloud TLS certificates (uploaded or managed via Let's Encrypt).

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Data extraction from stack input, label computation
│   ├── certificate.go # Certificate creation (routes to uploaded or managed)
│   └── outputs.go    # Output name constants
└── BUILD.bazel       # Bazel build configuration
```

## Resources Created

Exactly one of:

- `hcloud.UploadedCertificate` — stores a user-provided PEM certificate chain and private key. Created when the `uploaded` spec variant is set.
- `hcloud.ManagedCertificate` — requests a Let's Encrypt certificate for the specified domains. Created when the `managed` spec variant is set.

## Outputs

| Name | Description |
|------|-------------|
| `certificate_id` | Hetzner Cloud numeric ID of the created certificate |
| `type` | Certificate type: `"uploaded"` or `"managed"` |
| `fingerprint` | SHA256 fingerprint of the certificate |
| `not_valid_before` | Certificate validity start (ISO-8601) |
| `not_valid_after` | Certificate validity end (ISO-8601) |

## Usage

```bash
# Build
bazel build //apis/dev/planton/provider/hetznercloud/hetznercloudcertificate/v1/iac/pulumi:pulumi

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
