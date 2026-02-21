# HetznerCloudLoadBalancer Terraform Module

Terraform IaC module for provisioning a Hetzner Cloud load balancer with services, targets (server, label selector, IP), health checks, TLS termination, and optional private network attachment.

## Structure

```
.
├── main.tf           # Load balancer, services (for_each), targets (for_each x3), network attachment (count)
├── outputs.tf        # Stack output definitions
├── variables.tf      # Input variable definitions (metadata, spec, hcloud_token)
├── locals.tf         # LB name, standard labels, algorithm default, service port computation
└── provider.tf       # HetznerCloud provider configuration
```

## Resources Created

- `hcloud_load_balancer` — Provisions a load balancer with the specified type, location, algorithm, labels, and delete protection
- `hcloud_load_balancer_service` (via `for_each`, keyed by effective listen port) — Configures a listener with protocol, ports, dynamic `http` block (sticky sessions, certificates, redirect), and dynamic `health_check` block
- `hcloud_load_balancer_target.server` (via `for_each`, keyed by `server_id`) — Adds server backends
- `hcloud_load_balancer_target.label_selector` (via `for_each`, keyed by `selector`) — Adds label-selector-based dynamic backends
- `hcloud_load_balancer_target.ip` (via `for_each`, keyed by `ip`) — Adds external IP backends
- `hcloud_load_balancer_network` (conditional, via `count`) — Attaches the load balancer to a private network, created only when `var.spec.network` is non-null

## Outputs

| Name | Description |
|------|-------------|
| `load_balancer_id` | Hetzner Cloud numeric ID of the created load balancer |
| `ipv4_address` | Public IPv4 address assigned to the load balancer |
| `ipv6_address` | Public IPv6 address assigned to the load balancer |

## Usage

```bash
# Initialize
terraform init

# Plan
terraform plan \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"web-lb"}' \
  -var 'spec={"load_balancer_type":"lb11","location":"fsn1","services":[{"protocol":"http"}],"server_targets":[{"server_id":"12345"}]}'

# Apply
terraform apply \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"web-lb"}' \
  -var 'spec={"load_balancer_type":"lb11","location":"fsn1","services":[{"protocol":"http"}],"server_targets":[{"server_id":"12345"}]}'
```

For structured input, use a `.tfvars` file:

```bash
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```
