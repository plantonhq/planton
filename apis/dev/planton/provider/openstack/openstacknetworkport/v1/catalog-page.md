# OpenStack Network Port

Deploys an OpenStack Neutron port, providing a stable network identity (MAC address, fixed IPs, security groups) on a Neutron network. Explicit ports are preferred over instance-inline networking when you need stable IP addresses that survive instance rebuilds, pre-provisioned network identities for InfraChart orchestration, or fine-grained security group assignments.

## What Gets Created

When you deploy an OpenStackNetworkPort resource, Planton provisions:

- **Neutron Port** — an `openstack_networking_port_v2` resource on the specified network, with configured fixed IPs, security groups, MAC address, admin state, port security, and tags

## Prerequisites

- **OpenStack credentials** configured via environment variables or Planton provider config
- **An existing Neutron network** — provided as a literal UUID or via `valueFrom` reference to an OpenStackNetwork resource
- **Existing subnets** if specifying `fixedIps` with explicit `subnetId` values
- **Existing security groups** if specifying `securityGroupIds`

## Quick Start

Create a file `port.yaml`:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackNetworkPort
metadata:
  name: my-port
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OpenStackNetworkPort.my-port
spec:
  networkId:
    value: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
```

Deploy:

```shell
planton apply -f port.yaml
```

This creates a Neutron port named `my-port` on the specified network with default settings: admin state up, the project's default security group, and an auto-assigned IP from any subnet on the network.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `networkId` | `StringValueOrRef` | ID of the network to create this port on. Every port belongs to exactly one network. ForceNew: changing the network recreates the port. Can reference an OpenStackNetwork resource via `valueFrom`. | required |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `fixedIps` | `FixedIp[]` | auto-assigned | IP address allocations for this port. Each entry assigns an IP from a subnet on the port's network. If omitted, OpenStack auto-assigns one IP from any subnet on the network. Multiple entries create a multi-homed port. See nested fields below. |
| `securityGroupIds` | `StringValueOrRef[]` | project default SG | Security groups to apply to this port. Each entry can reference an OpenStackSecurityGroup resource via `valueFrom` or be a literal UUID. Mutually exclusive with `noSecurityGroups`. |
| `noSecurityGroups` | `bool` | `false` | Explicitly removes all security groups from this port, including the default security group. Use for load balancer VIPs or network appliance ports. Mutually exclusive with `securityGroupIds`. |
| `adminStateUp` | `bool` | `true` | Administrative state of the port. When `false`, the port is down and does not forward traffic. |
| `macAddress` | `string` | auto-assigned | Specific MAC address for this port. ForceNew: changing the MAC recreates the port. Use for network bonding, DPDK, or license-tied MAC addresses. |
| `portSecurityEnabled` | `bool` | inherited from network | Controls whether port security is enforced. When enabled, only traffic matching security groups and allowed address pairs is permitted. If omitted, inherits from the network's `portSecurityEnabled` setting. |
| `description` | `string` | — | Human-readable description, visible in the OpenStack API and Horizon. |
| `tags` | `string[]` | `[]` | Tags for filtering and organization in the OpenStack API. Must be unique. |
| `region` | `string` | provider default | Overrides the region from the provider config for this port. |

#### FixedIp Nested Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `subnetId` | `StringValueOrRef` | auto-selected | Subnet to allocate an IP address from. Can reference an OpenStackSubnet resource via `valueFrom` or be a literal UUID. If omitted, OpenStack auto-selects a subnet on the port's network. |
| `ipAddress` | `string` | auto-assigned | Specific IP address to request from the subnet's allocation pool. Must belong to the subnet's CIDR and be within an allocation pool. |

## Examples

### Basic Port on a Network

A port with a single auto-assigned IP, suitable for pre-provisioning a network identity before launching an instance:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackNetworkPort
metadata:
  name: web-port
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.OpenStackNetworkPort.web-port
spec:
  networkId:
    value: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
  description: Web server port
```

### Port with Fixed IP and Security Groups

A port with a specific IP address and multiple security groups, using `valueFrom` references to other Planton resources:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackNetworkPort
metadata:
  name: app-port
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: staging.OpenStackNetworkPort.app-port
spec:
  networkId:
    valueFrom:
      kind: OpenStackNetwork
      name: app-network
      fieldPath: status.outputs.network_id
  fixedIps:
    - subnetId:
        valueFrom:
          kind: OpenStackSubnet
          name: app-subnet
          fieldPath: status.outputs.subnet_id
      ipAddress: "10.0.1.100"
  securityGroupIds:
    - valueFrom:
        kind: OpenStackSecurityGroup
        name: ssh-sg
        fieldPath: status.outputs.security_group_id
    - valueFrom:
        kind: OpenStackSecurityGroup
        name: web-sg
        fieldPath: status.outputs.security_group_id
  description: Application server port with fixed IP
  tags:
    - staging
    - app-tier
```

### Full-Featured Port with No Security Groups

A port for a network appliance that bypasses all security groups, uses a specific MAC address, and disables port security:

```yaml
apiVersion: openstack.planton.dev/v1
kind: OpenStackNetworkPort
metadata:
  name: appliance-port
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: prod.OpenStackNetworkPort.appliance-port
spec:
  networkId:
    valueFrom:
      kind: OpenStackNetwork
      name: transit-network
      fieldPath: status.outputs.network_id
  fixedIps:
    - subnetId:
        valueFrom:
          kind: OpenStackSubnet
          name: transit-subnet
          fieldPath: status.outputs.subnet_id
      ipAddress: "172.16.0.1"
    - subnetId:
        valueFrom:
          kind: OpenStackSubnet
          name: mgmt-subnet
          fieldPath: status.outputs.subnet_id
  noSecurityGroups: true
  macAddress: "fa:16:3e:aa:bb:cc"
  portSecurityEnabled: false
  adminStateUp: true
  description: Network appliance transit port
  region: RegionOne
  tags:
    - production
    - appliance
    - transit
```

## Stack Outputs

After deployment, the following outputs are available in `status.outputs`:

| Output | Type | Description |
|--------|------|-------------|
| `port_id` | `string` | UUID of the created Neutron port. Primary FK target for downstream components. |
| `mac_address` | `string` | MAC address assigned to the port (auto-generated or explicitly set). |
| `all_fixed_ips` | `string[]` | All IP addresses assigned to this port, including both explicitly requested and auto-assigned IPs. |
| `all_security_group_ids` | `string[]` | All security group UUIDs applied to this port, including the default SG if no explicit SGs were set. |
| `region` | `string` | OpenStack region where the port was created. |

## Related Components

- [OpenStackNetwork](/docs/catalog/openstack/openstacknetwork) — the network this port belongs to
- [OpenStackSubnet](/docs/catalog/openstack/openstacksubnet) — defines IP address ranges that fixed IPs are allocated from
- [OpenStackSecurityGroup](/docs/catalog/openstack/openstacksecuritygroup) — security groups applied to this port
- [OpenStackFloatingIp](/docs/catalog/openstack/openstackfloatingip) — allocates a floating IP that can be associated with this port
- [OpenStackFloatingIpAssociate](/docs/catalog/openstack/openstackfloatingipassociate) — associates a floating IP to this port via `portId`
- [OpenStackInstance](/docs/catalog/openstack/openstackinstance) — attaches this port to a compute instance
