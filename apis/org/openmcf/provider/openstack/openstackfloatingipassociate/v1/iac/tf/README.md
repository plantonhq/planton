# OpenStackFloatingIpAssociate Terraform Module

Provisions an OpenStack Neutron floating IP association using the OpenStack Terraform provider v3.x.

## Resources Created

- `openstack_networking_floatingip_associate_v2.main` -- The floating IP to port association

## Variables

See `variables.tf` for the complete input schema. Key inputs:

- `spec.floating_ip` -- Floating IP address or UUID (StringValueOrRef)
- `spec.port_id` -- Port to associate with (StringValueOrRef)

## Outputs

See `outputs.tf` for all exported values:

- `id` -- Terraform resource ID
- `floating_ip` -- Associated floating IP address
- `port_id` -- Associated port UUID
- `fixed_ip` -- Mapped fixed IP
- `region` -- OpenStack region
