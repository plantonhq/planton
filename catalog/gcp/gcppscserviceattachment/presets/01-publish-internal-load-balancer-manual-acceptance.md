# Publish an Internal Load Balancer (Manual Acceptance)

## Use Case

Publish a service that lives behind an internal passthrough Network Load Balancer -- a database, a message broker, an internal API -- to named consumer projects, so they reach it through a Private Service Connect endpoint in their own VPC without peering and without caring about overlapping ranges. The production posture: named consumers with connection limits, a reject list, and reconciliation so a list edit takes effect on connected consumers.

## When to Use

- A shared platform service (database, cache, broker) consumed by several application projects
- Publishing a service to a partner organization's VPC without exposing your address space
- Any case where `ACCEPT_AUTOMATIC` would admit projects you do not know

## What This Creates

- A service attachment in `us-central1` in front of the `orders-db-ilb` forwarding rule
- One PSC NAT subnet (`orders-psc-nat`) consumer traffic is translated into
- `ACCEPT_MANUAL` with two consumer projects (10 and 4 endpoints) and one rejected project
- `reconcileConnections: true`, PROXY protocol off, `deletionPolicy: PREVENT`

## Customize

| Field | Default | Why Change |
|-------|---------|------------|
| `targetService` | `orders-db-ilb` | Your internal load balancer's regional forwarding rule (scheme `INTERNAL` or `INTERNAL_MANAGED`). |
| `natSubnets` | one subnet | Add PSC NAT subnets to serve more consumer endpoints; each connected endpoint consumes addresses. |
| `consumerAcceptLists` | two projects | Your consumer projects (`GcpProject` references or IDs), networks (`GcpVpcNetwork`), or endpoint URLs, each with a `connectionLimit`. |
| `enableProxyProtocol` | `false` | `true` when the backends speak the PROXY protocol and need the consumer's original address. |
| `domainNames` | — | A trailing-dot domain (`orders.internal.example.`) to register connected endpoints in Cloud DNS. |
| `deletionPolicy` | `PREVENT` | `DELETE` for ephemeral environments. |

Each consumer creates a regional `GcpGlobalForwardingRule` with an empty `loadBalancingScheme` whose `target` is this attachment's `self_link` output.
