# HetznerCloudServer Terraform Module

Terraform IaC module for provisioning Hetzner Cloud servers with SSH key injection, firewall attachment, placement group scheduling, public and private networking, cloud-init, backups, protections, and optional reverse DNS.

## Structure

```
.
├── main.tf           # Server resource with dynamic public_net and network blocks; conditional rDNS
├── outputs.tf        # Stack output definitions
├── variables.tf      # Input variable definitions (metadata, spec, hcloud_token)
├── locals.tf         # Server name derivation and standard label computation
└── provider.tf       # HetznerCloud provider configuration
```

## Resources Created

- `hcloud_server` — Provisions a server with the specified type, image, location, SSH keys, firewall IDs, placement group, public networking (via `dynamic "public_net"`), private network attachments (via `dynamic "network"`), cloud-init, backup settings, protections, and labels
- `hcloud_rdns` (conditional, via `count`) — Reverse DNS pointer record for the server's auto-assigned public IPv4 address, created only when `dns_ptr` is non-empty

## Outputs

| Name | Description |
|------|-------------|
| `server_id` | Hetzner Cloud numeric ID of the created server |
| `ipv4_address` | Public IPv4 address assigned to the server |
| `ipv6_address` | First IPv6 address of the assigned /64 network |
| `status` | Current server status (running, off, rebuilding, migrating) |

## Usage

```bash
# Initialize
terraform init

# Plan
terraform plan \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"web-01"}' \
  -var 'spec={"server_type":"cx22","image":"ubuntu-24.04","location":"fsn1"}'

# Apply
terraform apply \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"web-01"}' \
  -var 'spec={"server_type":"cx22","image":"ubuntu-24.04","location":"fsn1"}'
```

For structured input, use a `.tfvars` file:

```bash
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```
