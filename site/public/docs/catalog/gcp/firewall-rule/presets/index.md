---
title: "Presets"
description: "Ready-to-deploy configuration presets for Firewall Rule"
type: "preset-list"
componentSlug: "firewall-rule"
componentTitle: "Firewall Rule"
provider: "gcp"
icon: "package"
order: 200
presets:
  - slug: "01-ingress-allow-web"
    rank: "01"
    title: "Ingress Allow Web (HTTP/HTTPS)"
    excerpt: "This preset creates a standard web server firewall rule that allows inbound HTTP (port 80) and HTTPS (port 443) traffic from all IPv4 addresses. This is the most common firewall rule for any..."
  - slug: "02-ingress-allow-ssh-iap"
    rank: "02"
    title: "Ingress Allow SSH via IAP"
    excerpt: "This preset creates a firewall rule that allows SSH (port 22) access exclusively from Google's Identity-Aware Proxy (IAP) IP range (`35.235.240.0/20`). IAP provides authenticated, audited SSH access..."
  - slug: "03-egress-deny-all"
    rank: "03"
    title: "Egress Deny All"
    excerpt: "This preset creates a restrictive egress baseline by denying all outbound traffic from VMs in the VPC network. GCP's implied rules allow all egress at priority 65535 by default. This rule overrides..."
---

# Firewall Rule Presets

Ready-to-deploy configuration presets for Firewall Rule. Each preset is a complete manifest you can copy, customize, and deploy.
