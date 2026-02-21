# Development NLB

This preset creates a minimal OCI Network Load Balancer for development, testing, or learning. It deploys a public NLB with a single TCP listener on port 80, a TCP health check, and no backends. Backends can be added after deployment or dynamically via compute instance target IDs. No NSG is attached, no source IP preservation is configured, and no advanced features are enabled -- the goal is a functional NLB with the least configuration friction possible.

## When to Use

- Development and testing environments where a quick TCP load balancer is needed without production concerns
- Learning OCI networking by experimenting with NLB behavior before adding complexity
- Proof-of-concept deployments that need a load balancer endpoint to validate connectivity
- Ephemeral environments spun up and torn down frequently where security posture is not a concern
- Debugging connectivity by isolating whether issues are NLB-related or application-related

## Key Configuration Choices

- **Public NLB** (`isPrivate` not set, defaults to `false`) -- Receives a public IP for easy access from anywhere. No need to configure VPN or bastion access during development.
- **No source IP preservation** -- Omitted to keep the configuration minimal. Add `isPreserveSourceDestination: true` if your testing requires the real client IP at backends.
- **No NSG** (`networkSecurityGroupIds` omitted) -- Removes the need to create and configure a Network Security Group. The NLB accepts traffic on port 80 from any source. For production, use preset 01-public-tcp or 02-private-internal which include NSG configuration.
- **TCP health check** (`healthChecker.protocol: tcp`) -- A simple TCP connection check on port 80. Does not require the backend to expose an HTTP health endpoint, making it compatible with any TCP service out of the box.
- **No backends** (`backends` omitted) -- The backend set is created empty. Add backends after deployment or reference compute instance target IDs. This avoids requiring backend IPs at deployment time.
- **Five-tuple policy** (`policy: five_tuple`) -- Consistent with the production presets so configurations can be promoted from dev to production by switching presets.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the NLB will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<public-subnet-ocid>` | OCID of a public subnet for the NLB | OCI Console > Networking > VCNs > Subnets, or `OciSubnet` status outputs |

## Related Presets

- **01-public-tcp** -- Use instead for production internet-facing NLBs with source IP preservation, NSG, and backend configuration
- **02-private-internal** -- Use instead for production internal services that should not be reachable from the public internet
