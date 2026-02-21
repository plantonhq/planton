---
title: "Preset: TCPSSL Production"
description: "- Production internet-facing NLB with TLS termination at Layer 4. - Fixed public IPs via EIP binding for DNS A-records and firewall whitelisting. - Database or API traffic that needs encryption..."
type: "preset"
rank: "03"
presetSlug: "03-tcpssl-production"
componentSlug: "nlb-load-balancer"
componentTitle: "NLB Load Balancer"
provider: "alicloud"
icon: "package"
order: 3
---

# Preset: TCPSSL Production

## When to Use

- Production internet-facing NLB with TLS termination at Layer 4.
- Fixed public IPs via EIP binding for DNS A-records and firewall whitelisting.
- Database or API traffic that needs encryption without HTTP overhead.

## What It Creates

- Internet-facing NLB with fixed EIPs in each availability zone
- TCPSSL server group with connection draining (300s) and Wlc (weighted least connections) scheduling
- HTTP health checks on `/healthz` for application-level probing
- TCPSSL listener on port 443 with TLS 1.2 strict cipher policy

## Customization Points

- Replace `<placeholders>` with actual resource references and certificate IDs
- Add `caCertificateIds` and `caEnabled: true` for mutual TLS
- Adjust `connectionDrainTimeout` based on workload characteristics
- Add additional listeners for other ports
- Tune `healthCheckInterval` and thresholds for your SLA requirements
