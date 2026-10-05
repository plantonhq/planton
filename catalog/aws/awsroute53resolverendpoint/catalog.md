# AWS Route 53 Resolver Endpoint

Deploys a Route 53 Resolver endpoint — the hybrid-DNS bridge between a VPC and the networks outside it — with its forwarding rules and their VPC associations managed in-line. An inbound endpoint exposes ENI IP addresses that on-prem resolvers forward AWS-bound queries to; an outbound endpoint sends VPC queries out to on-prem (or any) name servers, steered by FORWARD rules and their SYSTEM overrides. Direction and security groups are fixed for life at the provider, so the shape you pick here is the shape you keep.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Resolver Endpoint** — the endpoint itself, with 2–10 ENI placements across your subnets (optionally with pinned private IPs), its security groups, IP family (`endpointType`), DNS transport protocols (Do53 / DoH / DoH-FIPS), and the two CloudWatch metrics toggles
- **Resolver Rules** — one per `rules` entry, keyed by rule name. FORWARD and DELEGATE rules bind to this endpoint; SYSTEM rules carry no endpoint binding because they restore recursive resolution instead of forwarding
- **Resolver Rule Associations** — one per (rule, VPC) pair from each rule's `vpcIds`, putting the rule into effect for queries originating in those VPCs

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with Route 53 Resolver permissions plus EC2 permissions for the endpoint's ENIs. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- Two or more subnets — in different availability zones, per AWS's recommendation — with free IP capacity for the endpoint ENIs (`ipAddresses` requires at least 2 entries).
- A security group allowing DNS (TCP and UDP port 53) between the endpoint and the resolvers or targets it talks to (`securityGroupIds`).
- For outbound FORWARD rules: a live network path (VPN, Direct Connect, or peering) from the endpoint's subnets to the target name servers. AWS validates the rule's shape, not the path.

## Deploy

### Console

Open the deployment store, find **AWS Route 53 Resolver Endpoint**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region and direction, the ENI subnet placements, security groups, and the forwarding rules with their VPC associations. Start from the **Inbound from On-Prem** preset or the **Outbound Forward to On-Prem** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsRoute53ResolverEndpoint
metadata:
  name: onprem-outbound
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  direction: OUTBOUND
  ipAddresses:
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: private-a
          fieldPath: status.outputs.subnet_id
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: private-b
          fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: dns-to-onprem
        fieldPath: status.outputs.security_group_id
  rules:
    - name: corp-domain
      domainName: corp.example.com
      ruleType: FORWARD
      targetIps:
        - ip: 10.20.0.53
        - ip: 10.20.1.53
      vpcIds:
        - valueFrom:
            kind: AwsVpc
            name: app-vpc
            fieldPath: status.outputs.vpc_id
    - name: aws-hosted-subdomain
      domainName: aws.corp.example.com
      ruleType: SYSTEM
      vpcIds:
        - valueFrom:
            kind: AwsVpc
            name: app-vpc
            fieldPath: status.outputs.vpc_id
```

```shell
planton apply -f resolver-endpoint.yaml
```

This creates an outbound endpoint on two ENIs, forwards `corp.example.com` queries from the app VPC to two on-prem name servers, and carves `aws.corp.example.com` back out to AWS's recursive resolution. An Infra Job tracks the provisioning in real time.

### InfraChart

When the endpoint deploys alongside its network in one chart, wire the subnet, security group, and VPC references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  direction: OUTBOUND
  ipAddresses:
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: private-a
          fieldPath: status.outputs.subnet_id
    - subnetId:
        valueFrom:
          kind: AwsSubnet
          name: private-b
          fieldPath: status.outputs.subnet_id
  securityGroupIds:
    - valueFrom:
        kind: AwsSecurityGroup
        name: dns-to-onprem
        fieldPath: status.outputs.security_group_id
```

The InfraPipeline resolves the dependency graph, deploys the VPC, subnets, and security group first, then provisions the endpoint on them.

## Key Configuration

These are the most important decisions when configuring a resolver endpoint. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Direction and security groups are one-way doors** — `direction` and `securityGroupIds` force a replacement at the provider: changing either destroys and re-creates the endpoint, and inbound endpoints hand out new ENI IPs that every on-prem conditional forwarder then has to be repointed at. Build the security group's rules to evolve inside one group rather than planning to swap groups later.

