---
title: "Private Network Firewall"
description: "This preset creates a firewall that restricts all inbound traffic to a single private CIDR range, blocking all access from the public internet. Only SSH and ICMP are permitted, and only from hosts..."
type: "preset"
rank: "03"
presetSlug: "03-private-network"
componentSlug: "firewall"
componentTitle: "Firewall"
provider: "hetznercloud"
icon: "package"
order: 3
---

# Private Network Firewall

This preset creates a firewall that restricts all inbound traffic to a single private CIDR range, blocking all access from the public internet. Only SSH and ICMP are permitted, and only from hosts within the specified private network. Servers behind this firewall are reachable exclusively via a bastion host, VPN, or Hetzner Cloud private network attachment.

This is the baseline firewall for internal services. Add additional rules for your application's specific ports (e.g., PostgreSQL 5432, MySQL 3306, Redis 6379) using the same private CIDR as the source.

## When to Use

- Database servers that should never be directly reachable from the public internet
- Internal API backends, message queues, or cache layers accessed only by other servers in the same private network
- Any workload where a zero-public-exposure posture is a hard requirement

## Key Configuration Choices

- **Private CIDR only** (`sourceIps: <private-network-cidr>`) -- no `0.0.0.0/0` or `::/0` entries, so the public internet cannot reach the server
- **SSH from private network** (`port: "22"`) -- administration requires routing through a bastion host or VPN connected to the private network
- **No application ports** -- start locked down and add rules for your specific service ports; this avoids shipping a preset that guesses your application's protocol
- **No outbound rules** -- unrestricted egress allows the server to reach external services for updates and dependencies

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<private-network-cidr>` | CIDR block of your Hetzner Cloud private network (e.g., `10.0.0.0/16`) | The `ipRange` field of your HetznerCloudNetwork resource, or the Network details page in the Hetzner Cloud Console |
