# HetznerCloudSnapshot Terraform Module

Terraform IaC module for creating Hetzner Cloud server snapshots stored as Images.

## Structure

```
.
├── main.tf           # Snapshot resource
├── outputs.tf        # Stack output definitions
├── variables.tf      # Input variable definitions (metadata, spec, hcloud_token)
├── locals.tf         # Standard label computation
└── provider.tf       # HetznerCloud provider configuration
```

## Resources Created

- `hcloud_snapshot` — Creates a server snapshot stored as a Hetzner Cloud Image. Captures the full disk of the source server at the moment of creation.

## Outputs

| Name | Description |
|------|-------------|
| `snapshot_id` | Hetzner Cloud image ID of the created snapshot |

## Usage

```bash
# Initialize
terraform init

# Plan
terraform plan \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"my-snapshot"}' \
  -var 'spec={"server_id":"12345678","description":"pre-upgrade baseline"}'

# Apply
terraform apply \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"my-snapshot"}' \
  -var 'spec={"server_id":"12345678","description":"pre-upgrade baseline"}'
```

For structured input, use a `.tfvars` file:

```bash
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```
