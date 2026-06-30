# OciNetworkFirewall

## Overview

OciNetworkFirewall is an Planton component that deploys an OCI Network Firewall with an inline firewall policy. It provides a single declarative manifest to create a network firewall appliance, its policy, and all policy sub-resources (address lists, services, service lists, URL lists, security rules).

## Purpose

OCI Network Firewall is a managed next-generation firewall service that inspects traffic at Layer 3/4 (IP/port) and Layer 7 (URL patterns). The firewall appliance sits in a subnet and inspects traffic routed through it via VCN route table entries. This component bundles the firewall, its policy, and all policy objects into one manifest because the policy and its sub-resources form an inseparable unit — a firewall without a policy cannot inspect traffic, and policy objects are meaningless without the firewall.

## Key Features

- **Inline policy declaration** — the firewall policy and all sub-resources are declared in one manifest, providing a complete picture of the security posture.
- **Address lists** — named collections of IP CIDRs or FQDNs referenced by security rules for source/destination matching.
- **Service definitions** — named TCP/UDP port range definitions for traffic matching.
- **Service lists** — groups of services for reuse across multiple security rules.
- **URL lists** — URL pattern lists for L7 HTTP(S) traffic inspection.
- **Security rules** — ordered rules with actions (allow, drop, reject, inspect) evaluated by list position. Priority is derived from position — first rule has highest priority.
- **IDS/IPS inspection** — intrusion detection and intrusion prevention modes for the `inspect` action.
- **Name-based cross-references** — security rules reference address lists, services, and URL lists by name within the same manifest.
- **Foreign key references** — `compartmentId`, `subnetId`, and `networkSecurityGroupIds` support `valueFrom`.

## Constraints

- `subnetId`, `ipv4Address`, `ipv6Address`, and `availabilityDomain` are ForceNew — changing them forces firewall recreation.
- Policy sub-resource names are ForceNew — renaming forces recreation.
- Security rules are evaluated in list order; priority is derived from position (1-based).
- The `inspect` action requires an `inspection` type (`intrusion_detection` or `intrusion_prevention`).
- All condition fields within a security rule are AND-ed; values within each field are OR-ed.
- Sub-resources reference each other by name — names must be consistent between definitions and rule conditions.

## Use Cases

| Scenario | Configuration |
|----------|---------------|
| Basic allow/deny firewall | Address lists + allow/drop rules |
| Web application protection | Services (HTTP/HTTPS) + URL lists + allow/block rules |
| Intrusion detection | `inspect` action with `intrusion_detection` on inbound traffic |
| Intrusion prevention | `inspect` action with `intrusion_prevention` for active blocking |
| Egress filtering | Address lists + URL lists to control outbound traffic |
| Microsegmentation | Address lists per application tier with strict allow rules |

## Production Features

- **Freeform tags** — automatically populated on both the firewall and its policy from `metadata.labels`.
- **IPv4 address output** — the firewall's IP is exported for configuring VCN route table entries.
- **Policy ID output** — enables advanced scenarios where the policy is referenced externally.
- **NSG binding** — optional network security groups for additional network access control on the firewall appliance.
