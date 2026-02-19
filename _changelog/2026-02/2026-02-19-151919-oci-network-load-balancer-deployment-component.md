# OCI Network Load Balancer Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Pulumi CLI Integration, Provider Framework

## Summary

Implemented the OciNetworkLoadBalancer deployment component (R12, enum 3321) -- an L4 TCP/UDP load balancer for Oracle Cloud Infrastructure. This is the second load balancer component in the OCI provider, complementing the existing L7 OciLoadBalancer (R11). The component bundles the NLB with backend sets, backends, and listeners into a single deployment unit with both Pulumi (Go) and Terraform (HCL) IaC modules.

## Problem Statement / Motivation

OpenMCF's OCI provider had an L7 application load balancer (OciLoadBalancer) but lacked L4 network load balancing capabilities. OCI Network Load Balancer is critical for:

### Pain Points

- No way to deploy L4 TCP/UDP load balancers through OpenMCF for OCI
- Workloads requiring source IP preservation (firewalls, security appliances, logging) had no NLB option
- OKE clusters using Network Load Balancer for Kubernetes service type LoadBalancer had no OpenMCF component
- Mixed TCP/UDP workloads (DNS servers, gaming backends) needed a single NLB with multi-protocol support

## Solution / What's New

Designed and implemented a complete OciNetworkLoadBalancer component following the established forge workflow, with careful attention to the significant differences between L4 and L7 load balancing.

### Key Differences from OciLoadBalancer (L7)

The NLB is architecturally distinct from the L7 LB:

- **Single subnet** (`subnet_id`) vs multiple subnets -- NLB deploys into one subnet
- **No shape/bandwidth** -- NLB is fully elastic, no bandwidth configuration needed
- **No SSL/certificates/hostnames/rule sets** -- L4 has no TLS termination or HTTP routing
- **Tuple-based policies** (FIVE_TUPLE, THREE_TUPLE, TWO_TUPLE) instead of round-robin/least-connections
- **Extended health checking** -- supports HTTP, HTTPS, TCP, UDP, and DNS protocols with a nested DNS health check configuration
- **Source IP preservation** -- native `is_preserve_source_destination` support
- **Backend target_id** -- can reference compute instances or private IPs by OCID
- **Protocol-specific idle timeouts** -- separate TCP, UDP, and L3IP timeouts per listener
- **PPv2 proxy protocol** support on listeners
- **Advanced failover** -- `is_fail_open`, `is_instant_failover_enabled`, `is_instant_failover_tcp_reset_enabled`

## Implementation Details

### Proto API (spec.proto)

- **14 top-level fields** including compartment_id (StringValueOrRef), subnet_id (singular StringValueOrRef), backend_sets, listeners
- **7 nested messages**: ReservedIp, BackendSet, HealthChecker, DnsHealthCheck, Backend, Listener, plus top-level spec
- **3 enums**: BackendSet.Policy (five_tuple/three_tuple/two_tuple), HealthChecker.Protocol (http/https/tcp/udp/dns), Listener.Protocol (tcp/udp/tcp_and_udp/any)
- Resolved a protobuf C++ scoping collision: `dns` enum value conflicted with `DnsHealthCheck dns` field; renamed field to `dns_health_check` (yields `dnsHealthCheck` in YAML)

### Validation Tests (spec_test.go)

- **49 Ginkgo/Gomega tests** (31 valid scenarios, 18 invalid scenarios)
- Covers all protocol variants, health checker types, DNS health check configuration, backend identification modes (ip_address vs target_id), failover features, PPv2, idle timeouts, and valueFrom references

### Pulumi Module (6 Go files)

| File | Purpose |
|------|---------|
| `main.go` | `Resources()` entry point, provider setup |
| `locals.go` | `Locals` struct, display name fallback, freeform tags |
| `outputs.go` | Output key constants |
| `network_load_balancer.go` | NLB resource creation, IP address extraction |
| `backend_set.go` | Backend sets, backends, health checker + DNS builder |
| `listener.go` | Listeners with DependsOn ordering |

### Terraform Module (5 HCL files)

- `main.tf` -- 4 resource types: NLB, backend_set (for_each), backend (for_each), listener (for_each)
- `variables.tf` -- Full type specification with optional defaults
- `locals.tf` -- Enum mapping (policy, protocol), tag computation, flat backend map
- `outputs.tf` -- network_load_balancer_id, ip_addresses
- `provider.tf` -- OCI provider >= 5.0

### Kind Registration

- `OciNetworkLoadBalancer = 3321` added to CloudResourceKind enum under "Advanced Networking" section
- `kind_map_gen.go` regenerated with new entry in `ProviderOciMap`

## Benefits

- **L4 load balancing** for TCP/UDP workloads in OpenMCF's OCI provider
- **Source IP preservation** for security appliances and logging
- **DNS health checking** -- unique to NLB, supports DNS-based backend health verification
- **Instant failover** -- NLB-exclusive feature for minimal disruption during backend failures
- **Infra-chart composability** -- all OCID fields use StringValueOrRef for wiring to OciCompartment, OciSubnet, OciNetworkSecurityGroup

## Impact

- **OCI Provider**: 12th resource kind implemented (12/37 total)
- **Phase 3 Progress**: 2 of 4 Advanced Networking components complete (OciLoadBalancer + OciNetworkLoadBalancer)
- **Users**: Can now deploy both L4 and L7 load balancers for OCI workloads through OpenMCF

## Validation Results

- `go build` -- clean
- `go vet` -- clean
- `go test` -- 49/49 passed
- `terraform validate` -- success

## Related Work

- **R11 OciLoadBalancer** -- L7 sibling component, used as the primary pattern reference
- **R13 OciDrg** -- next component in Phase 3 (Advanced Networking)

---

**Status**: Production Ready
