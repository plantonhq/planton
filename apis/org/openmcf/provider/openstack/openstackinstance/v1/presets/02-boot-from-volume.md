# Boot-from-Volume Instance

This preset creates a compute instance that boots from a Cinder volume instead of an ephemeral disk. The root volume is created from a Glance image and persists independently of the instance (unless `deleteOnTermination` is true). Use this for production workloads that need persistent root disks.

## When to Use

- Production instances where the root disk must survive instance deletion or migration
- Workloads on hypervisors with limited local storage
- Environments that require volume-level snapshots of the root disk

## Key Configuration Choices

- **Boot from volume** (`blockDevice` with `sourceType: image`, `destinationType: volume`) -- root disk is a Cinder volume, not ephemeral
- **50 GB root volume** -- adjust `volumeSize` to match your needs
- **Delete on termination** (`deleteOnTermination: true`) -- root volume is cleaned up when the instance is deleted; set to `false` for truly persistent volumes
- **No `imageName`/`imageId`** -- the image is specified in the block device mapping instead
- **Boot index 0** -- marks this as the boot volume

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<flavor-name>` | Instance flavor (e.g., `m1.medium`) | `openstack flavor list` |
| `<keypair-name>` | SSH keypair name | `openstack keypair list` or `OpenStackKeypair` status outputs |
| `<network-id>` | ID of the network to attach the instance to | OpenStack console or `OpenStackNetwork` status outputs |
| `<security-group-name>` | Name of the security group | `openstack security group list` or `OpenStackSecurityGroup` status outputs |
| `<image-id>` | UUID of the Glance image to boot from | `openstack image list` or `OpenStackImage` status outputs |

## Related Presets

- **01-standard-vm** -- Use instead for ephemeral root disks (simpler, faster provisioning)
