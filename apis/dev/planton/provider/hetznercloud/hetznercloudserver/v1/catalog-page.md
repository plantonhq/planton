# Hetzner Cloud Server

Provisions a Hetzner Cloud virtual machine running a chosen OS image on a specified server type in a given location. The server is the core compute resource and central hub in the Hetzner Cloud catalog — it references SSH keys, firewalls, placement groups, private networks, and Primary IPs at creation time. An optional reverse DNS record can be configured for the server's auto-assigned public IPv4 address.

## What Gets Created

- **Server** — an `hcloud_server` resource provisioning a virtual machine with the specified hardware profile, OS image, location, SSH keys, firewall rules, placement group, public networking configuration, private network attachments, cloud-init user data, backup settings, and standard labels computed from resource metadata.
- **Reverse DNS record** (when `dnsPtr` is set) — an `hcloud_rdns` resource mapping the server's auto-assigned public IPv4 address back to the specified hostname. Created only when the `dnsPtr` field is non-empty. Not needed when a Primary IP is attached via `publicNet.ipv4` — manage rDNS on the `HetznerCloudPrimaryIp` component instead.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or Planton provider config
- **SSH keys** if using `sshKeys` for server access — either pre-existing in the Hetzner Cloud project or managed as `HetznerCloudSshKey` components
- **A network with at least one subnet** if attaching to a private network via `networks`
- **Primary IPs in the same location** if attaching existing IPs via `publicNet.ipv4` or `publicNet.ipv6`

## Quick Start

Create a file `server.yaml`:

```yaml
apiVersion: hetzner-cloud.planton.dev/v1
kind: HetznerCloudServer
metadata:
  name: my-server
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.HetznerCloudServer.my-server
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
```

Deploy:

```shell
planton apply -f server.yaml
```

This provisions a shared x86 server (2 vCPU, 4 GB RAM) running Ubuntu 24.04 in Falkenstein with auto-assigned public IPv4 and IPv6 addresses.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `serverType` | `string` | Server type determining vCPU, RAM, and disk. Examples: `cx22`, `cpx11`, `cax11`, `ccx13`. Changing this triggers an in-place resize (server is stopped temporarily). | `min_len: 1` |
| `image` | `string` | OS image name or numeric snapshot ID. Examples: `ubuntu-24.04`, `debian-12`, `rocky-9`, `45346857`. Changing this forces server replacement. | `min_len: 1` |
| `location` | `string` | Hetzner Cloud location for the server. Known locations: `fsn1`, `nbg1`, `hel1`, `ash`, `hil`, `sin`. Primary IPs and Floating IPs assigned to the server must be in the same location. Changing this forces server replacement. | `min_len: 1` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `sshKeys` | `StringValueOrRef[]` | empty | SSH keys to inject at creation. Accepts literal SSH key names or IDs, or references to `HetznerCloudSshKey` resources via `valueFrom`. Changing this forces server replacement. |
| `userData` | `string` | empty | Cloud-init script (`#!/bin/bash`) or cloud-config YAML (`#cloud-config`) executed on first boot. Maximum 32 KB. Changing this forces server replacement. |
| `placementGroupId` | `StringValueOrRef` | unset | Placement group for anti-affinity scheduling. Can reference a `HetznerCloudPlacementGroup` resource via `valueFrom`. |
| `firewallIds` | `StringValueOrRef[]` | empty | Firewalls to apply at creation. Each entry can reference a `HetznerCloudFirewall` resource via `valueFrom`. |
| `publicNet` | `object` | unset | Public networking configuration. When omitted, the server receives auto-assigned public IPv4 and IPv6 (provider default). |
| `publicNet.ipv4Enabled` | `bool` | `true` | Enable public IPv4 for the server. Only takes effect when `publicNet` is set. |
| `publicNet.ipv6Enabled` | `bool` | `true` | Enable public IPv6 for the server. Only takes effect when `publicNet` is set. |
| `publicNet.ipv4` | `StringValueOrRef` | unset | Existing Primary IP (IPv4) to attach instead of auto-assigning. Can reference a `HetznerCloudPrimaryIp` resource via `valueFrom`. Must be in the same location. |
| `publicNet.ipv6` | `StringValueOrRef` | unset | Existing Primary IP (IPv6) to attach instead of auto-assigning. Can reference a `HetznerCloudPrimaryIp` resource via `valueFrom`. Must be in the same location. |
| `networks` | `NetworkAttachment[]` | empty | Private network attachments. Each entry attaches the server to a network. |
| `networks[].networkId` | `StringValueOrRef` | — | Network to attach to. Required within each network attachment. Can reference a `HetznerCloudNetwork` resource via `valueFrom`. |
| `networks[].ip` | `string` | auto-assigned | Specific IP address within the network's subnet range. |
| `networks[].aliasIps` | `string[]` | empty | Additional IP addresses for the server within this network. |
| `backups` | `bool` | `false` | Enable automatic daily backups (14-day retention, +20% cost). |
| `keepDisk` | `bool` | `false` | Preserve disk size when changing `serverType`. Prevents irreversible disk upgrades. |
| `deleteProtection` | `bool` | `false` | Prevent accidental deletion via the Hetzner Cloud API. |
| `rebuildProtection` | `bool` | `false` | Prevent accidental rebuild (re-image) of the server. |
| `shutdownBeforeDeletion` | `bool` | `false` | Send ACPI shutdown signal and wait for clean power-off before deletion. |
| `dnsPtr` | `string` | empty | Reverse DNS pointer record for the server's auto-assigned public IPv4. Creates an `hcloud_rdns` resource when non-empty. Do not use when a Primary IP is attached via `publicNet.ipv4`. |

