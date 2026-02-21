---
title: "SSH-Only Firewall"
description: "This preset creates a minimal-surface firewall that allows only SSH and ICMP inbound. All other inbound traffic is dropped by Hetzner Cloud's deny-by-default policy. This is the tightest useful..."
type: "preset"
rank: "02"
presetSlug: "02-ssh-only"
componentSlug: "firewall"
componentTitle: "Firewall"
provider: "hetznercloud"
icon: "package"
order: 2
---

# SSH-Only Firewall

This preset creates a minimal-surface firewall that allows only SSH and ICMP inbound. All other inbound traffic is dropped by Hetzner Cloud's deny-by-default policy. This is the tightest useful firewall for a server that still needs remote access.

No outbound rules are included, so the server can freely reach external services for package updates, API calls, and forwarded traffic.

## When to Use

- Bastion / jump hosts that serve as the single SSH entry point to a private network
- CI runners, build agents, or automation servers with no public-facing services
- Utility servers used for administration, monitoring, or tooling

## Key Configuration Choices

- **SSH only** (`port: "22"`) -- the single exposed TCP port minimizes the attack surface
- **ICMP retained** (`protocol: icmp`) -- ping and path MTU discovery remain available for diagnostics
- **No outbound rules** -- unrestricted egress allows the server to reach package repositories, APIs, and private network peers

## Placeholders to Replace

No placeholders -- this preset is ready to deploy after setting `metadata.name` to the desired firewall name.
