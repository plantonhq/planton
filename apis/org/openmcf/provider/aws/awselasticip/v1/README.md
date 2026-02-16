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

- An AWS account and region configured in your OpenMCF stack input.
- (Optional) A registered BYOIP address range if you need IPs from your own pool.

## Spec Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `public_ipv4_pool` | string | No | BYOIP pool ID to allocate from. Omit to use Amazon's pool. ForceNew. |
| `address` | string | No | Specific IP from BYOIP pool. Requires `public_ipv4_pool`. ForceNew. |
| `network_border_group` | string | No | Location scope (Local Zone / Wavelength). Omit for Region default. ForceNew. |

**Note:** For the 95%+ use case (allocate a standard VPC EIP), no spec fields are needed. Simply provide an empty spec.

**ForceNew warning:** All optional fields trigger EIP replacement when changed. Treat allocated EIPs as immutable.

## Outputs

| Output | Description |
|--------|-------------|
| `allocation_id` | EIP allocation ID (`eipalloc-xxx`). Primary reference for NLB, NAT Gateway. |
| `public_ip` | The public IPv4 address. |
| `arn` | EIP ARN for IAM policies. |
| `public_dns` | Public DNS hostname (e.g., `ec2-1-2-3-4.compute-1.amazonaws.com`). |

## Minimal Example

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: my-eip
spec: {}
```

## Production Example (NLB Static IPs)

Allocate three Elastic IPs for a three-AZ NLB:

```yaml
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: nlb-eip-az1
spec: {}
---
apiVersion: aws.openmcf.org/v1
kind: AwsElasticIp
metadata:
  name: nlb-eip-az2
spec: {}
---
apiVersion: aws.openmcf.org/v1
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

## What Is Deliberately Omitted (v1)

- **EIP association** (`instance`, `network_interface`) — Association has an independent lifecycle and should be managed by the consumer resource (NLB, NAT Gateway) or via a separate AwsEipAssociation component.
- **IPAM pool support** (`ipam_pool_id`) — Enterprise IPAM, <5% adoption.
- **Customer-owned IP pool** — AWS Outpost only, <1% adoption.
- **PTR record management** — Separate AWS resource, niche use case.
- **`domain` field** — Hardcoded to `"vpc"` (EC2-Classic is deprecated).
