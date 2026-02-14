# Standard VM Instance

This preset creates a compute instance booting from an image with a specified flavor, keypair, and network. The root disk is ephemeral (lives on the hypervisor). This is the simplest and most common instance configuration.

## When to Use

- General-purpose VMs for applications, web servers, or worker nodes
- Development and testing instances where persistent root disks are not required
- Quick provisioning where ephemeral storage is acceptable

## Key Configuration Choices

- **Flavor by name** (`flavorName`) -- more readable than flavor ID; use `openstack flavor list` to find available options
- **Image by name** (`imageName`) -- references a Glance image (e.g., `ubuntu-22.04`, `centos-9-stream`)
- **SSH keypair** (`keyPair`) -- enables key-based SSH authentication
- **Single network** -- one NIC on the specified network; add more entries to `networks` for multi-homed instances
- **Security group by name** -- references the SG name (not UUID); OpenStack resolves it within the project

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<flavor-name>` | Instance flavor (e.g., `m1.medium`, `c1.xlarge`) | `openstack flavor list` |
| `<image-name>` | Glance image name (e.g., `ubuntu-22.04`) | `openstack image list` |
| `<keypair-name>` | SSH keypair name | `openstack keypair list` or `OpenStackKeypair` status outputs |
| `<network-id>` | ID of the network to attach the instance to | OpenStack console or `OpenStackNetwork` status outputs |
| `<security-group-name>` | Name of the security group to apply | `openstack security group list` or `OpenStackSecurityGroup` status outputs |

## Related Presets

- **02-boot-from-volume** -- Use instead when the root disk must persist independently of the instance lifecycle
