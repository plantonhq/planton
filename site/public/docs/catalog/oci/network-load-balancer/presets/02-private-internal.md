---
title: "Private Internal NLB"
description: "This preset creates a private OCI Network Load Balancer for internal service-to-service communication within the VCN. The NLB receives only a private IP address and is not accessible from the public..."
type: "preset"
rank: "02"
presetSlug: "02-private-internal"
componentSlug: "network-load-balancer"
componentTitle: "Network Load Balancer"
provider: "oci"
icon: "package"
order: 2
---

# Private Internal NLB

This preset creates a private OCI Network Load Balancer for internal service-to-service communication within the VCN. The NLB receives only a private IP address and is not accessible from the public internet. It listens on a single TCP port and distributes traffic to backends using five-tuple hashing. Source IP preservation is not enabled because internal services typically do not need to identify the original caller by IP -- service identity is handled at the application layer via mTLS, tokens, or headers.

## When to Use

- Internal APIs or gRPC services that receive traffic from other services within the VCN
- Database connection proxies or middleware that front multiple database instances behind a single endpoint
- Backend services behind a public L7 load balancer (OciApplicationLoadBalancer) or public NLB that need internal load balancing for a second tier
- Any TCP service that must not be reachable from outside the VCN

## Key Configuration Choices

- **Private NLB** (`isPrivate: true`) -- No public IP is assigned. The NLB is reachable only from within the VCN or via peered VCNs and transit gateways. Use preset 01-public-tcp instead for internet-facing services.
- **Source IP preservation omitted** -- Unlike the public preset, source IP preservation is not enabled. Internal traffic flows between known services where the caller's IP is not a security signal. Omitting this avoids the `skipSourceDestinationCheck` side effect on the NLB's VNIC, keeping the network configuration simpler.
- **Five-tuple policy** (`policy: five_tuple`) -- Consistent with the public preset. Provides the most granular connection distribution for general TCP traffic.
- **TCP health check** (`healthChecker.protocol: tcp`) -- A simple TCP connection check on the backend port. This is the universal health check that works with any TCP service, including services that do not expose HTTP health endpoints. Switch to `http` if your backends have a dedicated health endpoint.
- **Single listener on port 8080** (`listeners[0].port: 8080`) -- A common port for internal services. Adjust to match your application's listening port. Unlike the public preset, there is no need for multiple listeners since internal services typically expose a single port.
- **NSG applied** (`networkSecurityGroupIds`) -- Even internal services benefit from network segmentation. Configure the referenced NSG to allow inbound TCP on port 8080 from the specific subnets or CIDRs that should reach this service.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the NLB will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<private-subnet-ocid>` | OCID of a private subnet for the NLB | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |
| `<internal-nlb-nsg-ocid>` | OCID of the NSG controlling access to the NLB (allow TCP 8080 from internal CIDRs) | OCI Console > Networking > NSGs, or `OciSecurityGroup` status outputs |
| `<backend-ip-1>` | Private IP address of the first backend server | OCI Console > Compute > Instances, or `OciComputeInstance` status outputs |
| `<backend-ip-2>` | Private IP address of the second backend server | OCI Console > Compute > Instances, or `OciComputeInstance` status outputs |

## Related Presets

- **01-public-tcp** -- Use instead for internet-facing NLBs with source IP preservation (web servers, OKE ingress controllers, public TCP APIs)
- **03-development** -- Use instead for dev/test environments where NSGs and backends are unnecessary
