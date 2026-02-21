---
title: "Private Backend Server"
description: "This preset creates a Hetzner Cloud server with no public IP address, reachable only through a private network. Both public IPv4 and IPv6 are explicitly disabled via the `publicNet` block, making the..."
type: "preset"
rank: "03"
presetSlug: "03-private-backend"
componentSlug: "hetzner-cloud-server"
componentTitle: "Hetzner Cloud Server"
provider: "hetznercloud"
icon: "package"
order: 3
---

# Private Backend Server

This preset creates a Hetzner Cloud server with no public IP address, reachable only through a private network. Both public IPv4 and IPv6 are explicitly disabled via the `publicNet` block, making the server invisible to the internet. All production protections (backups, delete/rebuild protection, graceful shutdown) are enabled.

This is the correct configuration for databases, message queues, caches, internal APIs, and any workload where a zero-public-exposure posture is a hard requirement. SSH access requires routing through a bastion host, VPN, or Hetzner Cloud Console's built-in rescue console.

## When to Use

- Database servers (PostgreSQL, MySQL, Redis, MongoDB) that must not accept connections from the public internet
- Internal microservices and API backends accessed exclusively by other servers in the same private network
- Message brokers, task queues, and cache layers (RabbitMQ, Kafka, Memcached) that serve only internal consumers
- Any server subject to compliance requirements mandating no public network interfaces

## Key Configuration Choices

- **Public networking disabled** (`publicNet.ipv4Enabled: false`, `publicNet.ipv6Enabled: false`) -- the server has no public IP and cannot be reached from or initiate connections to the internet; this is the key differentiator from the other server presets
- **Private network required** (`networks`) -- the only way to communicate with this server; pair with the HetznerCloudNetwork `01-single-zone` preset and the HetznerCloudFirewall `03-private-network` firewall preset
- **No user data** -- backend workloads vary too widely (database engines, queue brokers, custom applications) to ship a useful default cloud-init script; add your own provisioning in the `userData` field
- **No rDNS** -- reverse DNS requires a public IP, which this server does not have; if you need internal DNS, configure it at the application or network level
- **Daily backups enabled** (`backups: true`) -- critical for stateful backend services where data loss means business loss
- **All protections enabled** -- delete protection, rebuild protection, and graceful shutdown protect stateful services from accidental destruction

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<ssh-key-name>` | Name or numeric ID of an SSH key registered in Hetzner Cloud | The `metadata.name` of your HetznerCloudSshKey resource, or the SSH Keys page in the Hetzner Cloud Console |
| `<firewall-id>` | Numeric ID of a Hetzner Cloud firewall (use a private-network firewall) | The `status.outputs.firewall_id` of your HetznerCloudFirewall resource; the `03-private-network` firewall preset is designed for this use case |
| `<network-id>` | Numeric ID of a Hetzner Cloud network | The `status.outputs.network_id` of your HetznerCloudNetwork resource, or the Networks page in the Hetzner Cloud Console |

## Related Presets

- **01-quick-start** -- minimal server with public IPs for development and experimentation
- **02-production-web** -- production server with public IPs for web-facing workloads
