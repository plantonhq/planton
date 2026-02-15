---
title: "HTTPS Load Balancer with Let's Encrypt"
description: "This preset creates a Scaleway Load Balancer with automatic TLS certificate provisioning via Let's Encrypt, HTTP health checks, and two frontends (HTTPS on 443 and HTTP on 80). The LB is attached to..."
type: "preset"
rank: "01"
presetSlug: "01-https-letsencrypt"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "scaleway"
icon: "package"
order: 1
---

# HTTPS Load Balancer with Let's Encrypt

This preset creates a Scaleway Load Balancer with automatic TLS certificate provisioning via Let's Encrypt, HTTP health checks, and two frontends (HTTPS on 443 and HTTP on 80). The LB is attached to a Private Network so backend servers are not exposed to the public internet. This is the standard production configuration for web applications.

## When to Use

- Public-facing web applications or APIs requiring HTTPS
- Production workloads where backends run on a Private Network
- Any service that needs automatic TLS certificate management without manual renewal

## Key Configuration Choices

- **Small tier** (`type: LB-S`) -- up to 400 Mbps; sufficient for most web applications. Upgrade to `LB-GP-M` for high-traffic services
- **Private Network attached** (`privateNetworkId`) -- backends are reached via private IPs; this is the recommended production topology
- **Two backend servers** -- minimum for availability; the round-robin algorithm distributes connections evenly
- **HTTP health check** (`healthCheck.type: http`) -- probes `/health` expecting a 200 response; more reliable than TCP checks for web applications
- **HTTPS frontend on 443** -- terminates TLS using the Let's Encrypt certificate
- **HTTP frontend on 80** -- allows plaintext traffic; configure your application to redirect HTTP to HTTPS if desired
- **Let's Encrypt certificate** -- auto-provisioned and auto-renewed; the domain must resolve to the LB's public IP before deployment

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| ----------- | ----------- | ------------- |
| `<your-private-network-id>` | UUID of the Private Network the LB attaches to | Scaleway console or `ScalewayPrivateNetwork` status outputs |
| `<backend-server-ip-1>` | Private IP of the first backend server (e.g., `10.0.1.5`) | Scaleway console or `ScalewayInstance` status outputs |
| `<backend-server-ip-2>` | Private IP of the second backend server (e.g., `10.0.1.6`) | Scaleway console or `ScalewayInstance` status outputs |
| `<your-domain.com>` | Domain name for the TLS certificate (e.g., `app.example.com`) | Your domain registrar; must resolve to the LB's IP |

## Related Presets

- **02-http-simple** -- Use instead for development or internal services that do not need TLS
