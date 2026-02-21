---
title: "Development Load Balancer"
description: "This preset creates a minimal-cost public OCI Application Load Balancer for dev/test environments. It uses the flexible shape locked to the minimum 10 Mbps bandwidth, a single subnet, an HTTP-only..."
type: "preset"
rank: "03"
presetSlug: "03-development"
componentSlug: "application-load-balancer"
componentTitle: "Application Load Balancer"
provider: "oci"
icon: "package"
order: 3
---

# Development Load Balancer

This preset creates a minimal-cost public OCI Application Load Balancer for dev/test environments. It uses the flexible shape locked to the minimum 10 Mbps bandwidth, a single subnet, an HTTP-only listener, and a TCP health check that requires no application-level health endpoint. Backends are omitted so they can be added dynamically as development instances come and go. No SSL, NSG, delete protection, rule sets, or hostnames are configured.

## When to Use

- Development or testing environments where cost is the primary concern
- Rapid prototyping where the load balancer configuration will change frequently
- Environments where backends are added and removed dynamically (ephemeral dev instances)
- Scenarios where a simple HTTP endpoint is sufficient and HTTPS is not required

## Key Configuration Choices

- **Minimum bandwidth** (`shapeDetails: 10/10 Mbps`) -- Locks both minimum and maximum to the lowest allowed value (10 Mbps), minimizing cost. Dev/test traffic volumes are typically negligible. Increase `maximumBandwidthInMbps` if load testing through this LB.
- **Single subnet** (`subnetIds`) -- One public subnet is sufficient for dev/test. Multi-AD HA is unnecessary for non-production workloads.
- **TCP health check** (`backendSets[0].healthChecker.protocol: tcp`) -- Verifies only that the backend port is reachable, without requiring the application to implement a `/health` endpoint. Faster to set up during development. Switch to HTTP health checks for staging or production.
- **No backends** -- The backend set is created empty. Backends are added as development instances are provisioned, avoiding the need to update the preset every time an instance is created or destroyed.
- **Round-robin policy** (`backendSets[0].policy: round_robin`) -- Simplest distribution strategy. Appropriate for dev/test where backend uniformity and precision are not concerns.
- **No SSL** -- Avoids the overhead of certificate management for development. Traffic is plaintext HTTP on port 80.
- **No NSG** -- Dev environments typically do not need fine-grained network access control. The subnet's default security list provides basic isolation.
- **No delete protection** -- Dev load balancers are frequently created and destroyed. Delete protection would add friction to teardown.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the load balancer will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<public-subnet-ocid>` | OCID of a public subnet for the load balancer | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |

## Related Presets

- **01-internet-facing-https** -- Use instead for production workloads that need HTTPS termination, HA, and HTTP-to-HTTPS redirect
- **02-internal-http** -- Use instead for private internal services within the VCN
