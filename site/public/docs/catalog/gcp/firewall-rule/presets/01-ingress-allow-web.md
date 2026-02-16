---
title: "Ingress Allow Web (HTTP/HTTPS)"
description: "This preset creates a standard web server firewall rule that allows inbound HTTP (port 80) and HTTPS (port 443) traffic from all IPv4 addresses. This is the most common firewall rule for any..."
type: "preset"
rank: "01"
presetSlug: "01-ingress-allow-web"
componentSlug: "firewall-rule"
componentTitle: "Firewall Rule"
provider: "gcp"
icon: "package"
order: 1
---

# Ingress Allow Web (HTTP/HTTPS)

This preset creates a standard web server firewall rule that allows inbound HTTP (port 80) and HTTPS (port 443) traffic from all IPv4 addresses. This is the most common firewall rule for any internet-facing web application, load balancer frontend, or reverse proxy.

## When to Use

- Public-facing web servers, API gateways, or load balancer backends
- Any service that needs to accept HTTP/HTTPS traffic from the internet
- Development or staging environments where you want quick public access (consider restricting `sourceRanges` for staging)

## Key Configuration Choices

- **INGRESS + ALLOW** -- permits inbound TCP traffic on ports 80 and 443
- **Source range `0.0.0.0/0`** -- allows traffic from any IPv4 address; this is intentional for public web services but should be narrowed for internal-only services
- **Priority 1000** -- the GCP default; sits in the standard application rule range, leaving room for higher-priority emergency overrides (0-999) and lower-priority baselines (65000+)
- **No target tags or service accounts** -- rule applies to all instances in the VPC network; add `targetTags` or `targetServiceAccounts` to restrict which VMs are affected

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the firewall rule will be created | GCP Console or `GcpProject` outputs |
| `<vpc-network-name-or-self-link>` | VPC network name (e.g., `default`) or full self-link | `GcpVpc` status outputs |
| `<your-rule-name>` | Unique name for this firewall rule (1-63 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `allow-http-https-ingress`) |

## Related Presets

- **02-ingress-allow-ssh-iap** -- Allow SSH access via Google Identity-Aware Proxy (secure remote access)
- **03-egress-deny-all** -- Deny all outbound traffic as a restrictive baseline
