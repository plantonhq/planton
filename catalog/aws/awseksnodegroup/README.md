# Overview

The AwsEksNodeGroup API resource provisions a MANAGED EKS node group: an
EC2 worker fleet that AWS provisions, health-checks, and rolls for you,
registered as workers of an `AwsEksCluster`.

## Why We Created This API Resource

Node pools are where cluster capacity gets its shape -- general pools,
Spot pools, GPU pools, dedicated tiers. Modeling the pool as a composable
node lets you:

- **Attach everything by reference**: the cluster
  (`status.outputs.name`), the node role (an `AwsIamRole` carrying its
  own worker policies), the subnets, and optionally an
  `AwsLaunchTemplate` -- the architecture graph shows exactly what each
  pool runs on.
- **Choose launch mechanics honestly**: simple pools use the inline knobs
  (instance types, disk size); hardened pools reference a launch template
  (custom AMI, IMDSv2 enforcement, encrypted volumes) -- and AWS's
  mutual exclusions between the two styles are enforced in the spec, not
  discovered at deploy time.
- **Roll fleets safely**: surge-enabled update config
  (launch-before-terminate), version/release pins for controlled AMI
  rollouts, and managed node auto-repair.

## Key Features

### Fleet Shape

- **Instance surface**: multiple instance types (the Spot best
  practice), the full EKS AMI-family set (AL2023, Bottlerocket incl.
  FIPS/NVIDIA, Windows, legacy AL2, CUSTOM), On-Demand / Spot /
  Capacity-Block purchase models.
- **Scale-to-zero**: min and desired of 0 express a dormant pool.
- **Scheduling controls**: node labels and up to 50 taints (the
  dedicated-capacity mechanism), both updating in place.

### Launch-Template Composition

- Reference an `AwsLaunchTemplate` for custom AMI + bootstrap, IMDSv2
  posture, encrypted/provisioned-IOPS storage, and instance tags.
- `$Default` version tracking turns a template-version promotion into a
  fleet rollout; numeric pins freeze the fleet until the pin moves.

### Rollouts and Repair

- **Update config**: bounded unavailability (count or percentage) and
  the `MINIMAL` surge strategy that never dips below capacity.
- **Node auto-repair**: EKS replaces or reboots unhealthy nodes within
  declared parallelism/threshold bounds, with per-condition overrides.
- **Version discipline**: pin `version`/`release_version`, roll on your
  schedule, `force_update_version` for unsatisfiable disruption budgets.

### Scale-Out Latency

- **Warm pools**: pre-initialized nodes (stopped, running, or
  hibernated) that join in seconds instead of cold-boot minutes --
  with a standby floor, an optional capacity ceiling, and
  reuse-on-scale-in so a warm boot is never paid for twice. Updates in
  place.

## Benefits

- **Composability**: cluster, role, subnets, template, and SSH source
  security groups all attach through `valueFrom` references.
- **Honest constraints**: the launch-template exclusions and the
  update-config exactly-one rule are CEL-enforced at validation time.
- **Consistency**: identical behavior across Terraform and Pulumi.

## Outputs

- `nodegroup_name`: the pool's name
- `nodegroup_arn`: the pool's ARN (access entries, IAM policies)
- `asg_name`: the EC2 Auto Scaling group AWS manages behind the pool
- `remote_access_sg_id`: the SSH security group AWS creates when remote
  access is enabled without explicit source groups (empty otherwise)

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
