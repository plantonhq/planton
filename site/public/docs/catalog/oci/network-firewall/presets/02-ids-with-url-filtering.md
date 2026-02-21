---
title: "IDS-Enabled Firewall with URL Filtering"
description: "This preset creates an OCI Network Firewall with a policy that combines L4 traffic control, L7 URL-based filtering, and intrusion detection. Malicious URLs are rejected before reaching application..."
type: "preset"
rank: "02"
presetSlug: "02-ids-with-url-filtering"
componentSlug: "network-firewall"
componentTitle: "Network Firewall"
provider: "oci"
icon: "package"
order: 2
---

# IDS-Enabled Firewall with URL Filtering

This preset creates an OCI Network Firewall with a policy that combines L4 traffic control, L7 URL-based filtering, and intrusion detection. Malicious URLs are rejected before reaching application servers, inbound web traffic is inspected for known attack signatures using OCI's built-in IDS engine, DNS queries are allowed, and a default-deny rule drops everything else. This is the recommended configuration for regulated environments requiring deep packet inspection and threat intelligence enforcement.

## When to Use

- PCI-DSS, HIPAA, or SOC 2 environments requiring next-generation firewall capabilities with IDS
- Architectures that need L7 URL filtering to block known malicious domains and URL patterns
- Security postures where traffic must be inspected for exploitation attempts (SQL injection, XSS, known CVEs)
- Perimeter or inter-subnet firewalls in environments with strict compliance audit requirements

## Key Configuration Choices

- **FQDN address list** (`blocked-destinations` with type `fqdn`) -- blocks resolution and connections to known malicious fully qualified domain names. Replace the example entries with your organization's threat intelligence feeds.
- **URL filtering** (`blocked-urls` URL list) -- rejects HTTP/HTTPS requests matching URL patterns at L7. URL lists use simple pattern matching (wildcards supported). This is evaluated before IDS inspection for efficiency.
- **Intrusion detection** (`action: inspect`, `inspection: intrusion_detection`) -- inbound web traffic is analyzed against OCI's managed signature database for known attack patterns. IDS mode logs detected threats without blocking; switch to `intrusion_prevention` to automatically block detected attacks (may cause false-positive drops).
- **NSG attachment** (`networkSecurityGroupIds`) -- the firewall appliance itself is placed in an NSG, controlling what traffic reaches the firewall VNIC. This provides a defense-in-depth layer at the VNIC level before the firewall policy evaluates traffic.
- **Rule evaluation order** -- (1) block malicious URLs first (cheapest rejection), (2) inspect web traffic with IDS (expensive but necessary for allowed traffic), (3) allow DNS (essential for resolution), (4) allow outbound, (5) deny all. This ordering minimizes IDS processing by pre-filtering known-bad traffic.
- **DNS explicitly allowed** (`allow-dns` on UDP 53) -- internal hosts need DNS resolution. This is called out as a separate rule rather than folded into outbound to make the intent explicit and auditable.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the firewall and policy will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<firewall-subnet-ocid>` | OCID of the dedicated subnet for the firewall appliance | OCI Console > Networking > Subnets, or `OciSubnet` status outputs (`subnetId`) |
| `<firewall-nsg-ocid>` | OCID of the NSG controlling traffic to the firewall VNIC | OCI Console > Networking > NSGs, or `OciSecurityGroup` status outputs (`networkSecurityGroupId`) |

## Related Presets

- **01-web-perimeter** -- use instead when L4 allow/deny is sufficient and IDS/URL filtering overhead is not justified
