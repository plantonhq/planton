---
title: "Web Server Firewall"
description: "This preset creates a firewall for public-facing web servers, allowing inbound SSH, HTTP, HTTPS, and ICMP from all IPv4 and IPv6 addresses. It covers the most common Hetzner Cloud server deployment:..."
type: "preset"
rank: "01"
presetSlug: "01-web-server"
componentSlug: "hetzner-cloud-firewall"
componentTitle: "Hetzner Cloud Firewall"
provider: "hetznercloud"
icon: "package"
order: 1
---

# Web Server Firewall

This preset creates a firewall for public-facing web servers, allowing inbound SSH, HTTP, HTTPS, and ICMP from all IPv4 and IPv6 addresses. It covers the most common Hetzner Cloud server deployment: a web application that needs to be reachable on standard ports with SSH for administration.

No outbound rules are included. Hetzner Cloud firewalls allow all outbound traffic by default when no outbound rules are defined. Adding even one outbound rule switches outbound to deny-by-default, which would require explicitly allowing every outbound protocol to avoid accidental lockout.

## When to Use

- Web application servers running behind a reverse proxy or directly serving traffic
- API servers that need to be publicly reachable on HTTP/HTTPS
- Any server that should accept SSH connections and serve web traffic from the public internet

## Key Configuration Choices

- **Inbound SSH** (`port: "22"`, `sourceIps: 0.0.0.0/0, ::/0`) -- allows remote administration from any IP; restrict `sourceIps` to a known CIDR for tighter security
- **Inbound HTTP + HTTPS** (`port: "80"` and `port: "443"`) -- standard web traffic ports open to all sources
- **Inbound ICMP** (`protocol: icmp`) -- enables ping and path MTU discovery, essential for network diagnostics
- **No outbound rules** -- all outbound traffic is allowed by default; avoids consuming rule slots and prevents accidental protocol lockout

## Placeholders to Replace

No placeholders -- this preset is ready to deploy after setting `metadata.name` to the desired firewall name.
