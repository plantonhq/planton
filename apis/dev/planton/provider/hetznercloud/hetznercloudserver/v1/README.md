# HetznerCloudServer

The **HetznerCloudServer** resource provisions a cloud server in Hetzner Cloud — a virtual machine running a chosen OS image on a specific server type (vCPU, RAM, disk combination) in a given location. The server is the core compute primitive in Hetzner Cloud and the hub resource in the Planton Hetzner Cloud catalog: it ties together SSH keys for access, firewalls for security, placement groups for anti-affinity, private networks for internal communication, and Primary IPs for stable public addressing.

## What It Represents

A [Hetzner Cloud Server](https://docs.hetzner.cloud/#servers) is a virtual machine provisioned on shared or dedicated hardware (depending on the server type family). The server type determines the vCPU count, RAM, disk size, and pricing tier. The image determines the operating system installed at creation time. The location determines the physical datacenter where the server runs.

By default, a server receives auto-assigned public IPv4 and IPv6 addresses. These addresses change when the server is deleted and recreated. For stable public IPs that survive server replacement, attach Primary IPs via the `publicNet` block or assign Floating IPs after creation.

## Bundled Resources

| Terraform Resource | Count | Created When | Purpose |
|---|---|---|---|
| `hcloud_server` | 1 | Always | Provisions the server with the specified type, image, location, SSH keys, firewall rules, network attachments, placement group, public networking configuration, cloud-init, backup settings, and protections |
| `hcloud_rdns` | 0 or 1 | When `dnsPtr` is non-empty | Sets a reverse DNS pointer record for the server's auto-assigned public IPv4 address |

The rDNS resource is bundled because it targets the server's auto-assigned IPv4 — an address that only exists as a server attribute. When `dnsPtr` is omitted, only the server resource is created.

**Important:** If you assign a Primary IP via `publicNet.ipv4`, manage rDNS on the `HetznerCloudPrimaryIp` component instead. The server's `dnsPtr` field targets the auto-assigned IPv4, which is replaced when a Primary IP occupies the slot. Setting both creates conflicting rDNS management.

## Key Features

### Server Types

The `serverType` field selects the hardware profile. Hetzner Cloud offers several families:

| Family | Prefix | CPU | Example | vCPU / RAM / Disk |
|--------|--------|-----|---------|-------------------|
| Shared x86 (Intel) | `cx` | Intel | `cx22` | 2 vCPU / 4 GB / 40 GB |
| Shared x86 (AMD) | `cpx` | AMD | `cpx11` | 2 vCPU / 2 GB / 40 GB |
| Shared ARM64 | `cax` | Ampere | `cax11` | 2 vCPU / 4 GB / 40 GB |
| Dedicated x86 | `ccx` | AMD | `ccx13` | 2 vCPU / 8 GB / 80 GB |

Changing `serverType` triggers a server resize: the server is temporarily stopped, resized, and restarted. Use `keepDisk` to prevent irreversible disk upgrades during resizes.

### Images

The `image` field accepts an OS image name (e.g., `ubuntu-24.04`, `debian-12`, `rocky-9`) or a numeric snapshot ID (as a string). Changing the image forces server replacement — the existing server is destroyed and a new one is created.

### SSH Key Injection

The `sshKeys` field accepts a list of SSH key names or IDs (via `StringValueOrRef`). Keys are injected at server creation time only — changes to this list force server replacement. Each entry can reference a `HetznerCloudSshKey` resource's output via `valueFrom`.

### Placement Groups

The `placementGroupId` field assigns the server to an anti-affinity placement group. Servers in a `spread` placement group are guaranteed to run on different physical hosts. Accepts a `StringValueOrRef` referencing a `HetznerCloudPlacementGroup` resource.

### Firewall Attachment

The `firewallIds` field applies one or more firewalls to the server at creation time. Each entry accepts a `StringValueOrRef` referencing a `HetznerCloudFirewall` resource. Firewalls control inbound and outbound traffic at the infrastructure level (before packets reach the server's OS).

### Public Networking

If the `publicNet` block is omitted, the server receives auto-assigned public IPv4 and IPv6 addresses (provider default). When `publicNet` is set:

- `ipv4Enabled` / `ipv6Enabled` — control whether public IPv4/IPv6 are enabled (default: `true`)
- `ipv4` / `ipv6` — attach existing Primary IPs instead of auto-assigning. Each accepts a `StringValueOrRef` referencing a `HetznerCloudPrimaryIp` resource.

Setting `publicNet` with both protocols disabled creates a private-only server reachable only through private networks.

### Private Network Attachments

The `networks` field attaches the server to one or more private networks via inline network blocks. Each `NetworkAttachment` requires a `networkId` (via `StringValueOrRef` referencing a `HetznerCloudNetwork`) and optionally specifies a fixed `ip` within the subnet range and `aliasIps` for hosting multiple services on one server.

### Cloud-Init

The `userData` field accepts a cloud-init script (starting with `#!/bin/bash`) or cloud-config YAML (starting with `#cloud-config`). Maximum size is 32 KB. This runs on first boot only. Changing `userData` forces server replacement.

### Backups

When `backups` is `true`, Hetzner Cloud takes automatic daily backups retained for 14 days. Backups cost an additional 20% of the server price. Backups can be enabled or disabled without replacing the server.

### Resize Behavior (keepDisk)

When `keepDisk` is `true`, changing `serverType` only upgrades vCPU and RAM — the disk remains at its current size. This preserves the ability to downgrade to a smaller server type later. When `false` (default), a server type upgrade also enlarges the disk, which is irreversible.

### Protections

- `deleteProtection` — prevents accidental deletion via the Hetzner Cloud API
- `rebuildProtection` — prevents accidental rebuild (re-image) of the server

Both must be explicitly disabled before the protected action can be performed.

### Graceful Shutdown

When `shutdownBeforeDeletion` is `true`, an ACPI shutdown signal is sent and the system waits for the server to power off before deletion. This allows running applications to flush buffers and close connections cleanly.

### Reverse DNS

When `dnsPtr` is set, an `hcloud_rdns` resource maps the server's auto-assigned public IPv4 address to the specified hostname. Use this only when the server uses auto-assigned IPv4 (the default, or `publicNet` with `ipv4Enabled: true` and no `ipv4` Primary IP reference). If a Primary IP is attached via `publicNet.ipv4`, manage rDNS on the `HetznerCloudPrimaryIp` component instead.

### Automatic Labeling

Standard labels (`resource`, `name`, `kind`, `org`, `env`, `id`) are applied to the server from metadata. User-specified `metadata.labels` are merged in, with standard labels taking precedence. The rDNS resource does not support labels in the Hetzner Cloud API.

## Upstream Dependencies (What This Resource Needs)

| Dependency | Field | Required | Cardinality | Purpose |
|---|---|---|---|---|
| `HetznerCloudSshKey` | `spec.sshKeys[]` | No | 0..N | SSH keys injected at creation for secure access |
| `HetznerCloudPlacementGroup` | `spec.placementGroupId` | No | 0..1 | Anti-affinity scheduling across physical hosts |
| `HetznerCloudFirewall` | `spec.firewallIds[]` | No | 0..N | Network security rules applied at the infrastructure level |
| `HetznerCloudNetwork` | `spec.networks[].networkId` | No | 0..N | Private network attachments for internal communication |
| `HetznerCloudPrimaryIp` | `spec.publicNet.ipv4`, `spec.publicNet.ipv6` | No | 0..2 | Stable public IPs that survive server replacement |

All dependencies are optional. A server can be created with zero references to other components.

## Downstream Dependents (What References This Resource)

| Dependent | Field | Purpose |
|---|---|---|
| `HetznerCloudVolume` | `spec.serverId` | Block storage attachment to this server |
| `HetznerCloudSnapshot` | `spec.serverId` | Server image snapshot |
| `HetznerCloudFloatingIp` | `spec.serverId` | Reassignable IP assignment to this server |
| `HetznerCloudLoadBalancer` | `spec.targets[].serverId` | Load balancer target |

All dependents reference the server's `server_id` output via `StringValueOrRef`.

## Stack Outputs

| Output | Description |
|---|---|
| `server_id` | Hetzner Cloud numeric ID of the created server (as string). Referenced by Volume, Snapshot, FloatingIp, and LoadBalancer components. |
| `ipv4_address` | The public IPv4 address assigned to the server. Empty if public IPv4 is disabled via `publicNet.ipv4Enabled = false`. |
| `ipv6_address` | The first IPv6 address of the server's assigned /64 network. Empty if public IPv6 is disabled via `publicNet.ipv6Enabled = false`. |
| `status` | The current status of the server: `running`, `off`, `rebuilding`, or `migrating`. |

## References

- [Hetzner Cloud Servers Documentation](https://docs.hetzner.cloud/#servers)
- [Hetzner Cloud Server Types](https://docs.hetzner.cloud/#server-types)
- [Terraform hcloud_server Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/server)
- [Terraform hcloud_rdns Resource](https://registry.terraform.io/providers/hetznercloud/hcloud/latest/docs/resources/rdns)
- [Pulumi hcloud.Server Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/server/)
- [Pulumi hcloud.Rdns Resource](https://www.pulumi.com/registry/packages/hcloud/api-docs/rdns/)
- [Hetzner Cloud Cloud-Init Documentation](https://docs.hetzner.cloud/#servers-create-a-server)
