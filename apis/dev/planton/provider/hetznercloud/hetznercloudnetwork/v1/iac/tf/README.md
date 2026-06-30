# HetznerCloudNetwork Terraform Module

Terraform IaC module for creating private networks with subnets and static routes in Hetzner Cloud.

## Structure

```
.
├── main.tf           # Network, subnet (for_each), and route (for_each) resources
├── outputs.tf        # Stack output definitions
├── variables.tf      # Input variable definitions (metadata, spec, hcloud_token)
├── locals.tf         # Network name and standard label computation
├── provider.tf       # HetznerCloud provider configuration
└── BUILD.bazel       # Bazel build definition
```

## Resources Created

- `hcloud_network` — Network with top-level CIDR, labels, and protection settings
- `hcloud_network_subnet` (1 per subnet) — Subnets keyed by `ip_range` via `for_each`
- `hcloud_network_route` (1 per route, optional) — Routes keyed by `destination` via `for_each`

## Outputs

| Name | Description |
|------|-------------|
| `network_id` | Hetzner Cloud numeric ID of the created network |

## Usage

```bash
# Build
bazel build //apis/dev/planton/provider/hetznercloud/hetznercloudnetwork/v1/iac/tf:terraform_module

# Initialize
terraform init

# Plan
terraform plan -var-file=../hack/manifest.tfvars

# Apply
terraform apply -var-file=../hack/manifest.tfvars
```
