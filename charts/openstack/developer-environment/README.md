# OpenStack Developer Environment

The **OpenStack Developer Environment** InfraChart provisions an isolated developer sandbox on OpenStack --
a private network with a VM instance, SSH access, optional persistent storage, and optional floating IP
for external connectivity.

Deploy this chart to give each engineer at your organization a self-contained environment for development,
testing, or experimentation on OpenStack infrastructure.

---

## Included Cloud Resources

| Resource                           | Kind                             | Always Created | Controlled by          |
|------------------------------------|----------------------------------|----------------|------------------------|
| **Private Network**                | `OpenStackNetwork`               | Yes            | --                     |
| **Subnet**                         | `OpenStackSubnet`                | Yes            | --                     |
| **Router** (external gateway)      | `OpenStackRouter`                | Yes            | --                     |
| **Router Interface**               | `OpenStackRouterInterface`       | Yes            | --                     |
| **Security Group** (SSH + ICMP)    | `OpenStackSecurityGroup`         | Yes            | --                     |
| **SSH Keypair**                    | `OpenStackKeypair`               | Yes            | --                     |
| **Network Port**                   | `OpenStackNetworkPort`           | Yes            | --                     |
| **VM Instance**                    | `OpenStackInstance`              | Yes            | --                     |
| **Floating IP**                    | `OpenStackFloatingIp`            | *No*           | `floatingIpEnabled`    |
| **Floating IP Association**        | `OpenStackFloatingIpAssociate`   | *No*           | `floatingIpEnabled`    |
| **Persistent Volume**              | `OpenStackVolume`                | *No*           | `volumeEnabled`        |
| **Volume Attachment**              | `OpenStackVolumeAttach`          | *No*           | `volumeEnabled`        |

### How `floatingIpEnabled` works

* `floatingIpEnabled: true` (default) -- Allocates a floating IP from the external network and associates
  it with the VM's network port, making the instance reachable from outside the OpenStack tenant network.
* `floatingIpEnabled: false` -- No floating IP is allocated. The VM is only reachable from within the
  private network (useful for internal-only workloads or environments behind a VPN).

### How `volumeEnabled` works

* `volumeEnabled: true` (default) -- Creates a Cinder block storage volume and attaches it to the VM
  instance. Use this for persistent data that survives instance rebuilds.
* `volumeEnabled: false` -- No volume is created. The instance uses only its ephemeral disk.

---

## Dependency Graph

Resources are deployed in topological order by the Infra Pipeline:

```
Layer 0 (parallel):  Keypair, Network
Layer 1:             Subnet
Layer 2 (parallel):  Router, SecurityGroup
Layer 3 (parallel):  RouterInterface, FloatingIp*
Layer 4:             NetworkPort
Layer 5 (parallel):  Instance, Volume*
Layer 6 (parallel):  FloatingIpAssociate*, VolumeAttach*

* = conditional resource
```

The platform resolves `valueFrom` references automatically -- the Router Interface waits for the Router
and Subnet, the Instance waits for the Port and Keypair, and so on.

---

## Chart Input Values

| Parameter               | Description                                            | Default             |
|-------------------------|--------------------------------------------------------|---------------------|
| **external_network_name** | Pre-existing external (provider) network name        | `public`            |
| **network_cidr**        | CIDR block for the private subnet                      | `192.168.1.0/24`    |
| **dns_nameservers**     | Comma-separated DNS server IPs                         | `8.8.8.8,8.8.4.4`  |
| **instance_flavor**     | Nova flavor for the VM                                 | `m1.medium`         |
| **instance_image**      | Glance image name for the VM                           | `ubuntu-22.04`      |
| **ssh_public_key**      | SSH public key to inject into the VM                   | *(required)*        |
| **volume_size_gb**      | Persistent volume size in GB                           | `50`                |
| **volumeEnabled**       | Create and attach a persistent volume                  | `true`              |
| **floatingIpEnabled**   | Allocate and associate a floating IP                   | `true`              |

### Prerequisites

* An OpenStack environment with Neutron networking, Nova compute, and (optionally) Cinder block storage.
* A pre-existing **external network** accessible to your tenant (typically named `public` or `external`).
* A valid **Glance image** matching the `instance_image` parameter.
* A valid **Nova flavor** matching the `instance_flavor` parameter.

---

## Customization

* **Add more security group rules**: Edit `templates/network.yaml` to add inline rules for HTTP (80),
  HTTPS (443), or application-specific ports.
* **Change the network CIDR**: Override `network_cidr` to avoid collisions with other tenant networks.
* **Disable floating IP**: Set `floatingIpEnabled` to `false` for environments behind a VPN or bastion host.
* **Disable volume**: Set `volumeEnabled` to `false` for stateless or ephemeral development VMs.

---

## Important Notes

* The `ssh_public_key` parameter is required -- without it, you will not be able to SSH into the VM.
* The external network must already exist and be accessible to your OpenStack project. It is not created
  by this chart.
* All cross-resource references are wired with `valueFrom`; you should not need to look up UUIDs manually.
* The security group starts with SSH and ICMP rules only. Add application ports as needed for your workload.

---

© Planton. Licensed under [Apache-2.0](../../../LICENSE).
