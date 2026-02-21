# Hetzner Cloud Floating IP

Allocates a reassignable public IP address (IPv4 or IPv6 /64) in Hetzner Cloud that can be moved between servers in the same location. The IP is created at a specific home location with an optional server assignment and an optional reverse DNS record. Unlike a Primary IP (which occupies a server's primary public IP slot), a Floating IP is a secondary address suited for failover, high availability, and rolling deployment patterns.

## What Gets Created

- **Floating IP** — an `hcloud_floating_ip` resource allocating either a single IPv4 address or an IPv6 /64 network block at the specified home location, with optional server assignment, standard labels computed from resource metadata, and optional delete protection.
- **Reverse DNS record** (when `dnsPtr` is set) — an `hcloud_rdns` resource mapping the allocated IP address back to the specified hostname. Created only when the `dnsPtr` field is non-empty.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or OpenMCF provider config
- **A server in the same location** if using `serverId` to assign the Floating IP at creation time

## Quick Start

Create a file `floating-ip.yaml`:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFloatingIp
metadata:
  name: my-fip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudFloatingIp.my-fip
spec:
  type: ipv4
  homeLocation: fsn1
```

Deploy:

```shell
openmcf apply -f floating-ip.yaml
```

This allocates a single unassigned IPv4 address in Falkenstein. The address is assigned by Hetzner Cloud and returned in `status.outputs.ip_address`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `type` | `enum` (`ipv4`, `ipv6`) | IP address type. `ipv4` allocates a single address. `ipv6` allocates a /64 network block. Changing this forces replacement. | Required, defined values only |
| `homeLocation` | `string` | Hetzner Cloud location where the IP is homed. Known locations: `fsn1`, `nbg1`, `hel1`, `ash`, `hil`, `sin`. The Floating IP can only be assigned to servers in the same location. Changing this forces replacement. | `min_len: 1` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `description` | `string` | empty | Human-readable description for the Floating IP. Visible in the Hetzner Cloud console and API. |
| `serverId` | `StringValueOrRef` | unset | Server to assign this Floating IP to. Accepts a literal Hetzner Cloud server ID (as a string) or a reference to a `HetznerCloudServer` resource via `valueFrom`. The server must be in the same location as `homeLocation`. If omitted, the IP is created unassigned. |
| `dnsPtr` | `string` | empty | Reverse DNS pointer record for the allocated IP address. When set, an `hcloud_rdns` resource is created mapping the IP back to this hostname. Required for mail servers (SPF/DKIM verification relies on matching forward and reverse DNS). |
| `deleteProtection` | `bool` | `false` | Prevent accidental deletion of the Floating IP via the Hetzner Cloud API. Must be disabled before the IP can be deleted. |

## Examples

### Minimal IPv4

A single unassigned IPv4 address in Falkenstein — the simplest working configuration.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFloatingIp
metadata:
  name: dev-fip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudFloatingIp.dev-fip
spec:
  type: ipv4
  homeLocation: fsn1
```

### Mail Server IP with Reverse DNS

An IPv4 address with a reverse DNS pointer for email deliverability. The `dnsPtr` hostname must have a matching forward DNS A record.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFloatingIp
metadata:
  name: mail-fip
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudFloatingIp.mail-fip
spec:
  type: ipv4
  homeLocation: fsn1
  description: Production mail server failover IP
  dnsPtr: mail.example.com
  deleteProtection: true
```

### IPv6 with Delete Protection

An IPv6 /64 block in Helsinki with delete protection enabled.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFloatingIp
metadata:
  name: web-ipv6
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudFloatingIp.web-ipv6
spec:
  type: ipv6
  homeLocation: hel1
  description: IPv6 block for web frontends
  deleteProtection: true
```

### Server Composition via valueFrom

A Floating IP assigned to a HetznerCloudServer using `valueFrom`. The Floating IP receives the server's numeric ID from the server's stack outputs, establishing a dependency edge in the deployment DAG.

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudFloatingIp
metadata:
  name: app-failover
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudFloatingIp.app-failover
spec:
  type: ipv4
  homeLocation: fsn1
  description: Application failover IP
  serverId:
    valueFrom:
      kind: HetznerCloudServer
      name: app-01
      fieldPath: status.outputs.server_id
  dnsPtr: app.example.com
  deleteProtection: true
```

The server referenced by this Floating IP:

```yaml
apiVersion: hetzner-cloud.openmcf.org/v1
kind: HetznerCloudServer
metadata:
  name: app-01
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudServer.app-01
spec:
  serverType: cx22
  image: ubuntu-24.04
  location: fsn1
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `floating_ip_id` | `string` | Hetzner Cloud numeric ID of the created Floating IP. Can be used for monitoring or external automation. |
| `ip_address` | `string` | The allocated IP address. For IPv4, a single address (e.g., `203.0.113.42`). For IPv6, the first address in the /64 block (e.g., `2001:db8::1`). |
| `ip_network` | `string` | The allocated IPv6 /64 CIDR (e.g., `2001:db8::/64`). Empty for IPv4 Floating IPs. |

## Related Components

- [HetznerCloudServer](/docs/catalog/hetznercloud/hetznercloudserver) — The server the Floating IP can be assigned to via `serverId`
- [HetznerCloudPrimaryIp](/docs/catalog/hetznercloud/hetznercloudprimaryip) — Alternative IP type that occupies a server's primary public IP slot (location-bound, automatic OS configuration)
- [HetznerCloudFirewall](/docs/catalog/hetznercloud/hetznercloudfirewall) — Controls inbound/outbound traffic for servers using this IP
- [HetznerCloudNetwork](/docs/catalog/hetznercloud/hetznercloudnetwork) — Private networking commonly deployed alongside public IPs
