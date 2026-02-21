# HetznerCloudVolume Terraform Module

Terraform IaC module for provisioning Hetzner Cloud block storage volumes with optional server attachment and automount.

## Structure

```
.
├── main.tf           # Volume resource; conditional volume attachment via count
├── outputs.tf        # Stack output definitions
├── variables.tf      # Input variable definitions (metadata, spec, hcloud_token)
├── locals.tf         # Volume name derivation and standard label computation
└── provider.tf       # HetznerCloud provider configuration
```

## Resources Created

- `hcloud_volume` — Provisions a block storage volume with the specified size, location, optional filesystem format, labels, and delete protection
- `hcloud_volume_attachment` (conditional, via `count`) — Attaches the volume to a server with optional automount, created only when `server_id` is non-null

## Outputs

| Name | Description |
|------|-------------|
| `volume_id` | Hetzner Cloud numeric ID of the created volume |
| `linux_device` | Linux device path for the volume on the attached server |

## Usage

```bash
# Initialize
terraform init

# Plan
terraform plan \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"my-volume"}' \
  -var 'spec={"size":100,"location":"fsn1","format":"ext4"}'

# Apply
terraform apply \
  -var 'hcloud_token=your-api-token' \
  -var 'metadata={"name":"my-volume"}' \
  -var 'spec={"size":100,"location":"fsn1","format":"ext4"}'
```

For structured input, use a `.tfvars` file:

```bash
terraform plan -var-file=terraform.tfvars
terraform apply -var-file=terraform.tfvars
```
