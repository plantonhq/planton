---
title: "OpenStackContainerClusterTemplate Research Documentation"
description: "OpenStackContainerClusterTemplate Research Documentation deployment documentation"
icon: "package"
order: 100
componentName: "openstackcontainerclustertemplate"
---

# OpenStackContainerClusterTemplate Research Documentation

## Terraform Resource Analysis

**Resource**: `openstack_containerinfra_clustertemplate_v1`
**Provider**: `terraform-provider-openstack/openstack` v3.x

### Schema Analysis (17 spec fields, 5 StringValueOrRef FKs)

| Attribute | Type | Required | ForceNew | Notes |
|-----------|------|----------|----------|-------|
| `coe` | string | Yes | No | Container Orchestration Engine ("kubernetes") |
| `image` | StringValueOrRef | Yes | No | Base OS image (FK to OpenStackImage) |
| `keypair` | StringValueOrRef | No | No | SSH keypair (FK to OpenStackKeypair) |
| `external_network` | StringValueOrRef | No | No | Provider network (FK to OpenStackNetwork) |
| `fixed_network` | StringValueOrRef | No | No | Tenant network (FK to OpenStackNetwork) |
| `fixed_subnet` | StringValueOrRef | No | No | Tenant subnet (FK to OpenStackSubnet) |
| `network_driver` | string | No | No | Container network driver ("flannel", "calico") |
| `volume_driver` | string | No | No | Container volume driver ("cinder") |
| `dns_nameserver` | string | No | No | DNS nameserver for nodes |
| `docker_volume_size` | int32 | No | No | Docker volume size in GB |
| `flavor` | string | No | No | Nova flavor for worker nodes |
| `master_flavor` | string | No | No | Nova flavor for master nodes |
| `floating_ip_enabled` | bool | No | No | Create floating IP per node |
| `master_lb_enabled` | bool | No | No | Create load balancer for masters |
| `tls_disabled` | bool | No | No | Disable TLS for cluster API |
| `labels` | map(string,string) | No | No | Magnum labels (k8s version, runtime, etc.) |
| `region` | string | No | Yes | OpenStack region override |

### Key Design Decisions

#### All Fields Updatable (No ForceNew)

Unlike most OpenStack resources, Magnum cluster templates support PATCH-style updates.
Only `region` is ForceNew. This makes templates safe to modify in place without cluster recreation.

#### Five FK References

This component has more FK references than any other OpenStack component:
1. `image` -> OpenStackImage (required)
2. `keypair` -> OpenStackKeypair (optional)
3. `external_network` -> OpenStackNetwork (optional)
4. `fixed_network` -> OpenStackNetwork (optional)
5. `fixed_subnet` -> OpenStackSubnet (optional)

### Pulumi SDK Mapping

| Spec Field | Pulumi Type | Pulumi Field |
|------------|-------------|--------------|
| `coe` | `pulumi.StringInput` | `Coe` |
| `image` | `pulumi.StringInput` | `Image` |
| `keypair` | `pulumi.StringPtrInput` | `KeypairId` |
| `external_network` | `pulumi.StringPtrInput` | `ExternalNetworkId` |
| `fixed_network` | `pulumi.StringPtrInput` | `FixedNetwork` |
| `fixed_subnet` | `pulumi.StringPtrInput` | `FixedSubnet` |
| `network_driver` | `pulumi.StringPtrInput` | `NetworkDriver` |
| `volume_driver` | `pulumi.StringPtrInput` | `VolumeDriver` |
| `dns_nameserver` | `pulumi.StringPtrInput` | `DnsNameserver` |
| `docker_volume_size` | `pulumi.IntPtrInput` | `DockerVolumeSize` |
| `flavor` | `pulumi.StringPtrInput` | `Flavor` |
| `master_flavor` | `pulumi.StringPtrInput` | `MasterFlavor` |
| `floating_ip_enabled` | `pulumi.BoolPtrInput` | `FloatingIpEnabled` |
| `master_lb_enabled` | `pulumi.BoolPtrInput` | `MasterLbEnabled` |
| `tls_disabled` | `pulumi.BoolPtrInput` | `TlsDisabled` |
| `labels` | `pulumi.StringMapInput` | `Labels` |
| `region` | `pulumi.StringPtrInput` | `Region` |
