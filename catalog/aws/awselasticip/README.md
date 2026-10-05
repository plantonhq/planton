# AwsElasticIp

An **Elastic IP (EIP)** is a static, public IPv4 address that you allocate from Amazon's pool (or from your own Bring-Your-Own-IP range). Unlike the ephemeral public IPs that AWS assigns to EC2 instances, an Elastic IP persists until you explicitly release it — surviving instance stops, restarts, and re-associations.

## When to Use

- **Network Load Balancer with static IPs** — NLBs support binding one Elastic IP per subnet for a stable, whitelistable public endpoint.
- **NAT Gateway** — A NAT Gateway requires an Elastic IP to give private subnets a predictable outbound IP address.
- **EC2 instance with a fixed IP** — Assign a persistent public IP that survives stop/start cycles.
- **DNS or firewall allowlisting** — When external partners or services need to whitelist a static IP.

## When NOT to Use

- For load-balanced services where the IP can change — use an ALB or NLB without static IPs.
- When you need IPv6 — Elastic IPs are IPv4 only.
- For internal-only services — no public IP is needed.

## Prerequisites

- An AWS account and region configured in your Planton IaC input.
- (Optional) A registered BYOIP address range if you need IPs from your own pool.

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `public_ipv4_pool` | string | No | BYOIP pool ID to allocate from. Omit to use Amazon's pool. ForceNew. |
| `address` | string | No | Specific IP from a BYOIP or IPAM pool. Requires one of the pool fields. ForceNew. |
| `network_border_group` | string | No | Location scope (Local Zone / Wavelength). Omit for Region default. ForceNew. |
| `ipam_pool_id` | string | No | Amazon VPC IPAM public pool to allocate from — centrally planned/audited address space. ForceNew. |
| `instance` | ref/string | No | EC2 instance to attach the address to (references an AwsEc2Instance's `instance_id`). At most one of `instance` / `network_interface`. Updates in place. |
| `network_interface` | ref/string | No | ENI to attach the address to — the precise form (references an AwsEc2Instance's `primary_network_interface_id`, or a literal `eni-...`). Updates in place. |
| `associate_with_private_ip` | string | No | The specific private IP on the target the address maps to (multi-IP targets). Requires a target. Updates in place. |
| `reverse_dns_domain_name` | string | No | Reverse DNS (PTR) domain for the address — mail providers require it. AWS grants it only after a forward A record for the domain already resolves to the EIP. Updates in place. |

**Note:** For the most common case (allocate a standard VPC EIP), no spec fields are needed. Simply provide an empty spec.

**Mutability:** the allocation-shaping fields (pools, address, border group) are ForceNew — changing them replaces the EIP with a NEW public address. The association fields and reverse DNS update in place without re-allocating.

## Outputs

| Output | Description |
|--------|-------------|
| `allocation_id` | EIP allocation ID (`eipalloc-xxx`). Primary reference for NLB, NAT Gateway. |
| `public_ip` | The public IPv4 address. |
| `arn` | EIP ARN for IAM policies. |
| `public_dns` | Public DNS hostname (e.g., `ec2-1-2-3-4.compute-1.amazonaws.com`). |
| `association_id` | Association ID (`eipassoc-xxx`) when the spec attaches the address; empty when unattached. |
| `ptr_record` | The granted reverse DNS record when `reverse_dns_domain_name` is set; empty otherwise. |

## Minimal Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticIp
metadata:
  name: my-eip
spec: {}
```

## Production Example (NLB Static IPs)

Allocate three Elastic IPs for a three-AZ NLB:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticIp
metadata:
  name: nlb-eip-az1
spec: {}
---
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticIp
metadata:
  name: nlb-eip-az2
spec: {}
---
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticIp
metadata:
  name: nlb-eip-az3
spec: {}
```

Then reference from an NLB:

```yaml
spec:
  subnetMappings:
    - subnetId:
        valueFrom:
          kind: AwsVpc
          name: prod-vpc
          fieldPath: status.outputs.public_subnets.[0].id
      allocationId:
        valueFrom:
          kind: AwsElasticIp
          name: nlb-eip-az1
          fieldPath: status.outputs.allocation_id
```

## Attaching the Address

Two composition directions exist, and both are first-class:

- **Consumers that take an allocation** (NLB subnet mappings, NAT gateways)
  reference this component's `allocation_id` output — the EIP spec stays
  empty.
- **EC2 instances and ENIs** cannot pull an address themselves (the AWS
  instance API has no EIP argument), so the attachment is declared HERE:
  `spec.instance` or `spec.network_interface` points the address at its
  target, mirroring the AWS provider's own inline association surface.

## What Is Deliberately Omitted

- **Customer-owned IP pool** (`customer_owned_ipv4_pool`) — AWS Outposts
  surface; the catalog's recorded Outposts exclusion class.
- **`domain` field** — Module-wired to `"vpc"`: AWS retired EC2-Classic
  addresses, and the provider itself refuses to read them.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
