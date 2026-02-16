# AwsElasticIp — Architecture & Research

## What Is an Elastic IP?

An AWS Elastic IP address is a static, public IPv4 address designed for dynamic cloud computing. Unlike standard public IPs that AWS assigns from its pool on instance launch (and reclaims on stop), an Elastic IP is allocated to your account and remains allocated until you explicitly release it.

Elastic IPs are fundamentally an **addressing primitive** — they don't do anything on their own. Their value comes from being referenced by other resources that need a stable public IP:

| Consumer | How It Uses the EIP | Reference Field |
|----------|-------------------|-----------------|
| Network Load Balancer | Binds one EIP per subnet for static ingress IPs | `allocation_id` in subnet mapping |
| NAT Gateway | Uses the EIP as the outbound IP for private subnets | `allocation_id` |
| EC2 Instance | Associates the EIP for a persistent public IP | `allocation_id` via EIP Association |
| VPN Gateway | Customer Gateway with known IP | `public_ip` for config |

## Addressing Model

### VPC Domain (Modern)

All Elastic IPs are VPC-scoped. EC2-Classic ("standard" domain) was fully retired in 2023. A VPC EIP:

- Has an `allocation_id` (format: `eipalloc-0123456789abcdef0`)
- Can be associated with instances, network interfaces, or NLB subnet mappings
- Lives in a specific Region (or Local Zone / Wavelength zone if `network_border_group` is set)
- Is billed at $0.005/hour when **not** associated with a running resource (idle EIP charges)

### BYOIP (Bring Your Own IP)

Organizations that own public IP address ranges can register them with AWS and allocate EIPs from their own pool:

1. Register the IP range with AWS via BYOIP provisioning (outside OpenMCF scope)
2. Create an IPv4 address pool from the registered range
3. Reference the pool ID in `public_ipv4_pool` when allocating EIPs
4. Optionally request a specific IP from the pool via `address`

This is useful for organizations that need their services to be reachable at known, pre-existing IP addresses — for example, when migrating on-premises services to AWS while keeping the same IPs.

### Network Border Groups

By default, an EIP is scoped to the entire AWS Region. You can narrow the scope using `network_border_group`:

- **Region** (default): `us-east-1` — EIP can be used with any resource in the Region
- **Local Zone**: `us-west-2-lax-1a` — EIP can only be used with resources in that Local Zone
- **Wavelength Zone**: `us-east-1-wl1-bos-wlz-1` — EIP for 5G edge applications

This is a ForceNew attribute — you cannot move an EIP between border groups after allocation.

## Immutability

All configurable fields on an EIP are ForceNew in the Terraform/Pulumi providers:

| Field | ForceNew? | Implication |
|-------|-----------|-------------|
| `domain` | Yes | Always "vpc", hardcoded |
| `address` | Yes | Cannot change the IP after allocation |
| `public_ipv4_pool` | Yes | Cannot move between pools |
| `network_border_group` | Yes | Cannot move between zones |

This means **an allocated EIP is effectively immutable**. Changing any attribute requires destroying the old EIP and creating a new one (new IP address). Plan accordingly — if a public IP is embedded in DNS records or firewall rules, replacing the EIP means updating all those references.

## Cost Model

| State | Cost |
|-------|------|
| Associated with running EC2 instance (one EIP per instance) | Free |
| Associated with NLB or NAT Gateway | Free |
| **Not associated** with any resource (idle) | $0.005/hour (~$3.60/month) |
| Additional EIP on an instance (beyond the first) | $0.005/hour |

AWS charges for idle EIPs to discourage address hoarding. Always release EIPs you're no longer using.

## Limits

| Limit | Default | Adjustable |
|-------|---------|------------|
| EIPs per Region | 5 | Yes (via quota increase) |
| EIPs per VPC | No limit | — |

The default limit of 5 EIPs per Region is commonly raised for production workloads. Request an increase via the AWS Service Quotas console.

## Security Considerations

- **No security groups**: EIPs themselves have no firewall rules. Security is controlled by the resource the EIP is associated with (the instance's security groups, the NLB's listeners, etc.).
- **Public exposure**: An EIP makes a resource publicly addressable. Ensure the associated resource has appropriate security groups and NACLs.
- **Idle EIP detection**: Unassociated EIPs are a common finding in AWS security audits. Monitor for idle EIPs and release them.

## Design Rationale

### Why no EIP association in this component?

Association (binding an EIP to an instance or ENI) has an **independent lifecycle** from the EIP allocation:

- An EIP might be moved between instances during maintenance
- An NLB doesn't use "association" — it takes `allocation_id` directly in its subnet mapping
- A NAT Gateway takes `allocation_id` directly
- Bundling association would couple two concerns with different change frequencies

Following the OpenMCF bundling principle: **split resources that have independent lifecycles**. The EIP is the allocation; association is the consumer's responsibility (or a separate AwsEipAssociation component).

### Why hardcode domain to "vpc"?

EC2-Classic was fully retired in August 2023. The `domain = "standard"` value in the Terraform provider will error on creation. There is no reason for users to see or set this field — it's always "vpc".

### Why include BYOIP fields in v1?

Even though <5% of users need BYOIP, the fields are simple strings with no structural complexity. Including them costs nothing in UX (they're optional, most users never see them) but prevents a v2 migration for BYOIP users.
