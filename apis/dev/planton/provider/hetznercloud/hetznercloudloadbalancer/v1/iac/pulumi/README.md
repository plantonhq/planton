# HetznerCloudLoadBalancer Pulumi Module

Pulumi (Go) IaC module for provisioning a Hetzner Cloud load balancer with services, targets (server, label selector, IP), health checks, TLS termination, and optional private network attachment.

## Structure

```
.
├── main.go           # Entry point: loads stack input, calls module.Resources
├── module/
│   ├── main.go       # Provider setup and orchestration
│   ├── locals.go     # Data extraction from stack input, label computation
│   ├── load_balancer.go  # LB creation, services, targets, network attachment, health checks; output exports
│   └── outputs.go    # Output name constants
└── Pulumi.yaml       # Project configuration
```

## Resources Created

- `hcloud.LoadBalancer` — Provisions a load balancer with the specified type, location, algorithm, labels, and delete protection
- `hcloud.LoadBalancerService` (per service) — Configures a listener with protocol, ports, HTTP settings (sticky sessions, certificates, redirect), and health check
- `hcloud.LoadBalancerTarget` (per target) — Adds a backend target: server (by ID), label selector (dynamic), or external IP
- `hcloud.LoadBalancerNetwork` (conditional) — Attaches the load balancer to a private network, created only when `network` is set in the spec

## Outputs

| Name | Description |
|------|-------------|
| `load_balancer_id` | Hetzner Cloud numeric ID of the created load balancer |
| `ipv4_address` | Public IPv4 address assigned to the load balancer |
| `ipv6_address` | Public IPv6 address assigned to the load balancer |

## Usage

```bash
# Build
bazel build //apis/dev/planton/provider/hetznercloud/hetznercloudloadbalancer/v1/iac/pulumi:pulumi

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
