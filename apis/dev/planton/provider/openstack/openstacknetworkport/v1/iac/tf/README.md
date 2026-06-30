# OpenStackNetworkPort Terraform Module

Provisions an OpenStack Neutron port using the OpenStack Terraform provider v3.x.

## Resources Created

- `openstack_networking_port_v2.main` -- The Neutron port

## Variables

See `variables.tf` for the complete input schema. Key inputs:

- `metadata.name` -- Port name
- `spec.network_id` -- Network to create the port on (StringValueOrRef)
- `spec.fixed_ips` -- IP allocations from subnets (with nested StringValueOrRef)
- `spec.security_group_ids` -- Security groups to apply (repeated StringValueOrRef)

## Outputs

See `outputs.tf` for all exported values:

- `port_id` -- Port UUID (FK target for downstream components)
- `mac_address` -- Assigned MAC address
- `all_fixed_ips` -- Computed list of all assigned IPs
- `all_security_group_ids` -- Computed list of all applied SG UUIDs
- `region` -- OpenStack region

## Usage

```bash
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```
