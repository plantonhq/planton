---
title: "Deny-All Allowlist Security Group"
description: "This preset creates a strict security group for internal services. All inbound traffic is dropped by default, with rules accepting only TCP traffic from the private network range (10.0.0.0/8) and SSH..."
type: "preset"
rank: "02"
presetSlug: "02-deny-all-allowlist"
componentSlug: "instance-security-group"
componentTitle: "Instance Security Group"
provider: "scaleway"
icon: "package"
order: 2
---

# Deny-All Allowlist Security Group

This preset creates a strict security group for internal services. All inbound traffic is dropped by default, with rules accepting only TCP traffic from the private network range (10.0.0.0/8) and SSH from a specific bastion host. This is the standard pattern for databases, caches, message queues, and backend workers that should never be directly internet-accessible.

## When to Use

- Database servers (PostgreSQL, MySQL, MongoDB) accessible only from application instances
- Cache servers (Redis, Memcached) on the Private Network
- Backend workers or internal APIs that receive traffic only from other services
- Any instance that should accept traffic exclusively from known private sources

## Key Configuration Choices

- **Allowlist model** (`inboundDefaultPolicy: drop`) -- all inbound traffic blocked unless explicitly accepted
- **Private network access** (`ipRange: 10.0.0.0/8`) -- allows all TCP traffic from RFC 1918 private addresses; adjust to your Private Network CIDR for tighter control
- **Bastion SSH** (`portRange: "22"`, `ipRange: <bastion-ip>/32`) -- SSH access only from a single bastion host IP
- **All outbound allowed** (`outboundDefaultPolicy: accept`) -- instances can reach external services for updates and health checks
- **Stateful** (`stateful: true`) -- return traffic for accepted connections is automatically permitted

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-bastion-ip>` | IP address of the bastion host or Public Gateway (e.g., `10.0.1.1`) | `ScalewayPublicGateway` status outputs or Scaleway console |

## Related Presets

- **01-web-server** -- Use instead for public-facing web servers that need HTTP/HTTPS open to the internet
