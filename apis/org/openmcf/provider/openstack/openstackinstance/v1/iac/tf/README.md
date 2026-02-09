# OpenStackInstance Terraform Module

Provisions an OpenStack Compute instance with full networking, storage, and placement support.

## Resources Created

- `openstack_compute_instance_v2` -- A compute instance with dynamic network, block_device, and scheduler_hints blocks

## Key Terraform Patterns

- **Dynamic blocks** for `network`, `block_device`, and `scheduler_hints`
- **Conditional nulls** for optional fields (`!= "" ? value : null`)
- **List comprehension** for security group FK resolution (`[for sg in var.spec.security_groups : sg.value]`)
- **server_group_id** mapped to `scheduler_hints { group = ... }` via dynamic block with 0-1 iteration
