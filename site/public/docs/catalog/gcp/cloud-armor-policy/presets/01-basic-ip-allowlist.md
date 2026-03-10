---
title: "Preset: Basic IP Allowlist"
description: "Restrict access to your backend to known corporate or VPN IP ranges. All traffic from allowlisted CIDR blocks is permitted; everything else receives a 403. Ideal for internal dashboards, admin..."
type: "preset"
rank: "01"
presetSlug: "01-basic-ip-allowlist"
componentSlug: "cloud-armor-policy"
componentTitle: "Cloud Armor Policy"
provider: "gcp"
icon: "package"
order: 1
---

# Preset: Basic IP Allowlist

## Use Case

Restrict access to your backend to known corporate or VPN IP ranges. All traffic from allowlisted CIDR blocks is permitted; everything else receives a 403. Ideal for internal dashboards, admin panels, or APIs that should only be reachable from your office or VPN.

## When to Use

- Internal admin UIs or dashboards
- APIs consumed only by trusted networks (office, VPN, partner data centers)
- Backends that must not be exposed to the public internet

## What This Creates

- A Cloud Armor policy with two rules
- Allow rule (priority 1000): permits traffic from `10.0.0.0/8` (RFC 1918 private) and `172.16.0.0/12` (RFC 1918 private)
- Default deny rule (priority 2147483647): blocks all other IPs with 403 Forbidden

Rule 2147483647 is Cloud Armor's reserved default priority; placing a deny there ensures any IP not matched by earlier rules is blocked.

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `srcIpRanges` | `10.0.0.0/8`, `172.16.0.0/12` | Add your VPN egress CIDRs, data center ranges, or partner IPs. Max 10 ranges per rule. |
| `action` (deny rule) | `deny(403)` | Use `deny(404)` to obscure existence of the resource, or `deny(502)` to mimic backend failure. |
| `projectId.value` | `my-gcp-project` | Your GCP project ID. |

To add more allowed ranges, extend the allow rule’s `srcIpRanges`. To split rules by priority (e.g., a higher-priority allow for a specific /24), add additional rules with lower priority numbers.
