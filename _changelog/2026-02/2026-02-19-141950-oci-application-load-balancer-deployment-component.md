# OCI Load Balancer Deployment Component

**Date**: February 19, 2026
**Type**: Feature
**Components**: API Definitions, Provider Framework, Resource Management

## Summary

Implemented the OciApplicationLoadBalancer deployment component (R11, enum 3320), the most complex OCI component to date. This Layer 7 load balancer bundles 7 sub-resource types (load balancer, backend sets, backends, listeners, certificates, hostnames, rule sets) into a single atomic deployment unit with both Pulumi and Terraform modules.

## Problem Statement / Motivation

OCI's Application Load Balancer is a critical networking component for any production deployment, providing HTTP/HTTPS traffic distribution, SSL termination, and advanced routing. Unlike simpler resources, an OCI load balancer requires multiple tightly coupled sub-resources to function -- a backend set without a load balancer is useless, and a listener without a backend set has nowhere to route traffic.

### Pain Points

- An OCI load balancer requires at minimum: the LB itself, one backend set with health checker, and one listener
- Backend sets, listeners, certificates, hostnames, and rule sets are all scoped to a specific load balancer
- Creating these as separate Planton components would require users to manage 5+ YAML manifests for a basic LB setup
- Rule sets are essential for production (HTTP-to-HTTPS redirect) but have a complex polymorphic schema with 11 action types

## Solution / What's New

A single `OciApplicationLoadBalancer` component that bundles the full load balancer stack into one deployment unit, following the same bundling philosophy as `OciVcn` (which bundles gateways).

### Key Features

- **Flexible shape support** with configurable bandwidth (10-8000 Mbps)
- **Backend sets** with three load balancing policies (round robin, least connections, IP hash)
- **Health checking** with HTTP and TCP protocols, configurable intervals, retries, and thresholds
- **SSL termination** on both listener (client-facing) and backend set (re-encryption) contexts
- **Session persistence** via oneof: LB-managed cookie or application cookie (mutually exclusive)
- **Virtual hostname routing** for multi-domain deployments
- **Rule sets** covering HTTP redirects, header manipulation, access control, and connection limits
- **4 listener protocols**: HTTP, HTTP/2, TCP, gRPC

### Excluded (by design)

- `path_route_set`: deprecated by Oracle in favor of routing policies
- `routing_policy`: not in plan scope; can reference externally via `routing_policy_name`

## Implementation Details

### Proto API

The spec defines 17 nested message types with 3 enums:

```
OciApplicationLoadBalancerSpec
  ShapeDetails, ReservedIp
  BackendSet (Policy enum, HealthChecker, Backend, SslConfiguration)
    oneof: LbCookieSessionPersistenceConfig | SessionPersistenceConfig
  Listener (Protocol enum, ConnectionConfiguration, SslConfiguration)
  Certificate, Hostname
  RuleSet > RuleSetItem (Action enum, RedirectUri, Condition, IpMaxConnection)
```

Design decisions:
- **oneof for session persistence**: enforces mutual exclusivity at the schema level
- **Shared SslConfiguration message**: used by both backend sets and listeners, with `has_session_resumption` only relevant for listener context
- **Nested Protocol enums**: HealthChecker.Protocol (http, tcp) and Listener.Protocol (http, http2, tcp, grpc) are scoped to their respective messages to avoid collision
- **Flat RuleSetItem**: action enum determines which fields are relevant, matching the OCI API model
- **idle_timeout_in_seconds as int64**: clean proto type, converted to string for the Terraform provider quirk

### Pulumi Module (8 Go files)

Orchestration order in `Resources()`:
1. Create load balancer
2. Create certificates (depends on LB)
3. Create backend sets + backends (depends on LB)
4. Create hostnames (depends on LB)
5. Create rule sets (depends on LB)
6. Create listeners (depends on all above via explicit DependsOn)

IP addresses are extracted from the load balancer's `IpAddressDetails` output via `ApplyT`.

### Terraform Module (9 HCL files)

- Backend sets use `for_each` keyed by name
- Backends use a flattened `for_each` keyed by `"backendset:ip:port"` for uniqueness
- Listeners use `depends_on` for all sub-resources to ensure ordering
- Rule set items use nested `dynamic` blocks for conditions, redirect URIs, and IP max connections

### Validation Tests

59 Ginkgo/Gomega test cases:
- 30 valid scenarios (minimal, HTTPS, multiple backend sets, session persistence, rule sets, etc.)
- 29 invalid scenarios (missing required fields, invalid enums, boundary violations)

### Files Created

```
apis/dev/planton/provider/oci/ociapplicationloadbalancer/v1/
  spec.proto                        # 17 messages, 3 enums, ~370 lines
  api.proto, stack_input.proto, stack_outputs.proto
  spec_test.go                      # 59 test cases
  iac/pulumi/main.go                # Entrypoint
  iac/pulumi/module/
    main.go, locals.go, outputs.go
    load_balancer.go, backend_set.go, listener.go
    certificate.go, hostname.go, rule_set.go
  iac/tf/
    provider.tf, variables.tf, locals.tf, outputs.tf
    main.tf, backend_set.tf, listener.tf
    certificate.tf, hostname.tf, rule_set.tf
```

Kind registered as `OciApplicationLoadBalancer = 3320` in `cloud_resource_kind.proto`.

## Benefits

- Users can deploy a complete production load balancer from a single YAML manifest
- HTTP-to-HTTPS redirect (the most common LB rule) is natively supported via rule sets
- Multi-domain routing via hostnames enables serving multiple applications from one LB
- Both Pulumi and Terraform modules have feature parity
- 59 validation tests catch configuration errors before deployment

## Impact

- First component in Phase 3 (Advanced Networking) of the OCI provider expansion
- Enables the OKE Environment infra chart (requires LB for ingress)
- 11 of 37 OCI resource kinds now implemented (30%)
- Users can now build complete OCI web application stacks with networking, compute, containers, and load balancing

## Related Work

- Follows bundling philosophy from DD02 (VCN bundles gateways)
- Builds on patterns established by R07 OciComputeInstance and R08 OciContainerEngineCluster
- Next: R12 OciNetworkLoadBalancer (Layer 4 load balancing)

---

**Status**: Production Ready
