# HetznerCloudPrimaryIp Terraform Module

Terraform IaC module for allocating persistent public IP addresses in Hetzner Cloud with optional reverse DNS.

## Structure

```
.
├── main.tf           # Primary IP and conditional rDNS resources
├── outputs.tf        # Stack output definitions
├── variables.tf      # Input variable definitions (metadata, spec, hcloud_token)
├── locals.tf         # IP name and standard label computation
├── provider.tf       # Hetzner Cloud provider configuration
└── BUILD.bazel       # Bazel build definition
```

## Resources Created

- `hcloud_primary_ip` — Allocates an IPv4 address or IPv6 /64 block with labels and protection settings
- `hcloud_rdns` (conditional, via `count`) — Reverse DNS pointer record, created only when `dns_ptr` is non-empty

## Outputs

| Name | Description |
|------|-------------|
| `primary_ip_id` | Hetzner Cloud numeric ID of the created Primary IP |
| `ip_address` | The allocated IP address |
| `ip_network` | The allocated IPv6 /64 CIDR (empty for IPv4) |

## Usage

```bash
# Initialize
terraform init

# Plan
terraform plan \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"mail-ip"}' \
  -var 'spec={"type":"ipv4","location":"fsn1"}'

# Apply
terraform apply \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"mail-ip"}' \
  -var 'spec={"type":"ipv4","location":"fsn1"}'
```

For structured input, use a `.tfvars` file:

```bash
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```