**The endpoint is the family's one always-on cost** — endpoint ENIs bill hourly (two minimum) whether or not queries flow; rules and associations are negligible beside them. Share one outbound endpoint across many rules and many VPCs — never create an endpoint per rule or per VPC.

**FORWARD needs reachability the control plane never checks** — creating a FORWARD rule with unreachable targets succeeds, and queries then time out at runtime. "Rule is COMPLETE but resolution fails" is a path problem: check the security group's DNS egress and the VPN/Direct Connect/peering route to the targets before suspecting the rule.

**SYSTEM rules are subdomain surgery** — a SYSTEM rule only means something as an override of a broader FORWARD rule: forward `corp.example.com` on-prem but let `aws.corp.example.com` resolve recursively in AWS. The most specific matching domain wins. A SYSTEM rule without a covering FORWARD rule changes nothing, and validation rejects target IPs on it.

**Retarget rules, never detach them** — moving a rule to a different endpoint updates in place, but clearing its endpoint binding forces the provider to replace the rule, and its VPC associations re-create with it. `ipAddresses` edits, by contrast, churn in place: the provider adds before it removes, so the two-address floor never breaks.

**Pin inbound IPs before anyone depends on them** — each `ipAddresses` entry may pin a fixed private IP (`ip` / `ipv6`); unset lets AWS pick one. For inbound endpoints, on-prem forwarders are configured against those exact addresses — pinning them keeps a future subnet reshuffle from silently changing what on-prem points at.

**Metrics toggles never revert by omission** — `rniEnhancedMetricsEnabled` and `targetNameServerMetricsEnabled` are tri-state on purpose: once set, removing the field leaves the last value at AWS. Revert with an explicit `false`, not by deleting the line.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `ipAddresses[].subnetId` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** | `securityGroupIds` | `status.outputs.security_group_id` |
| **AwsVpc** | `rules[].vpcIds` | `status.outputs.vpc_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `ip_addresses` | The ENI IP addresses the endpoint answers or originates on | Inbound: the addresses on-prem conditional forwarders target. Outbound: the source addresses on-prem firewalls must allow |
| `endpoint_id` | The endpoint's id (`rslvr-in-...` / `rslvr-out-...`) | Addressing the endpoint in AWS tooling; the provider's import ID |
| `endpoint_arn` | The endpoint's ARN | IAM policies scoped to this endpoint |

`host_vpc_id`, `rule_ids`, and `rule_association_ids` are also exported, but they are operational echoes — the VPC AWS derives from the subnets and the AWS-generated rule and association IDs keyed by rule name — kept for audits and imports rather than composition.

## Common Patterns

**Inbound from on-prem** — two ENIs across AZs guarded by a DNS security group; after deploy, point on-prem conditional forwarders at the `ip_addresses` output so on-prem clients can resolve private hosted zone names. Start from the **Inbound from On-Prem** preset.

**Outbound forward with a recursive carve-out** — the classic hybrid split: one FORWARD rule sends `corp.example.com` to on-prem name servers while a SYSTEM rule keeps `aws.corp.example.com` on AWS's recursive resolution. Start from the **Outbound Forward to On-Prem** preset.

**Shared outbound hub** — one outbound endpoint in a network VPC carrying every forwarding rule, each rule associated to all the VPCs that should honor it. This concentrates the hourly ENI cost in one place instead of paying it per VPC; the trade is that the hub VPC's subnets and security group become shared DNS infrastructure with a wide blast radius.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) — the VPCs whose queries each rule steers, wired via `rules[].vpcIds`
- [**AWS Subnet**](/infra-catalog/aws-subnet) — where the endpoint ENIs live, wired via `ipAddresses[].subnetId`
- [**AWS Security Group**](/infra-catalog/aws-security-group) — controls DNS traffic to and from the endpoint ENIs, wired via `securityGroupIds`
- [**AWS Route 53 Resolver Query Logging**](/infra-catalog/aws-route53-resolver-query-log) — logs the queries flowing through the same VPCs for audit and troubleshooting
- [**AWS Route 53 Resolver DNS Firewall**](/infra-catalog/aws-route53-resolver-firewall) — filters the same VPCs' DNS traffic before it ever reaches forwarding
