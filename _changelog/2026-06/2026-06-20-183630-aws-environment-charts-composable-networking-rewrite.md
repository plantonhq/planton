# AWS Environment Charts: Composable Networking Rewrite

**Date**: June 20, 2026
**Type**: Breaking Change
**Provider**: AWS
**Chart(s)**: eks-environment, ecs-environment, kafka-streaming, microservices-backend, ml-workbench

## Summary

Rebuilt the network foundation of all five AWS environment charts to compose the
new decomposed Planton networking primitives — a thin `AwsVpc` plus standalone
`AwsInternetGateway`, public/private `AwsSubnet` (with inline `routes`),
`AwsElasticIp`, and `AwsNatGateway` — wired together by `valueFrom`. The charts
previously emitted a bundled `AwsVpc` (with `vpcCidr` / `subnetSize` /
`subnetsPerAvailabilityZone` / `isNatGatewayEnabled`) and read list outputs
(`status.outputs.privateSubnets[].id`) that no longer exist, so they failed to
build against the current platform schema. The same pass also threads the
now-required `region` field through every `Aws*` resource and standardizes all
cross-resource `fieldPath` references on snake_case.

## Problem Statement / Motivation

Planton decomposed AWS networking: the VPC component became a thin, real-VPC-only
resource, and subnets, internet gateways, NAT gateways, and Elastic IPs became
first-class standalone components composed by reference. This is a deliberate
"LEGO block" design — lifecycles differ, and an explicit graph is easier for both
humans and coding agents to reason about and recombine.

### Pain Points

- **Charts no longer built.** Every networking chart emitted the removed bundled
  `AwsVpc` shape and consumed deleted list outputs (`privateSubnets[]` /
  `publicSubnets[]`). `planton chart build` rejected them.
- **`region` is now required on every `Aws*` resource and is not injected.** These
  five charts predated that requirement and set `region` nowhere, so even the
  non-VPC resources (EKS, ALB, RDS, MSK, SageMaker, IAM, KMS, …) were invalid.
