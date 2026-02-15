---
title: "Development Droplet"
description: "This preset creates a minimal DigitalOcean Droplet for development and testing. It uses the smallest general-purpose instance with no backups, keeping costs low while still providing VPC isolation."
type: "preset"
rank: "02"
presetSlug: "02-development"
componentSlug: "droplet"
componentTitle: "Droplet"
provider: "digitalocean"
icon: "package"
order: 2
---

# Development Droplet

This preset creates a minimal DigitalOcean Droplet for development and testing. It uses the smallest general-purpose instance with no backups, keeping costs low while still providing VPC isolation.

## When to Use

- Development, staging, or testing environments
- Short-lived instances for CI/CD build agents
- Prototyping and experimentation

## Key Configuration Choices

- **Minimal sizing** (`size: s-1vcpu-1gb`) -- smallest general-purpose Droplet. Sufficient for light dev workloads and testing.
- **No backups** -- `enableBackups` omitted (defaults to `false`). Dev environments are ephemeral and can be recreated.
- **VPC isolation** (`vpc`) -- still placed in a VPC for network consistency between dev and production.
- **Development tag** -- enables separate firewall rules for dev environments.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|-------------|-------------|---------------|
| `<vpc-id>` | UUID of the target VPC | DigitalOcean VPC console or `DigitalOceanVpc` status outputs |
| `nyc1` | Target DigitalOcean region slug | Must match the VPC's region |

## Related Presets

- **01-production** -- Use instead for production workloads requiring backups and larger instance sizing