## Examples

### Minimal Server

A shared x86 server running Ubuntu 24.04 in Falkenstein with auto-assigned public IPs.

```yaml
apiVersion: hetzner-cloud.planton.dev/v1
kind: HetznerCloudServer
metadata:
  name: dev-box
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: my-org
    pulumi.planton.dev/project: my-project
    pulumi.planton.dev/stack.name: dev.HetznerCloudServer.dev-box
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
```

### Server with SSH Key and Cloud-Init

A server with SSH key access and a cloud-init script that installs Nginx on first boot.

```yaml
apiVersion: hetzner-cloud.planton.dev/v1
kind: HetznerCloudServer
metadata:
  name: web-01
  org: acme-corp
  env: staging
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme-corp
    pulumi.planton.dev/project: web-platform
    pulumi.planton.dev/stack.name: staging.HetznerCloudServer.web-01
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
  sshKeys:
    - value: "deploy-key"
  userData: |
    #!/bin/bash
    apt-get update && apt-get install -y nginx
    systemctl enable nginx
```

### Production Server with Firewall and Private Network

A server composed with other Planton components via `valueFrom` references, with backups and protections enabled.

```yaml
apiVersion: hetzner-cloud.planton.dev/v1
kind: HetznerCloudServer
metadata:
  name: app-01
  org: acme-corp
  env: production
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme-corp
    pulumi.planton.dev/project: infrastructure
    pulumi.planton.dev/stack.name: production.HetznerCloudServer.app-01
spec:
  serverType: cpx21
  image: debian-12
  location: fsn1
  sshKeys:
    - valueFrom:
        kind: HetznerCloudSshKey
        name: prod-key
        fieldPath: status.outputs.ssh_key_id
  firewallIds:
    - valueFrom:
        kind: HetznerCloudFirewall
        name: web-firewall
        fieldPath: status.outputs.firewall_id
  networks:
    - networkId:
        valueFrom:
          kind: HetznerCloudNetwork
          name: main-vpc
          fieldPath: status.outputs.network_id
      ip: "10.0.1.10"
  backups: true
  deleteProtection: true
  rebuildProtection: true
  shutdownBeforeDeletion: true
```

### HA Server with Placement Group

An anti-affinity server in a spread placement group for high availability. Uses a dedicated server type for guaranteed CPU performance.

```yaml
apiVersion: hetzner-cloud.planton.dev/v1
kind: HetznerCloudServer
metadata:
  name: db-primary
  org: acme-corp
  env: production
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme-corp
    pulumi.planton.dev/project: databases
    pulumi.planton.dev/stack.name: production.HetznerCloudServer.db-primary
    role: database
spec:
  serverType: ccx13
  image: ubuntu-24.04
  location: fsn1
  sshKeys:
    - valueFrom:
        kind: HetznerCloudSshKey
        name: prod-key
        fieldPath: status.outputs.ssh_key_id
  placementGroupId:
    valueFrom:
      kind: HetznerCloudPlacementGroup
      name: db-spread
      fieldPath: status.outputs.placement_group_id
  firewallIds:
    - valueFrom:
        kind: HetznerCloudFirewall
        name: db-firewall
        fieldPath: status.outputs.firewall_id
  networks:
    - networkId:
        valueFrom:
          kind: HetznerCloudNetwork
          name: main-vpc
          fieldPath: status.outputs.network_id
  backups: true
  keepDisk: true
  deleteProtection: true
  rebuildProtection: true
  shutdownBeforeDeletion: true
```

