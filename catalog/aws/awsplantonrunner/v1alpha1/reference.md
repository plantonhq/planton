# AwsPlantonRunner

> Generated from the protobuf schema by `make generate-reference` -- do not
> edit by hand. To change a fact on this page, change the proto field comment
> or validation rule it is derived from, then regenerate.

**apiVersion**: `aws.planton.dev/v1alpha1`

AwsPlantonRunnerSpec defines a standing Planton runner appliance inside
your AWS network: an always-on worker that receives deploy operations
from the Planton control plane and executes them from WITHIN the VPC.

Why it exists: some targets are reachable only from inside the network --
the canonical case is a Kubernetes cluster with a private API endpoint,
which no hosted runner fleet can reach. Placing a runner in the VPC makes
that cluster deployable and operable (day-2 updates, destroys, CloudOps)
with zero inbound exposure: the runner only ever dials OUT to the control
plane, so no ports open toward the internet.

The appliance is standing infrastructure, not a bootstrap step. It
survives rebuilds of the clusters it deploys to, which is exactly what
makes teardown orderly: in-cluster workloads are destroyed through the
runner, the cluster is destroyed by the AWS path, and the runner itself
is destroyed last.

ENROLLMENT IS TOKEN-FIRST: the runner is born with a runner TOKEN, never
an identity. On first boot it presents the token to the control plane,
registers ITSELF, and receives its own individually revocable identity.
Task replacement re-joins with the same token (the token's lineage
re-admits the runner it originally admitted -- no other token can). The
token lives in a module-created Secrets Manager secret; the ECS agent
injects it at task start, never as plaintext in any task definition.

