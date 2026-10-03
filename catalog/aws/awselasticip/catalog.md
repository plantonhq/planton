# AWS Elastic IP

Deploys a static public IPv4 address from Amazon's address pool or from a Bring-Your-Own-IP (BYOIP) pool, with optional network border group scoping for Local Zones and Wavelength zones. Unlike an instance's ephemeral public IP, the address persists until the Infra Component is destroyed, which is what makes it the anchor for stable endpoints: NLB subnet mappings, NAT gateway egress, and pet instances that must keep their address across stop/start cycles. For the common case, no spec fields beyond `region` are needed -- the optional fields cover attachment, BYOIP, IPAM pools, edge zones, and reverse DNS.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Elastic IP** -- a VPC-domain static public IPv4 address allocated from Amazon's pool (default) or from a BYOIP pool when `publicIpv4Pool` is specified
- **AWS Tags** -- resource metadata tags (organization, environment, resource kind, resource ID) applied to the Elastic IP

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- an active connection in the Connect module with credentials for the target AWS account. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.
- **Planton Runner** -- required when using Runner-based credential delivery. Not needed for inline credentials or cross-account trust authentication modes.

### AWS Account

- **Elastic IP quota** -- AWS accounts have a default limit of 5 Elastic IPs per region. Request a quota increase before allocating additional addresses. Each EIP incurs hourly charges when not associated with a running resource.
- **BYOIP pool** (optional) -- if using a Bring-Your-Own-IP pool, the pool must be provisioned and advertised in the target region before referencing it in `publicIpv4Pool`.
- **Network border group** (optional) -- required when allocating an EIP in an AWS Local Zone or Wavelength zone. Verify the target zone is enabled in your account.

## Deploy

### Console

Open the deployment store, find **AWS Elastic IP**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Standard Elastic IP** preset in the [Presets](#presets) tab to pre-populate a default configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsElasticIp
metadata:
  name: nlb-frontend
  org: acme-corp
  env: prod
spec:
  region: us-west-2
```

```shell
planton apply -f elastic-ip.yaml
```

This allocates a standard VPC Elastic IP from Amazon's address pool in us-west-2. No BYOIP or network border group is configured. The allocated IP persists until the Infra Component is destroyed -- it does not change across Infra Job runs. An Infra Job tracks the provisioning in real time.

### InfraChart

When the address deploys alongside the instance it fronts in one chart, wire the attachment via ValueFromRef:

```yaml
spec:
  region: us-west-2
  instance:
    valueFrom:
      kind: AwsEc2Instance
      name: bastion
      fieldPath: status.outputs.instance_id
```

The InfraPipeline resolves the dependency graph, deploys the instance first, then allocates the address and associates it.

## Key Configuration

These are the most important decisions when configuring an Elastic IP. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**BYOIP pool** -- Set `publicIpv4Pool` to a BYOIP pool ID when you need the Elastic IP to come from your own registered IP address range. Use `address` to request a specific IP from that pool. Both fields are ForceNew -- changing either requires replacing the EIP with a new allocation.

**Network border group** -- Set `networkBorderGroup` to scope the EIP to a specific AWS Local Zone or Wavelength zone instead of the default regional scope. Required when associating the EIP with resources in edge locations. This field is also ForceNew.

**IPAM pool** -- Set `ipamPoolId` to allocate the address from an Amazon VPC IPAM public pool, so the allocation is planned and audited against centrally managed address space. Combine with `address` to recover a specific address the pool holds. ForceNew.

**Instance / ENI attachment** -- Point the address at its target from the EIP side: `instance` (an AwsEc2Instance reference) or `networkInterface` (the precise per-ENI form), optionally pinning `associateWithPrivateIp` on multi-IP targets. At most one target may be set -- AWS associates an address with exactly one thing. These update in place: changing the target re-associates the same address.

**Reverse DNS** -- Set `reverseDnsDomainName` for workloads that send mail from the address. AWS validates SERVER-SIDE that a forward A record for the domain already resolves to the EIP before granting the PTR record: create the EIP, point DNS at its `public_ip`, then set this field on a follow-up apply.

**Mutability** -- The allocation-shaping fields (`publicIpv4Pool`, `address`, `networkBorderGroup`, `ipamPoolId`) trigger resource replacement when changed -- a NEW public address. The association fields and `reverseDnsDomainName` update in place.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsEc2Instance** | `instance` | `status.outputs.instance_id` |
| **AwsEc2Instance** | `networkInterface` (the precise per-ENI form; literal `eni-...` ids also work) | `status.outputs.primary_network_interface_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `allocation_id` | EIP allocation identifier (e.g., eipalloc-0123...) | Network Load Balancer subnet mappings, NAT Gateway attachment |
| `public_ip` | Static public IPv4 address | DNS A records, firewall allowlists, application configuration |
| `arn` | Amazon Resource Name of the EIP | IAM policies, resource-level permissions |
| `public_dns` | Public DNS hostname (e.g., ec2-1-2-3-4.compute-1.amazonaws.com) | DNS CNAME records, reverse-lookup verification |
| `association_id` | Association ID (eipassoc-...) when the spec attaches the address | Attachment evidence, troubleshooting |
| `ptr_record` | Granted reverse DNS record when `reverseDnsDomainName` is set | Mail-deliverability verification |

The `allocation_id` is the primary output consumed by downstream resources. Network Load Balancers reference it in `subnetMappings` to bind a static IP per subnet. NAT Gateways reference it to provide a stable outbound IP for private subnets.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Standard EIP** -- An Elastic IP from Amazon's pool with no special configuration. The most common pattern for NLB frontends, NAT Gateways, and EC2 instances needing stable public addresses. Start from the **Standard Elastic IP** preset.

**BYOIP pool allocation** -- An Elastic IP from a customer-owned IP range, useful for maintaining IP reputation (e.g., email servers) or meeting contractual requirements for specific IP addresses. Start from the **BYOIP Pool Elastic IP** preset.

**Instance-attached address** -- A stable public IP for a pet instance (bastion, license server), declared on the EIP side because the instance API cannot pull an address itself. Start from the **Instance-Attached Elastic IP** preset.

## Works With

- [**AWS NAT Gateway**](/infra-catalog/aws-nat-gateway) -- consumes this component's `allocation_id` as the gateway's stable outbound address
- [**AWS NLB**](/infra-catalog/aws-nlb) -- consumes `allocation_id` in subnet mappings to bind a static IP per subnet
- [**AWS EC2 Instance**](/infra-catalog/aws-ec2-instance) -- the attachment target this component references, by `instance_id` or `primary_network_interface_id`