### Full-Featured Server with Primary IP

A server with a stable IPv6 Primary IP, private networking, all protections, and reverse DNS for the auto-assigned IPv4.

```yaml
apiVersion: hetzner-cloud.planton.dev/v1
kind: HetznerCloudServer
metadata:
  name: web-prod-01
  org: acme-corp
  env: production
  labels:
    planton.dev/provisioner: pulumi
    pulumi.planton.dev/organization: acme-corp
    pulumi.planton.dev/project: infrastructure
    pulumi.planton.dev/stack.name: production.HetznerCloudServer.web-prod-01
spec:
  serverType: cpx31
  image: ubuntu-24.04
  location: fsn1
  sshKeys:
    - valueFrom:
        kind: HetznerCloudSshKey
        name: prod-key
        fieldPath: status.outputs.ssh_key_id
  placementGroupId:
    valueFrom:
      kind: HetznerCloudPlacementGroup
      name: web-spread
      fieldPath: status.outputs.placement_group_id
  firewallIds:
    - valueFrom:
        kind: HetznerCloudFirewall
        name: web-firewall
        fieldPath: status.outputs.firewall_id
  publicNet:
    ipv4Enabled: true
    ipv6Enabled: true
    ipv6:
      valueFrom:
        kind: HetznerCloudPrimaryIp
        name: web-ipv6
        fieldPath: status.outputs.primary_ip_id
  networks:
    - networkId:
        valueFrom:
          kind: HetznerCloudNetwork
          name: main-vpc
          fieldPath: status.outputs.network_id
      ip: "10.0.1.20"
      aliasIps:
        - "10.0.1.21"
  userData: |
    #!/bin/bash
    apt-get update && apt-get install -y nginx certbot python3-certbot-nginx
    systemctl enable nginx
  backups: true
  keepDisk: true
  deleteProtection: true
  rebuildProtection: true
  shutdownBeforeDeletion: true
  dnsPtr: web-prod-01.example.com
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `server_id` | `string` | Hetzner Cloud numeric ID of the created server. Referenced by `HetznerCloudVolume`, `HetznerCloudSnapshot`, `HetznerCloudFloatingIp`, and `HetznerCloudLoadBalancer` via `StringValueOrRef`. |
| `ipv4_address` | `string` | The public IPv4 address assigned to the server. Empty if public IPv4 is disabled via `publicNet.ipv4Enabled = false`. |
| `ipv6_address` | `string` | The first IPv6 address of the server's assigned /64 network. Empty if public IPv6 is disabled via `publicNet.ipv6Enabled = false`. |
| `status` | `string` | The current status of the server: `running`, `off`, `rebuilding`, or `migrating`. |

## Related Components

- [HetznerCloudSshKey](/docs/catalog/hetznercloud/hetznercloudsshkey) — SSH keys for server access, referenced via `sshKeys`
- [HetznerCloudPlacementGroup](/docs/catalog/hetznercloud/hetznercloudplacementgroup) — Anti-affinity scheduling, referenced via `placementGroupId`
- [HetznerCloudFirewall](/docs/catalog/hetznercloud/hetznercloudfirewall) — Network security rules, referenced via `firewallIds`
- [HetznerCloudNetwork](/docs/catalog/hetznercloud/hetznercloudnetwork) — Private networking, referenced via `networks[].networkId`
- [HetznerCloudPrimaryIp](/docs/catalog/hetznercloud/hetznercloudprimaryip) — Stable public IPs, referenced via `publicNet.ipv4` / `publicNet.ipv6`
- [HetznerCloudFloatingIp](/docs/catalog/hetznercloud/hetznercloudfloatingip) — Reassignable IPs that can be assigned to this server
- [HetznerCloudVolume](/docs/catalog/hetznercloud/hetznercloudvolume) — Block storage that can be attached to this server
- [HetznerCloudSnapshot](/docs/catalog/hetznercloud/hetznercloudsnapshot) — Server image snapshots taken from this server
