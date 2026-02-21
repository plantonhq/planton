# Web Perimeter Firewall

This preset creates an OCI Network Firewall with a policy designed for a standard web-facing perimeter. Inbound HTTP and HTTPS traffic from the internet is allowed to reach internal web servers, all outbound traffic from internal networks is permitted, and a default-deny rule drops everything else. The policy uses address lists, service definitions, and a service list to keep rules readable and maintainable as the firewall evolves.

## When to Use

- Protecting a web application tier behind a perimeter firewall in a DMZ subnet
- Implementing a default-deny posture for a VCN that hosts public-facing services
- Replacing or augmenting network security groups with stateful L4 inspection at the subnet level
- Any architecture where traffic between VCN subnets must pass through a next-generation firewall

## Key Configuration Choices

- **Address lists** -- `any-ipv4` (0.0.0.0/0) represents all internet sources; `internal-networks` covers RFC 1918 ranges (10/8, 172.16/12, 192.168/16). Adjust `internal-networks` to match your actual VCN CIDR blocks for tighter scoping.
- **Service list** (`web-traffic`) -- groups HTTP (port 80) and HTTPS (port 443) into a single reusable reference. Security rules reference the service list rather than individual services, reducing rule count and simplifying audits.
- **Rule evaluation order** -- security rules are evaluated in list order (first match wins). The order is: (1) allow web inbound, (2) allow all outbound, (3) deny everything else. This implements a classic allowlist model.
- **Default deny** (`deny-all` with action `drop`) -- any traffic not explicitly allowed by preceding rules is silently dropped. This is the security baseline for perimeter firewalls. Use `reject` instead of `drop` if you need ICMP unreachable or TCP RST responses for faster client failure detection.
- **No IDS/IPS** -- this preset performs L4 allow/deny only. For deep packet inspection and intrusion detection, see the `02-ids-with-url-filtering` preset.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the firewall and policy will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<firewall-subnet-ocid>` | OCID of the dedicated subnet for the firewall appliance (separate from application subnets) | OCI Console > Networking > Subnets, or `OciSubnet` status outputs (`subnetId`) |

## Related Presets

- **02-ids-with-url-filtering** -- use instead when L7 URL filtering and intrusion detection/prevention are required