- **Inconsistent output casing.** Some charts used camelCase `fieldPath`
  (`status.outputs.vpcId`), others snake_case. The verified source of truth (each
  component's `default_kind_field_path` and its presets) is snake_case.
- **Latent resource drift.** A few resources had fallen behind their schemas
  independent of networking (see Implementation Details).

## Solution / What's New

Each chart now stands up its network explicitly, following the platform's
existing decomposed-network grain (the OpenStack `project-landing-zone` chart):

- A thin `AwsVpc` (region + `cidrBlock` + DNS toggles).
- An `AwsInternetGateway` attached to the VPC.
- One **public** and one **private** `AwsSubnet` per Availability Zone. A subnet
  is "public" or "private" purely by where its `0.0.0.0/0` route points — public
  subnets route to the Internet Gateway (and set `mapPublicIpOnLaunch`), private
  subnets route to a NAT gateway.
- `AwsElasticIp` + `AwsNatGateway` for private-subnet egress, controlled by a new
  `nat_mode` parameter.

### `nat_mode` — selectable egress topology

A new `nat_mode` value on every chart:

- `single` (default): one Elastic IP + NAT gateway in the first AZ's public
  subnet; all private subnets route to it. Cost-conscious.
- `per_az`: one Elastic IP + NAT gateway per AZ; each private subnet routes to the
  NAT in its own AZ. No cross-AZ dependency; higher cost.
- `none`: private subnets get no default route (no outbound internet).

### Chart Structure

`templates/network.yaml` (or `templates/compute.yaml` for ml-workbench) now emits,
in dependency order: `AwsVpc` → `AwsInternetGateway` → public `AwsSubnet`(s) →
`AwsElasticIp` + `AwsNatGateway` (per `nat_mode`) → private `AwsSubnet`(s) →
`AwsSecurityGroup`. Downstream consumers reference standalone subnets by
`status.outputs.subnet_id`.

## Implementation Details

### Resources Included (per chart, network foundation)

| Resource | Purpose |
|----------|---------|
| `AwsVpc` | Thin VPC: region, `cidrBlock`, DNS support/hostnames |
| `AwsInternetGateway` | Public egress; route target for public subnets |
| `AwsSubnet` (public, per AZ) | `mapPublicIpOnLaunch` + `routes[0.0.0.0/0 → internet_gateway]` |
| `AwsElasticIp` | Static IP for the NAT gateway(s) |
| `AwsNatGateway` | Private egress; route target for private subnets |
| `AwsSubnet` (private, per AZ) | `routes[0.0.0.0/0 → nat_gateway]` (omitted when `nat_mode: none`) |
| `AwsSecurityGroup` | Chart-specific ingress/egress |

Per-chart consumers were repointed to the standalone subnets and given `region`:
EKS cluster + node group (eks), ALB + Fargate service (ecs), MSK (kafka, 3 AZs),
Aurora RDS + ElastiCache Redis (microservices), SageMaker domain (ml-workbench).

### Conditional Resources

- `nat_mode` (`single` | `per_az` | `none`) gates the Elastic IP / NAT gateway
  count and the private-subnet routes.
- Existing toggles preserved: `dnsEnabled` / `httpsEnabled` (ecs, microservices),
  `databaseEnabled` / `cacheEnabled` / `messagingEnabled` (microservices),
  `vpcEnabled` / `customImagesEnabled` (ml-workbench), `create_hosted_zone` /
  `enable_kms_encryption` (eks).

### Subnet routing example (verified against the component preset)

```yaml
routes:
  - destinationCidrBlock: 0.0.0.0/0
    targetType: internet_gateway
    targetId:
      valueFrom:
        kind: AwsInternetGateway
        name: "{{ values.env }}-igw"
        fieldPath: status.outputs.internet_gateway_id
```

`routes[].targetId` carries no default kind (the kind is implied by `targetType`),
so every route reference spells out `kind` + `fieldPath` explicitly.

### Resource drift fixed along the way

Surfaced by validating each rendered resource against the live schema:

- **`AwsKmsKey`** now requires `deletion_window_days` in 7–30 (added `30`).
- **`AwsRdsCluster`** validates `preferred_maintenance_window` /
  `preferred_backup_window` patterns even when empty (added explicit non-overlapping
  UTC windows).
- **`AwsRedisElasticache`** requires `description` (added).
- **`AwsSqsQueue`** dropped `queueType` (a STANDARD queue is the default; FIFO is
  `fifo_queue: true`) — removed the invalid field.

### Design decisions

- **Fixed AZ slots over loops.** Per-AZ subnets/CIDRs are explicit
  (`availability_zone_N`, `public_subnet_N_cidr`, `private_subnet_N_cidr`) rather
  than generated from a list with index lookups. This keeps the templates robust,
  readable, and structurally identical across all five charts so the pattern is
  learnable from any one of them.
- **Per-chart metadata idioms preserved.** eks uses plain metadata, ecs keeps its
  `planton.dev/provisioner: pulumi` label, and kafka/microservices/ml-workbench
  keep `group:` (and existing consumer `relationships`). New network resources rely
  on `valueFrom` for graph edges.

## Benefits

- The charts build and deploy against the current platform again.
- Network topology is explicit and tunable (CIDRs per subnet, `nat_mode` for cost
  vs. availability) instead of opaque inside a bundled VPC.
- Consistent snake_case references and a uniform network block across all five
  charts reduce cognitive load and copy-paste errors.

## Impact

Anyone deploying these five AWS environment charts. Existing `values.yaml`
overrides need updating (see Breaking Changes / Migration Guide). No state is
migrated — these are configuration templates.

## Breaking Changes

- **VPC shape changed.** `subnet_size` is removed. The bundled VPC no longer
  carves subnets; the chart now creates standalone subnets with explicit CIDRs.
- **New required-ish values.** `aws_region` (the availability zones must belong to
  it) and per-subnet CIDRs (`public_subnet_N_cidr`, `private_subnet_N_cidr`) now
  exist with sensible defaults. `nat_mode` replaces the old implicit single NAT.
- **Output references changed.** `status.outputs.privateSubnets[].id` /
  `publicSubnets[].id` no longer exist; consumers read standalone
  `AwsSubnet` `status.outputs.subnet_id`.

## Migration Guide

For existing values files:

1. Add `aws_region` (e.g. `us-east-1`) matching your `availability_zone_*`.
2. Remove `subnet_size`. Optionally set `public_subnet_N_cidr` /
   `private_subnet_N_cidr` (defaults: public `/24`s, private `/20`s within
   `10.0.0.0/16`).
3. Choose `nat_mode` (`single` is the default and matches the old single-NAT
   behavior; use `per_az` for HA).

## Usage Example

```yaml
# values.yaml (excerpt)
aws_region: us-east-1
availability_zone_1: us-east-1a
availability_zone_2: us-east-1b
vpc_cidr: 10.0.0.0/16
nat_mode: per_az   # one NAT gateway per AZ
```

```bash
planton chart build aws/eks-environment
```

## Testing Strategy

Every rewritten chart was validated two ways:

- **Offline, per-resource** with `planton validate-manifest` (pinned to the target
  schema), rendering each chart across `nat_mode` = single / per_az / none and the
  other conditional toggles. This caught the four drift items above.
- **Server-side, end to end** with `planton chart build aws/<chart>` against the
  live platform — render + DAG + schema validation. All five charts pass (eks
  additionally verified under `per_az` and `none`; ml-workbench under
  `vpcEnabled: true`). A deliberately invalid field was confirmed to be rejected,
  proving the server-side check is real.

## Related Work

- OpenStack `project-landing-zone` — the decomposed-network grain this rewrite
  follows.

---

**Status**: ✅ Production Ready
**Timeline**: Single session
