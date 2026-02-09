# OpenStackVolume Terraform Module

Provisions an OpenStack Cinder block storage volume using the `openstack_blockstorage_volume_v3` resource.

## Usage

```bash
# Initialize
terraform init

# Plan
terraform plan -var-file=terraform.tfvars.json

# Apply
terraform apply -var-file=terraform.tfvars.json
```

## Resources Created

- `openstack_blockstorage_volume_v3.main` -- The Cinder volume

## Inputs

See `variables.tf` for the full variable specification.

## Outputs

| Output | Description |
|--------|-------------|
| `volume_id` | UUID of the volume |
| `name` | Volume name |
| `size` | Size in GB |
| `volume_type` | Cinder volume type |
| `availability_zone` | AZ where volume was created |
| `region` | OpenStack region |
