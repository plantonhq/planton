# AWS Network ACL

Deploys a network ACL — the stateless subnet-level firewall — with its inbound rules, outbound rules, and subnet associations managed in-line. Rules are evaluated in rule-number order per direction, first match wins, and unlike security groups a rule can DENY: block a bad CIDR, quarantine a subnet, enforce tier boundaries. Replies are not tracked, so every connection needs explicit rules in both directions, and anything matching no rule hits AWS's invisible catch-all deny.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Network ACL** — in the referenced VPC. The VPC binding is the only replace-forcing field; rules and associations all update in place
- **Inbound and Outbound Rules** — the in-line `ingress` and `egress` entries: rule number, allow/deny, protocol by name or number, an IPv4 or IPv6 CIDR, port range, and ICMP type/code. AWS's own catch-all deny rules (32767 for IPv4, 32768 for IPv6) always exist below them and are not manageable
- **Subnet Associations** — each subnet in `subnetIds` is atomically re-parented onto this ACL (a subnet has exactly one ACL — AWS replaces, never attaches). On destroy, associations tear down first and the subnets return to the VPC's default NACL

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** — an active connection in the Connect module with EC2 VPC permissions. Map it as the default for your environment, or specify it explicitly when creating the Infra Component.

### AWS Account

- The VPC the ACL belongs to (`vpcId`) and the subnets to filter (`subnetIds`) — nothing else.

## Deploy

### Console

Open the deployment store, find **AWS Network ACL**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and the spec: region, the VPC, the ordered rules in each direction, and the subnets to associate. Start from the **Web Tier** preset or the **Data Tier with a Quarantine Deny** preset in the [Presets](#presets) tab.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsNetworkAcl
metadata:
  name: web-tier-acl
  org: acme-corp
  env: prod
spec:
  region: us-east-1
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: app-vpc
      fieldPath: status.outputs.vpc_id
  ingress:
    - ruleNo: 100
      action: allow
      protocol: tcp
      cidrBlock: 0.0.0.0/0
      fromPort: 443
      toPort: 443
    - ruleNo: 110
      action: allow
      protocol: tcp
      cidrBlock: 0.0.0.0/0
      fromPort: 1024
      toPort: 65535
  egress:
    - ruleNo: 100
      action: allow
      protocol: tcp
      cidrBlock: 0.0.0.0/0
      fromPort: 443
      toPort: 443
    - ruleNo: 110
      action: allow
      protocol: tcp
      cidrBlock: 0.0.0.0/0
      fromPort: 1024
      toPort: 65535
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: public-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: public-b
        fieldPath: status.outputs.subnet_id
```

```shell
planton apply -f network-acl.yaml
```

This creates a web-tier ACL on both public subnets: HTTPS and ephemeral-port replies allowed in both directions, everything else falling through to AWS's catch-all deny. An Infra Job tracks the provisioning in real time.

### InfraChart

When the ACL deploys alongside its VPC and subnets in one chart, wire the references via ValueFromRef:

```yaml
spec:
  region: us-east-1
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: app-vpc
      fieldPath: status.outputs.vpc_id
  ingress:
    - ruleNo: 100
      action: allow
      protocol: tcp
      cidrBlock: 0.0.0.0/0
      fromPort: 443
      toPort: 443
  subnetIds:
    - valueFrom:
        kind: AwsSubnet
        name: public-a
        fieldPath: status.outputs.subnet_id
```

The InfraPipeline resolves the dependency graph, deploys the VPC and subnets first, then places the ACL over them.

## Key Configuration

These are the most important decisions when configuring a network ACL. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Stateless bites twice** — every connection needs rules in BOTH directions. Inbound 443 without an outbound ephemeral-port allow (1024–65535) drops every response, and the failure reads like a hung service, not a firewall — the classic NACL outage. When a NACL'd subnet "can't reach anything", check the reply direction first.

**Deny order is everything** — first match wins by `ruleNo`, so a deny numbered higher than the allow it should beat never fires. Put denies low (90 before the 100-series allows), leave gaps (100, 200, ...) so rules can be inserted without renumbering, and remember AWS's catch-all deny backstops everything unmatched — an empty ACL blocks all traffic.

**Association is replacement, and removal opens up** — listing a subnet here atomically replaces its previous ACL association. Removing it from `subnetIds` hands the subnet back to the VPC's DEFAULT NACL, which AWS ships as allow-all — so removing a subnet from a restrictive ACL can silently open it up, not lock it down.

**The VPC is a one-way door** — `vpcId` is fixed for life; changing it replaces the ACL. Everything else — rules in both directions and the subnet list — updates in place, so iterating on the policy is cheap once the ACL exists.

**NACLs complement security groups, never replace them** — the ACL is the subnet's coarse, stateless, deny-capable screen; the security group is the instance's fine, stateful allow-list. Tier boundaries and CIDR blocks belong here; service-to-service allows belong in security groups.

**Protocols are numbers underneath** — AWS stores protocol numbers and the provider normalizes names ("tcp" → 6) when diffing, so either spelling is stable. On `-1`/`all` rules AWS stores no ports — leave `fromPort` and `toPort` unset, which validation enforces so plans stay clean.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsVpc** | `vpcId` | `status.outputs.vpc_id` |
| **AwsSubnet** | `subnetIds` | `status.outputs.subnet_id` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `network_acl_id` | The ACL's id (`acl-...`) | Addressing the ACL in AWS tooling; the provider's import ID |
| `network_acl_arn` | The ACL's ARN | IAM policies scoped to this ACL |

`owner_id` is also exported, but it is an audit echo — the owning AWS account — not a composition input.

## Common Patterns

**Web tier** — the public-subnet posture done statelessly correct: HTTPS and HTTP in, ephemeral ports in (responses to outbound calls), HTTPS and ephemeral ports out (responses to inbound clients). Start from the **Web Tier** preset.

**Data tier with a quarantine deny** — what security groups cannot do: an explicit DENY for a known-bad range, numbered below the allows so it always wins, in front of a database-port-only allow from the VPC. Egress permits only ephemeral-port replies back into the VPC — the data tier initiates nothing. Start from the **Data Tier with a Quarantine Deny** preset.

**Emergency CIDR block** — when a range is attacking, add a low-numbered deny for it on the affected subnets' ACL. It takes effect immediately, beats every allow above it, and — being an in-place rule edit — is just as fast to remove when the incident closes.

## Works With

- [**AWS VPC**](/infra-catalog/aws-vpc) — the VPC the ACL belongs to, wired via `vpcId`
- [**AWS Subnet**](/infra-catalog/aws-subnet) — the subnets the ACL filters, wired via `subnetIds`
- [**AWS Security Group**](/infra-catalog/aws-security-group) — the stateful, instance-level allow-list this subnet-level screen complements
