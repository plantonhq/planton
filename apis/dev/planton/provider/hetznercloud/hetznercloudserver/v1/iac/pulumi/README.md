# HetznerCloudServer Pulumi Module

Pulumi (Go) IaC module for provisioning Hetzner Cloud servers with SSH key injection, firewall attachment, placement group scheduling, public and private networking, cloud-init, backups, protections, and optional reverse DNS.

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Data extraction from stack input, label computation
│   ├── server.go     # Server creation, public net, network attachments, conditional rDNS; output exports
│   └── outputs.go    # Output name constants
└── Pulumi.yaml       # Project configuration
```

## Resources Created

- `hcloud.Server` — Provisions a server with the specified type, image, location, SSH keys, firewall IDs, placement group, public networking configuration, private network attachments, cloud-init, backup settings, protections, and labels
- `hcloud.Rdns` (conditional) — Reverse DNS pointer record for the server's auto-assigned public IPv4 address, created only when `dnsPtr` is non-empty in the spec

## Outputs

| Name | Description |
|------|-------------|
| `server_id` | Hetzner Cloud numeric ID of the created server |
| `ipv4_address` | Public IPv4 address assigned to the server |
| `ipv6_address` | First IPv6 address of the assigned /64 network |
| `status` | Current server status (running, off, rebuilding, migrating) |

## Usage

```bash
# Build
bazel build //apis/dev/planton/provider/hetznercloud/hetznercloudserver/v1/iac/pulumi:pulumi

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
