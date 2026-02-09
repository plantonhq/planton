# OpenStackContainerClusterTemplate Terraform Module

Provisions an OpenStack Magnum cluster template using the OpenStack Terraform provider v3.x.

## Resources Created

- `openstack_containerinfra_clustertemplate_v1.main` -- The Magnum cluster template

## Variables

See `variables.tf` for the complete input schema. Key inputs:

- `spec.coe` -- Container Orchestration Engine (e.g., "kubernetes")
- `spec.image` -- Base OS image for cluster nodes (StringValueOrRef)
- `spec.keypair` -- SSH keypair for node access (StringValueOrRef, optional)
- `spec.external_network` -- External network for outbound connectivity (StringValueOrRef, optional)
- `spec.fixed_network` -- Fixed network for cluster nodes (StringValueOrRef, optional)
- `spec.fixed_subnet` -- Fixed subnet for cluster nodes (StringValueOrRef, optional)

## Outputs

See `outputs.tf` for all exported values:

- `template_id` -- The UUID of the cluster template
- `name` -- The name of the cluster template
- `coe` -- The Container Orchestration Engine
- `region` -- The OpenStack region