The spec models intent -- where the runner lives (subnets), how big it is
(cpu/memory), which build it runs (runner_version), and the token it
joins with. The compute substrate is an implementation detail of the IaC
modules (see the kind's README); it deliberately has no
representation here.

## Example

```yaml
# Minimal AwsPlantonRunner manifest for local module testing. The token
# value below is an obviously-fake placeholder with the right shape --
# real deployments supply a managed-secret reference ($secret/<slug>)
# that the platform fills with a runner token before the infrastructure
# applies. The token authorizes joining and is never the runner's
# identity: the runner registers itself on first boot and receives its
# own individually revocable identity.
apiVersion: aws.planton.dev/v1alpha1
kind: AwsPlantonRunner
metadata:
  name: awsplantonrunner-demo
spec:
  region: us-west-2
  subnets:
    - value: subnet-0a1b2c3d4e5f60001
    - value: subnet-0a1b2c3d4e5f60002
  token: prt_FAKE_PLACEHOLDER_VALUE
```

## Spec Fields

| Path | Type | Required | Default | References |
|---|---|---|---|---|
| `spec.region` | `string` | yes |  |  |
| `spec.subnets` | `[]string \| valueFrom` | yes |  | AwsSubnet (`status.outputs.subnet_id`) |
| `spec.securityGroups` | `[]string \| valueFrom` |  |  | AwsSecurityGroup (`status.outputs.security_group_id`) |
| `spec.assignPublicIp` | `bool` |  |  |  |
| `spec.cpu` | `int32` |  | `512` |  |
| `spec.memory` | `int32` |  | `1024` |  |
| `spec.runnerVersion` | `string` |  | `latest` |  |
| `spec.imageRepository` | `string` |  | `ghcr.io/plantonhq/planton/runner` |  |
| `spec.controlPlaneEndpoint` | `string` |  |  |  |
| `spec.token` | `string` (sensitive) | yes |  |  |
| `spec.taskRole` | `string \| valueFrom` |  |  | AwsIamRole (`status.outputs.role_arn`) |
| `spec.logRetentionDays` | `int32` |  | `30` |  |

## Field Details

### spec.region

`string` · required

The AWS region the runner is deployed in. Deploy the runner in the
same region as the private endpoints it needs to reach.
Example: "us-west-2", "eu-west-1".

- rule: {"string":{"minLen":"1"}}

### spec.subnets

`[]string | valueFrom` · required

The subnets the runner's network interfaces are placed in -- this is
what puts the runner inside the network whose private endpoints it
must reach. Prefer private subnets with a NAT route (the runner needs
outbound internet to pull its image and dial the control plane); at
least two subnets in different availability zones lets the runner
reschedule across an AZ event. Public subnets work too when paired
with assign_public_ip. Reference AwsSubnet subnet_id outputs or pass
literal subnet IDs. All subnets must belong to the same VPC.

- references: AwsSubnet (`status.outputs.subnet_id`)
- rule: {"required":true}
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSubnet, name: <that resource's name>, fieldPath: status.outputs.subnet_id}} -- a bare string does not parse

### spec.securityGroups

`[]string | valueFrom`

Additional security groups attached to the runner's network
interfaces, BESIDE the outbound-only group the deployment creates for
itself. The runner accepts no inbound traffic, so extra groups are
only needed when a private target (a cluster API endpoint, a
database) admits traffic by source security group -- attach the group
that target trusts. Reference AwsSecurityGroup security_group_id
outputs or pass literal group IDs.

- references: AwsSecurityGroup (`status.outputs.security_group_id`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsSecurityGroup, name: <that resource's name>, fieldPath: status.outputs.security_group_id}} -- a bare string does not parse

### spec.assignPublicIp

`bool`

Assign a public IPv4 address to the runner's network interfaces.
Required when the subnets are PUBLIC (routed to an internet gateway)
and the VPC has no NAT gateway -- without it the runner cannot pull
its image or reach the control plane and never starts. Keep false for
private subnets with a NAT route (the recommended posture).

### spec.cpu

`int32` · optional (explicit presence)

CPU allocated to the runner, in AWS CPU units (1024 = 1 vCPU). The
default 512 (0.5 vCPU) comfortably runs the runner's control loops
plus typical IaC operations; size up for large stacks or high
operation concurrency. Serverless compute admits only these values:
256, 512, 1024, 2048, 4096, 8192, 16384 -- and each pairs with a
bounded memory range (see memory).

- default: `512`

### spec.memory

`int32` · optional (explicit presence)

Memory allocated to the runner, in MiB. The default 1024 pairs with
the default cpu of 512. Valid values depend on cpu -- for example cpu
512 admits 1024-4096 MiB in 1024 steps, cpu 1024 admits 2048-8192 --
and an invalid pairing is rejected by AWS only at deploy time, so the
spec validates the combination up front. Memory pressure shows up as
failed IaC operations mid-apply; when in doubt, size memory up before
cpu.

- default: `1024`

### spec.runnerVersion

`string` · optional (explicit presence)

The runner build to deploy: an image tag of the official runner
container image. "latest" tracks the newest release; pin a specific
version tag for change control. New tasks pull the tag on every
(re)start.

- default: `latest`

### spec.imageRepository

`string` · optional (explicit presence)

The container image repository the runner is pulled from. Override
only for air-gapped or mirrored registries hosting a copy of the
official image; the digest-identical mirror is your responsibility.

- default: `ghcr.io/plantonhq/planton/runner`

### spec.controlPlaneEndpoint

`string`

The control-plane endpoint the runner joins, as host:port. Leave
unset for Planton's hosted control plane (the runner's built-in
default); set it for a self-hosted instance (e.g.
"planton.example.com:443"). This is the one bootstrap coordinate the
join cannot deliver -- everything else (work queue, execution mode,
tunnel, API endpoints) arrives in the join response, so the runner
self-configures on arrival and no mode knob exists here.

- rule: control plane endpoint must be host:port, e.g. "planton.example.com:443" -- no scheme prefix
- rule: {"ignore":"IGNORE_IF_ZERO_VALUE"}

### spec.token

`string` · required · sensitive

The runner token that authorizes this runner to JOIN the control
plane. Create one with `planton runner token create` (or in the
console under Organization Settings -> Runner Tokens); on Planton, the
platform mints a token and writes it at exactly the managed-secret
reference this field names, before the infrastructure applies -- there
is no manual credential step. The token only gates joining and is
never the runner's identity: the runner receives its own individually
revocable identity when it registers itself on arrival, and revoking
this token never touches runners it already admitted. This is a
secret: supply it as a managed-secret reference, never inline
plaintext; it reaches the runner through Secrets Manager, not through
any launch configuration.

- rule: {"required":true}

### spec.taskRole

`string | valueFrom`

The IAM role the runner itself holds at runtime -- its AWS identity
when cloud operations or IaC runs use role-based (keyless) access
instead of injected keys. Compose an AwsIamRole with a trust policy
for "ecs-tasks.amazonaws.com" and exactly the permissions the
runner's workloads need, then reference its role_arn output. When
unset, the deployment creates a permissionless role so the identity
seam always exists and permissions can be granted later without
replacing the runner.

- references: AwsIamRole (`status.outputs.role_arn`)
- rule: write as {value: <literal>} or {valueFrom: {kind: AwsIamRole, name: <that resource's name>, fieldPath: status.outputs.role_arn}} -- a bare string does not parse

### spec.logRetentionDays

`int32` · optional (explicit presence)

Days the runner's logs are retained. The runner logs every operation
it executes -- these logs are the audit trail for what changed your
infrastructure, so retain them meaningfully. Must be one of the
retention periods CloudWatch supports (1, 3, 5, 7, 14, 30, 60, 90,
120, 150, 180, 365, ...).

- default: `30`

## Validation Rules

- `cpu_valid`: cpu must be one of the serverless compute sizes: 256, 512, 1024, 2048, 4096, 8192, or 16384 CPU units (1024 = 1 vCPU)
- `cpu_memory_combination`: cpu and memory must form a valid serverless compute pairing -- cpu 256 pairs with memory 512, 1024, or 2048; cpu 512 with 1024-4096 in steps of 1024; cpu 1024 with 2048-8192 in steps of 1024; cpu 2048 with 4096-16384 in steps of 1024; cpu 4096 with 8192-30720 in steps of 1024; cpu 8192 with 16384-61440 in steps of 4096; cpu 16384 with 32768-122880 in steps of 8192
- `log_retention_valid`: log_retention_days must be one of the retention periods CloudWatch supports: 1, 3, 5, 7, 14, 30, 60, 90, 120, 150, 180, 365, 400, 545, 731, 1096, 1827, 2192, 2557, 2922, 3288, or 3653 days

## Outputs

Reference an output from another manifest as `valueFrom: {kind: AwsPlantonRunner, name: <resource-name>, fieldPath: status.outputs.<output>}`.

| Output | Type | Description |
|---|---|---|
| `status.outputs.service_arn` | `string` | The ARN of the compute service keeping the runner running (e.g. "arn:aws:ecs:us-west-2:123456789012:service/<name>/<name>"). The primary handle for inspecting the appliance with AWS tooling. |
| `status.outputs.service_name` | `string` | The service's name (metadata.name). |
| `status.outputs.cluster_arn` | `string` | The ARN of the dedicated compute cluster the runner runs in. |
| `status.outputs.task_definition_arn` | `string` | The full task definition ARN (family:revision) of the running revision -- changes on every configuration or version change. |
| `status.outputs.log_group_name` | `string` | The CloudWatch log group carrying the runner's logs -- the audit trail of every operation the runner executed (e.g. "aws logs tail <log_group_name> --follow"). |
| `status.outputs.security_group_id` | `string` | The id of the outbound-only security group created for the runner. Private targets that admit traffic by source security group (a cluster API endpoint, a database) reference this id to trust the runner. |
| `status.outputs.execution_role_arn` | `string` | The ARN of the execution role -- the setup identity that pulls the runner image, writes its logs, and reads its token secret. |
| `status.outputs.task_role_arn` | `string` | The ARN of the runner's runtime IAM role -- the identity the runner holds while executing work. Grant this role permissions to let keyless cloud operations run through the runner (it is the referenced task_role when one was supplied, else the permissionless role created with the appliance). |
| `status.outputs.token_secret_arn` | `string` | The ARN of the Secrets Manager secret holding the runner token. The ECS agent injects it at task start; the token authorizes joining and is never the runner's identity. |
| `status.outputs.region` | `string` | The AWS region the runner was deployed in. Echoed so downstream tooling and verifiers can target the correct region. |
| `status.outputs.runner_name` | `string` | The name the runner registers itself under with the control plane -- the value shown by `planton runner list` the moment it joins. |

## References

Fields that can point at another resource's outputs:

| Field | Kind | Output |
|---|---|---|
| `spec.subnets` | AwsSubnet | `status.outputs.subnet_id` |
| `spec.securityGroups` | AwsSecurityGroup | `status.outputs.security_group_id` |
| `spec.taskRole` | AwsIamRole | `status.outputs.role_arn` |

## See Also

- [Overview](../README.md)
