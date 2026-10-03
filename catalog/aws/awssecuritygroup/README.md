# AwsSecurityGroup

AWS EC2 Security Group resource for Planton. Provisions a stateful virtual firewall in a VPC with inline ingress and egress rules supporting IPv4/IPv6 CIDRs, managed prefix lists, references to other security groups, and self-referencing rules. Other resources compose onto the group by referencing its exported `security_group_id`.

## When to use

- Any workload that places network interfaces in a VPC needs at least one security group: EC2 instances, load balancers, EKS/ECS workloads, RDS/ElastiCache/MSK/OpenSearch data stores, MWAA environments, VPC endpoints.
- You want tiered network security (web tier / app tier / data tier) expressed as composable nodes, where the data tier's ingress references the app tier's group instead of hardcoding CIDRs.
- You want egress governance: with inline rules, an empty egress list denies all outbound traffic, so the manifest is the complete, auditable statement of what a group permits.

## Prerequisites

| Prerequisite | Why | Planton Resource |
|---|---|---|
| VPC | Every security group belongs to exactly one VPC | `AwsVpc` |
| Other security groups (optional) | Rules can reference sibling groups as traffic sources/destinations | `AwsSecurityGroup` |
| Managed prefix lists (optional) | Rules can target named CIDR sets (AWS services, office ranges) by stable ID | (external) |

## API envelope

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSecurityGroup
metadata:
  name: <resource-id>
spec: { ... }
```

## Spec fields reference

### Group

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `region` | string | **yes** | — | AWS region for the group. |
| `vpcId` | StringValueOrRef | **yes** | — | VPC the group belongs to. Supports `value` or `valueFrom` (AwsVpc). **ForceNew**. |
| `description` | string | **yes** | — | Purpose of the group (max 255 chars). **ForceNew** — AWS cannot edit a group description in place. |
| `revokeRulesOnDelete` | bool | no | false | Forcibly revoke this group's rules — and rules in other groups referencing it — before delete, so cross-referenced groups tear down without a `DependencyViolation`. |
| `additionalVpcIds` | list(StringValueOrRef) | no | [] | Share this group into OTHER VPCs (same account/region) so resources there attach the same group instead of per-VPC copies. Supports `valueFrom` (AwsVpc). Associations add/remove in place. AWS ignores cross-VPC group references in rules — prefer CIDR/prefix-list rules on shared groups. |

### Rules (`ingress[]` / `egress[]`)

Rules are managed INLINE on the group: the manifest owns the complete rule set, and rules added outside of it are removed on the next apply. An empty `ingress` denies all inbound; an empty `egress` denies all outbound (the allow-all rule AWS adds to new groups is revoked).

| Field | Type | Required | Default | Description |
|---|---|---|---|---|
| `protocol` | string | **yes** | — | `tcp`, `udp`, `icmp`, `icmpv6`, an IANA protocol number, or `-1` (all protocols). |
| `fromPort` | int32 | no | 0 | Range start (−1–65535). For ICMP: the ICMP *type* (−1 = all types). Must be 0 with protocol `-1` (CEL-enforced). |
| `toPort` | int32 | no | 0 | Range end (−1–65535). For ICMP: the ICMP *code* (−1 = all codes). Must be 0 with protocol `-1` (CEL-enforced). |
| `ipv4Cidrs` | list(string) | no | [] | IPv4 CIDR sources (ingress) / destinations (egress). |
| `ipv6Cidrs` | list(string) | no | [] | IPv6 CIDR sources / destinations. |
| `prefixListIds` | list(string) | no | [] | Managed prefix list IDs (`pl-...`). Target a named CIDR set — an AWS service like S3/DynamoDB gateway endpoints, or a customer-maintained office/partner range — by stable ID instead of hardcoding addresses. |
| `sourceSecurityGroupIds` | list(StringValueOrRef) | no | [] | Ingress ONLY: groups whose members may send traffic (CEL rejects it on egress rules). Supports `valueFrom` (AwsSecurityGroup). |
| `destinationSecurityGroupIds` | list(StringValueOrRef) | no | [] | Egress ONLY: groups whose members may receive traffic (CEL rejects it on ingress rules). Supports `valueFrom` (AwsSecurityGroup). |
| `selfReference` | bool | no | false | Allow traffic from/to this group itself — the standard intra-cluster pattern. |
| `description` | string | no | "" | Per-rule description (max 255 chars). |

A single rule may carry several sources at once (CIDRs + prefix lists + groups + self); AWS expands them into individual permissions server-side.

## Outputs

| Output | Description |
|---|---|
| `security_group_id` | Group ID (`sg-...`) — the join key other resources reference to attach the group. |
| `security_group_arn` | Group ARN — the form IAM policy conditions and resource-level permissions expect. |
| `owner_id` | Owning AWS account ID — needed for cross-account rule references (`<owner_id>/<group_id>`). |
| `additional_vpc_association_ids` | Association ids of the group's shares into other VPCs, keyed by VPC id (import form `<group_id>,<vpc_id>`); empty for single-VPC groups. |

## Deliberately omitted (with reasons)

| Provider surface | Why omitted |
|---|---|
| Standalone rule resources (`aws_vpc_security_group_ingress_rule` / `_egress_rule`) | Rules have no independent lifecycle here and are never referenced individually — they are folded into the group. AWS forbids mixing inline and standalone rules on one group, so the module never emits standalone rule resources. Per-rule TAGS are the standalone resources' one extra — operational metadata on individual permissions, never modeled. |
| Exclusive rule-set lockdown (`aws_vpc_security_group_rules_exclusive`) | Asserts a full rule-id set and removes everything else — the exclusive-management class; inline rules already own the rule set declaratively. |
| `name` / `name_prefix` overrides | The group name is `metadata.name` on both engines — one identity basis for the whole catalog. |
| Cross-account `<owner_id>/<group_id>` source modeling | A literal string in `sourceSecurityGroupIds.value` already carries the form; a structured field is not warranted until a real cross-account chart composes it. |

## Presets

- **web-tier** — HTTP/HTTPS from the internet, all outbound.
- **database-tier** — PostgreSQL from an app-tier group only.
- **bastion** — SSH from a corporate CIDR.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsSecurityGroup
metadata:
  name: app-tier
spec:
  region: us-west-2
  vpcId:
    valueFrom:
      kind: AwsVpc
      name: platform-vpc
      fieldPath: status.outputs.vpc_id
  description: App tier - service mesh + S3 egress via gateway endpoint
  revokeRulesOnDelete: true
  ingress:
    - protocol: tcp
      fromPort: 8080
      toPort: 8080
      selfReference: true
      description: Service-to-service traffic within the tier
  egress:
    - protocol: tcp
      fromPort: 443
      toPort: 443
      prefixListIds:
        - pl-63a5400a
      description: HTTPS to S3 via the gateway endpoint prefix list
```

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
