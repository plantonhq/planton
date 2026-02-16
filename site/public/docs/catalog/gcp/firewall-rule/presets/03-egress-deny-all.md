---
title: "Egress Deny All"
description: "This preset creates a restrictive egress baseline by denying all outbound traffic from VMs in the VPC network. GCP's implied rules allow all egress at priority 65535 by default. This rule overrides..."
type: "preset"
rank: "03"
presetSlug: "03-egress-deny-all"
componentSlug: "firewall-rule"
componentTitle: "Firewall Rule"
provider: "gcp"
icon: "package"
order: 3
---

# Egress Deny All

This preset creates a restrictive egress baseline by denying all outbound traffic from VMs in the VPC network. GCP's implied rules allow all egress at priority 65535 by default. This rule overrides that default at priority 65534, establishing a deny-by-default egress posture. Specific outbound destinations are then permitted by adding higher-priority allow rules.

## When to Use

- Production environments following the principle of least privilege for outbound traffic
- Compliance-driven environments (PCI-DSS, HIPAA, SOC 2) that require explicit egress whitelisting
- Networks where you want to prevent unexpected outbound connections (data exfiltration, C2 callbacks, accidental external API calls)
- As a baseline rule paired with specific egress allow rules for approved destinations (e.g., Google APIs, package registries, partner endpoints)

## Key Configuration Choices

- **EGRESS + DENY** -- blocks all outbound traffic from VMs in the network
- **Protocol `all`** -- matches every IP protocol (TCP, UDP, ICMP, etc.), not just specific ports
- **Destination `0.0.0.0/0`** -- matches all IPv4 destinations
- **Priority 65534** -- the lowest usable priority before GCP's implied rules at 65535; any rule with a lower priority number (higher precedence) will override this deny, allowing you to build an explicit allowlist on top
- **No target tags or service accounts** -- applies network-wide; this is intentional for a baseline deny rule

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<gcp-project-id>` | GCP project ID where the firewall rule will be created | GCP Console or `GcpProject` outputs |
| `<vpc-network-name-or-self-link>` | VPC network name (e.g., `default`) or full self-link | `GcpVpc` status outputs |
| `<your-rule-name>` | Unique name for this firewall rule (1-63 chars, lowercase, hyphens) | Choose a descriptive name (e.g., `deny-all-egress`) |

## Related Presets

- **01-ingress-allow-web** -- Allow HTTP/HTTPS traffic from the internet for web servers
- **02-ingress-allow-ssh-iap** -- Allow SSH access via Google Identity-Aware Proxy
