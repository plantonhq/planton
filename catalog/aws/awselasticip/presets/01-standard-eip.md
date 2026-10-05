# Standard Elastic IP

**Use case:** Allocate a static public IPv4 address from Amazon's default pool.

This is the most common pattern — a zero-configuration EIP that provides a stable public IP for use with NLBs, NAT Gateways, or EC2 instances.

## What You Get

- A VPC Elastic IP allocated from Amazon's public IP pool
- Outputs: `allocation_id`, `public_ip`, `arn`, `public_dns`

## When to Use

- You need a static IP for an NLB subnet mapping
- You need a static outbound IP via NAT Gateway
- You need a persistent public IP for an EC2 instance
- You need a whitelistable IP for external service integrations

## Cost

- **Every public IPv4 address is billed hourly — associated or idle.** Since February 2024, AWS charges the same hourly rate for all public IPv4 addresses regardless of attachment; associating the address no longer makes it free.
- The only exemption is an address from a BYOIP pool you brought yourself (see the byoip-pool preset) — owning the range means AWS doesn't charge for using it.

The verified figure for this preset lives in the kind's generated estimate at `catalog/_pricing/estimates/awselasticip.yaml` — computed from the pinned price book, never hand-typed here.
