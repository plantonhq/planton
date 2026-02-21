---
title: "Standard x86"
description: "This preset creates an OCI Functions Application with x86 processor architecture, NSG-protected networking, and APM tracing enabled. Functions deployed to this application run on Intel/AMD x86-64..."
type: "preset"
rank: "01"
presetSlug: "01-standard-x86"
componentSlug: "functions-application"
componentTitle: "Functions Application"
provider: "oci"
icon: "package"
order: 1
---

# Standard x86

This preset creates an OCI Functions Application with x86 processor architecture, NSG-protected networking, and APM tracing enabled. Functions deployed to this application run on Intel/AMD x86-64 infrastructure, which offers the broadest compatibility with existing container images and third-party libraries. This is the standard starting point for serverless workloads on OCI.

## When to Use

- First Functions application in a project or environment
- Serverless event processing, API handlers, or scheduled tasks using standard x86 container images
- Applications where existing Docker images were built for x86-64 and recompiling for ARM is not practical
- Any Functions workload where APM observability is desired for debugging and performance monitoring

## Key Configuration Choices

- **x86 architecture** (`shape: generic_x86`) -- runs functions on Intel/AMD x86-64 infrastructure. This is the safest default because virtually all container images and native libraries are x86-compatible. For cost savings on CPU-bound workloads, consider switching to `generic_arm` (Ampere A1) -- ARM shapes offer better price-performance but require ARM-compatible images. The shape is immutable after creation.
- **Private subnet** (`subnetIds`) -- functions execute in a private subnet. They can reach other resources within the VCN (databases, caches, internal APIs) and access the internet via a NAT Gateway configured on the subnet. At least one subnet is required; multiple subnets across availability domains improve resilience.
- **NSG protection** (`networkSecurityGroupIds`) -- restricts network traffic to and from the functions application. Configure egress rules to allow only the destinations your functions need (databases, external APIs). Ingress rules are typically not needed since functions are invoked through the OCI Functions service, not via direct network connections.
- **APM tracing** (`traceConfig` with `isEnabled: true`) -- sends distributed trace data to an OCI Application Performance Monitoring domain. This enables end-to-end request tracing across functions, API gateways, and other instrumented services. Essential for debugging cold start latency and identifying performance bottlenecks. Omit `traceConfig` entirely if APM is not set up in your environment.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the application will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<private-subnet-ocid>` | OCID of the private subnet where functions will execute | OCI Console > Networking > Subnets, or `OciSubnet` status outputs |
| `<functions-nsg-ocid>` | OCID of the NSG controlling function network access | OCI Console > Networking > NSGs, or `OciSecurityGroup` status outputs |
| `<apm-domain-ocid>` | OCID of the APM domain for distributed tracing | OCI Console > Observability & Management > Application Performance Monitoring > Domains |

## Related Presets

- **02-secure-production** -- Use instead for regulated environments requiring container image signature verification
