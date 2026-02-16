---
title: "Simple HTTP Load Balancer"
description: "This preset creates a basic HTTP load balancer that forwards port 80 traffic to backend Droplets on port 8080. Uses explicit droplet IDs for targeting. Health checks ensure only healthy backends..."
type: "preset"
rank: "02"
presetSlug: "02-http-basic"
componentSlug: "load-balancer"
componentTitle: "Load Balancer"
provider: "digitalocean"
icon: "package"
order: 2
---

# Simple HTTP Load Balancer

This preset creates a basic HTTP load balancer that forwards port 80 traffic to backend Droplets on port 8080. Uses explicit droplet IDs for targeting. Health checks ensure only healthy backends receive traffic. No SSL—suitable for development, staging, or internal services.

## When to Use

- Development or staging environments where HTTPS is not required
- Internal services behind a private load balancer
- Explicit control over which Droplets are in the pool (vs tag-based)
- Simple HTTP applications

## Key Configuration Choices

- **HTTP only** (`entryPort: 80`, `entryProtocol: http`) -- no TLS; use behind a CDN or reverse proxy for production HTTPS.
- **Explicit droplet IDs** (`dropletIds`) -- list Droplet IDs or references; mutually exclusive with `dropletTag`. Use references to `DigitalOceanDroplet` for IaC.
- **Port 8080** (`targetPort: 8080`) -- common app port; change to match your application.
- **Health check** (`path: /health`) -- backends must respond 2xx; ensure your app exposes this endpoint.
- **VPC required** (`vpc`) -- load balancer must be in the same VPC as Droplets.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-id>` | UUID of the target VPC | DigitalOcean VPC console or `DigitalOceanVpc` status outputs |
| `<droplet-id-1>`, `<droplet-id-2>` | Droplet IDs to attach | DigitalOcean console or `DigitalOceanDroplet` status outputs |
| `nyc3` | Target DigitalOcean region slug | Must match the VPC's region |

## Related Presets

- **01-https-ssl-termination** -- Use for production HTTPS with tag-based targeting
