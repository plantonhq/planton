# Internal Passthrough NLB VIP

An internal passthrough Network Load Balancer's VIP: a regional forwarding rule on the `INTERNAL` scheme that names its regional backend service directly — there is no proxy in a passthrough load balancer. Clients in the VPC reach the service on a private IP and, with `serviceLabel`, on a stable internal DNS name (the `service_name` output).

## When to Use

- Load balancing TCP or UDP traffic between services inside a VPC (databases, message brokers, internal APIs) without terminating the connection
- Fronting a service that must see the client's real source IP (passthrough preserves it)
- A next-hop load balancer for network appliances (pair with `isMirroringCollector` for Packet Mirroring collectors)

## Remix Notes

- The backend service must be a regional `GcpBackendService` with `loadBalancingScheme: INTERNAL`, a TCP or UDP `protocol`, and a health check; give it `failoverPolicy` and `connectionTrackingPolicy` for the passthrough-specific behaviors.
- `ports` lists up to five individual ports or ranges; use `allPorts: true` (with `ipProtocol: L3_DEFAULT` to forward every protocol) for services that listen everywhere.
- `allowGlobalAccess` lets clients in every region reach the VIP; leave it off to confine access to the rule's region.
- `network` and `subnetwork` are the VPC and subnet the VIP is drawn from — the subnet is required on a custom-mode network.
- Reference the backend service, network, and subnet via `valueFrom` instead of literal self-links.
