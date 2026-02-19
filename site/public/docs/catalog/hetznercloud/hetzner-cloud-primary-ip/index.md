---
title: "Hetzner Cloud Primary IP"
description: "Hetzner Cloud Primary IP deployment documentation"
icon: "package"
order: 100
componentName: "hetznercloudprimaryip"
---

# Hetzner Cloud Primary IP

Allocates a persistent public IP address (IPv4 or IPv6 /64) in Hetzner Cloud that survives server deletion. The IP is created at a specific location and can be assigned to servers via `HetznerCloudServer`. An optional reverse DNS record can be configured for mail servers and services requiring identity verification through reverse lookups.

## What Gets Created

- **Primary IP** — an `hcloud_primary_ip` resource allocating either a single IPv4 address or an IPv6 /64 network block at the specified location, with standard labels computed from resource metadata and optional delete protection. `auto_delete` is always `false` and `assignee_type` is always `"server"`.
- **Reverse DNS record** (when `dnsPtr` is set) — an `hcloud_rdns` resource mapping the allocated IP address back to the specified hostname. Created only when the `dnsPtr` field is non-empty.

## Prerequisites

- **Hetzner Cloud API token** configured via environment variable (`HCLOUD_TOKEN`) or OpenMCF provider config

## Quick Start

Create a file `primary-ip.yaml`:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPrimaryIp
metadata:
  name: my-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudPrimaryIp.my-ip
spec:
  type: ipv4
  location: fsn1
```

Deploy:

```shell
openmcf apply -f primary-ip.yaml
```

This allocates a single IPv4 address in Falkenstein. The address is assigned by Hetzner Cloud and returned in `status.outputs.ip_address`.

## Configuration Reference

### Required Fields

| Field | Type | Description | Validation |
|-------|------|-------------|------------|
| `type` | `enum` (`ipv4`, `ipv6`) | IP address type. `ipv4` allocates a single address. `ipv6` allocates a /64 network block. Changing this forces replacement. | Required, defined values only |
| `location` | `string` | Hetzner Cloud location where the IP is allocated. Known locations: `fsn1`, `nbg1`, `hel1`, `ash`, `hil`, `sin`. The Primary IP can only be assigned to servers in the same location. Changing this forces replacement. | `min_len: 1` |

### Optional Fields

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `dnsPtr` | `string` | empty | Reverse DNS pointer record for the allocated IP address. When set, an `hcloud_rdns` resource is created mapping the IP back to this hostname. Required for mail servers (SPF/DKIM verification relies on matching forward and reverse DNS). |
| `deleteProtection` | `bool` | `false` | Prevent accidental deletion of the Primary IP via the Hetzner Cloud API. Must be disabled before the IP can be deleted. |

## Examples

### Minimal IPv4

A single IPv4 address in Falkenstein — the simplest working configuration.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPrimaryIp
metadata:
  name: dev-ip
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: my-org
    pulumi.openmcf.org/project: my-project
    pulumi.openmcf.org/stack.name: dev.HetznerCloudPrimaryIp.dev-ip
spec:
  type: ipv4
  location: fsn1
```

### Mail Server IP with Reverse DNS

An IPv4 address with a reverse DNS pointer for email deliverability. The `dnsPtr` hostname must have a matching forward DNS A record.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPrimaryIp
metadata:
  name: mail-ip
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudPrimaryIp.mail-ip
spec:
  type: ipv4
  location: fsn1
  dnsPtr: mail.example.com
  deleteProtection: true
```

### IPv6 with Delete Protection

An IPv6 /64 block in Helsinki with delete protection enabled.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPrimaryIp
metadata:
  name: web-ipv6
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudPrimaryIp.web-ipv6
spec:
  type: ipv6
  location: hel1
  deleteProtection: true
```

### Server Composition via valueFrom

A Primary IP referenced by a HetznerCloudServer using `valueFrom`. The server receives the IP's numeric ID from the Primary IP's stack outputs, establishing a dependency edge in the deployment DAG.

```yaml
apiVersion: hetznercloud.openmcf.org/v1
kind: HetznerCloudPrimaryIp
metadata:
  name: app-ip
  org: acme-corp
  env: production
  labels:
    openmcf.org/provisioner: pulumi
    pulumi.openmcf.org/organization: acme-corp
    pulumi.openmcf.org/project: infrastructure
    pulumi.openmcf.org/stack.name: production.HetznerCloudPrimaryIp.app-ip
spec:
  type: ipv4
  location: fsn1
  dnsPtr: app.example.com
  deleteProtection: true
```

The server references this IP:

```yaml
apiVersion: hetznercloud.openmcf.org/v1
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
  primaryIpId:
    valueFrom:
      kind: HetznerCloudPrimaryIp
      name: app-ip
      fieldPath: status.outputs.primary_ip_id
```

## Stack Outputs

| Output | Type | Description |
|--------|------|-------------|
| `primary_ip_id` | `string` | Hetzner Cloud numeric ID of the created Primary IP. Referenced by HetznerCloudServer via `StringValueOrRef`. |
| `ip_address` | `string` | The allocated IP address. For IPv4, a single address (e.g., `203.0.113.42`). For IPv6, the first address in the /64 block (e.g., `2001:db8::1`). |
| `ip_network` | `string` | The allocated IPv6 /64 CIDR (e.g., `2001:db8::/64`). Empty for IPv4 Primary IPs. |

## Related Components

- [HetznerCloudServer](/docs/catalog/hetznercloud/hetzner-cloud-server) — Assigns the Primary IP to a server's primary public IP slot via `primaryIpId`
- [HetznerCloudFloatingIp](/docs/catalog/hetznercloud/hetzner-cloud-floating-ip) — Alternative IP type that can be reassigned between servers in any location (not location-bound like Primary IPs)
- [HetznerCloudFirewall](/docs/catalog/hetznercloud/hetzner-cloud-firewall) — Controls inbound/outbound traffic for servers using this IP
- [HetznerCloudNetwork](/docs/catalog/hetznercloud/hetzner-cloud-network) — Private networking commonly deployed alongside public IPs for server connectivity
