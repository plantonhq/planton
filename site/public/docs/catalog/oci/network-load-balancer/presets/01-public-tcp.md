---
title: "Public TCP Pass-Through NLB"
description: "This preset creates a public OCI Network Load Balancer that distributes TCP traffic at Layer 4 with source IP preservation enabled. Backends receive the original client IP address in every packet,..."
type: "preset"
rank: "01"
presetSlug: "01-public-tcp"
componentSlug: "network-load-balancer"
componentTitle: "Network Load Balancer"
provider: "oci"
icon: "package"
order: 1
---

# Public TCP Pass-Through NLB

This preset creates a public OCI Network Load Balancer that distributes TCP traffic at Layer 4 with source IP preservation enabled. Backends receive the original client IP address in every packet, which is critical for logging, firewalls, and any application that needs to identify the true client. The NLB is fully elastic with no bandwidth shape configuration required -- it scales automatically with traffic. Two listeners (443 and 80) pass TCP connections through to backends, which are responsible for their own TLS termination if needed.

## When to Use

- Production TCP services that need a public entry point with source IP visibility (web servers, API gateways, OKE ingress controllers)
- Any workload where backends perform TLS termination themselves and need the original client IP for access logs or security policies
- High-throughput, low-latency scenarios where Layer 7 processing overhead is unnecessary
- Services that use TCP-based protocols beyond HTTP (databases with public endpoints, custom TCP protocols)

## Key Configuration Choices

- **Public NLB** (`isPrivate: false`) -- Receives a public IP address accessible from the internet. Use preset 02-private-internal instead for VCN-internal services.
- **Source IP preservation** (`isPreserveSourceDestination: true`) -- The flagship feature of the NLB. Backends see the real client IP in the packet headers, not the NLB's IP. This also enables `skipSourceDestinationCheck` on the NLB's VNIC automatically. Essential for security appliances, access logging, and IP-based rate limiting.
- **Five-tuple policy** (`policy: five_tuple`) -- Hashes on source IP, source port, destination IP, destination port, and protocol. This provides the most granular connection distribution and is the standard choice for general TCP traffic. Use `three_tuple` or `two_tuple` only if you need session affinity at the IP level.
- **HTTP health check on /health** (`healthChecker.protocol: http`) -- Application-level health verification is more reliable than TCP connect checks for production. The NLB sends an HTTP GET to `/health` on port 80 every 10 seconds and expects a 200 response. Adjust `urlPath` and `port` to match your application's health endpoint.
- **Backend source preservation** (`isPreserveSource: true`) -- Preserves the source IP at the backend set level in addition to the NLB level. This ensures consistent behavior when the NLB has multiple backend sets with different requirements.
- **Two TCP listeners (443 + 80)** -- Port 443 handles TLS traffic (terminated by backends, not the NLB) and port 80 handles plaintext TCP. If your backends redirect HTTP to HTTPS at the application layer, both listeners route to the same backend set. Remove the port 80 listener if you want to block plaintext traffic entirely.
- **NSG applied** (`networkSecurityGroupIds`) -- Controls which source IPs and ports can reach the NLB at the network level. Configure the referenced NSG to allow inbound TCP on ports 80 and 443 from the desired source CIDRs.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the NLB will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<public-subnet-ocid>` | OCID of a public subnet for the NLB | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<nlb-nsg-ocid>` | OCID of the NSG controlling access to the NLB (allow TCP 80, 443 inbound) | OCI Console > Networking > NSGs, or `OciSecurityGroup` status outputs |
| `<backend-ip-1>` | Private IP address of the first backend server | OCI Console > Compute > Instances, or `OciComputeInstance` status outputs |
| `<backend-ip-2>` | Private IP address of the second backend server | OCI Console > Compute > Instances, or `OciComputeInstance` status outputs |

## Related Presets

- **02-private-internal** -- Use instead for NLBs that serve traffic only within the VCN (internal APIs, gRPC services, database proxies)
- **03-development** -- Use instead for dev/test environments where source IP preservation, NSGs, and backends are unnecessary
