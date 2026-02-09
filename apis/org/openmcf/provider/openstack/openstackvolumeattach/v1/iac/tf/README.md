# OpenStackVolumeAttach Terraform Module

Attaches an OpenStack Cinder volume to a compute instance using the `openstack_compute_volume_attach_v2` resource.

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Resources Created

- `openstack_compute_volume_attach_v2.main` -- The volume-to-instance attachment

## Outputs

| Output | Description |
|--------|-------------|
| `id` | Terraform resource ID |
| `instance_id` | Attached instance UUID |
| `volume_id` | Attached volume UUID |
| `device` | Device path in instance |
| `region` | OpenStack region |
