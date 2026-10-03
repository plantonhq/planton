# AWS Planton Runner

Deploys a standing Planton runner appliance inside your AWS VPC -- an always-on, outbound-only worker that receives deploy operations from the control plane and executes them from within the network. It is the piece that makes private endpoints (most notably private EKS cluster APIs) deployable and operable, with subnets, security groups, and the runtime role wired through ValueFromRef.

## What Gets Created

When you deploy this Infra Component, the IaC module provisions:

- **Secrets Manager secret** -- holds the runner token; the container reads it at start through native secret injection, so the token never appears in the task definition
- **IAM execution role** -- the setup identity: pulls the runner image, writes logs, and reads exactly the one token secret
- **IAM runtime role** -- the runner's own AWS identity while executing work; created permissionless, and only when `taskRole` does not reference an existing role
- **CloudWatch log group** -- the audit trail of every operation the runner executes, with configurable retention
- **Security group** -- outbound-only, no inbound rules, created in the VPC derived from the referenced subnets
- **ECS cluster** -- dedicated to this runner, keeping the appliance's blast radius and teardown self-contained
- **ECS task definition and Fargate service** -- the runner container itself, kept at exactly one copy and restarted automatically on exit
- **AWS Tags** -- resource tags derived from the resource's organization, environment, and name

## Before You Deploy

### Planton Setup

- **AWS Provider Connection** -- the credential used to provision the appliance itself. Required.
- **Runner token** -- nothing to create by hand: the platform mints a runner token and writes it at exactly the managed-secret reference the manifest declares, before the infrastructure applies. Choose a secret slug and reference it as `$secret/<slug>` in the `token` field; plaintext is rejected. The token only authorizes joining -- the runner registers itself on first boot and receives its own individually revocable identity, and revoking the token never touches runners it already admitted. (Self-service alternative: `planton runner token create`, or the console under Organization Settings → Runner Tokens.)

### AWS Account

- **Subnets** -- the VPC placement decision. Prefer private subnets with a NAT route (the runner needs outbound internet for its image and the control plane) across two Availability Zones; public subnets work with `assignPublicIp: true`. Reference `AwsSubnet` resources via ValueFromRef or pass literal subnet IDs.
- **Security groups (optional)** -- only when a private target (a cluster API endpoint, a database) admits traffic by source security group; attach the group that target trusts.
- **IAM role (optional)** -- a role trusted by `ecs-tasks.amazonaws.com` when the runner needs AWS permissions of its own for keyless cloud access.

## Deploy

### Console

Open the deployment store, find **AWS Planton Runner**, and click **Deploy**. The creation wizard walks you through preset selection, environment and connection configuration, and spec fields. Start from the **Private VPC Worker** preset in the [Presets](#presets) tab to pre-populate a working configuration.

### CLI

Create a manifest and apply it:

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsPlantonRunner
metadata:
  name: vpc-runner
  org: acme-corp
  env: prod
spec:
  region: us-west-2
  subnets:
    - valueFrom:
        kind: AwsSubnet
        name: private-a
        fieldPath: status.outputs.subnet_id
    - valueFrom:
        kind: AwsSubnet
        name: private-b
        fieldPath: status.outputs.subnet_id
  token: $secret/vpc-runner-token
```

```shell
planton apply -f runner.yaml
```

This minimal manifest deploys a pull-based worker at the default sizing (0.5 vCPU, 1 GiB) tracking the latest runner release, with a permissionless runtime role and 30-day log retention -- sizing, version pinning, and the runtime identity are not configured. The runner registers itself as `prod-vpc-runner` (`<env>-<metadata.name>`) the moment it joins. An Infra Job tracks the provisioning in real time.

### InfraChart

When the runner deploys alongside its network and identity in one chart, wire the subnet, security group, and role references via ValueFromRef:

```yaml
spec:
  subnets:
    - valueFrom:
        kind: AwsSubnet
        name: private-a
        fieldPath: status.outputs.subnet_id
  securityGroups:
    - valueFrom:
        kind: AwsSecurityGroup
        name: eks-api-clients
        fieldPath: status.outputs.security_group_id
  taskRole:
    valueFrom:
      kind: AwsIamRole
      name: runner-runtime
      fieldPath: status.outputs.role_arn
```

The InfraPipeline resolves the dependency graph, deploys the subnets, security groups, and IAM role first, then provisions the runner with the resolved values.

## Key Configuration

These are the most important decisions when configuring the runner. Explore the full field reference in the [API Explorer](#api-explorer) tab.

**Subnet placement** -- `subnets` is the whole point of the appliance: whatever those subnets' route tables can reach, the runner can deploy to. All subnets must belong to the same VPC; two Availability Zones let the service reschedule across an AZ event.

**Control-plane endpoint** -- `controlPlaneEndpoint` (host:port) is only for self-hosted control planes; leave it unset for Planton's hosted endpoint. It is the one bootstrap coordinate the join cannot deliver -- everything else arrives in the join response, so the runner self-configures its execution mode on arrival and no mode knob exists.

**Sizing** -- `cpu` and `memory` must form one of the fixed serverless pairings (validated up front; AWS would otherwise reject them only at deploy time). The 512/1024 default handles typical IaC operations; memory pressure shows up as failed operations mid-apply, so size memory up before CPU.

**Runner build** -- empty `runnerVersion` tracks the newest release on every task (re)start; pin a version tag for change control. `imageRepository` is only for air-gapped or mirrored registries hosting a digest-identical copy.

**Runtime identity** -- leave `taskRole` empty to get a permissionless role (the identity seam always exists, so permissions can be granted later without replacing the runner), or reference an `AwsIamRole` composed with exactly the permissions keyless cloud access needs.

## Outputs and Dependencies

### What This Kind Consumes

| Dependency | Field | ValueFromRef Path |
|------------|-------|-------------------|
| **AwsSubnet** | `subnets` | `status.outputs.subnet_id` |
| **AwsSecurityGroup** (optional) | `securityGroups` | `status.outputs.security_group_id` |
| **AwsIamRole** (optional) | `taskRole` | `status.outputs.role_arn` |

### What This Kind Provides

After provisioning, `status.outputs` contains values that downstream Infra Components can consume via ValueFromRef:

| Output | Description | Common Downstream Use |
|--------|-------------|----------------------|
| `security_group_id` | The outbound-only group's id | Private targets reference it to trust the runner (e.g. an EKS API's allowed sources) |
| `task_role_arn` | The runtime identity -- the referenced `taskRole` when one was supplied, else the permissionless role created with the appliance | Granting the runner AWS permissions for keyless operations |
| `log_group_name` | The CloudWatch log group | Tailing the operation audit trail (`aws logs tail <name> --follow`) |
| `token_secret_arn` | The token secret's ARN -- the token authorizes joining and is never the runner's identity | Auditing secret access; rotation tooling |
| `runner_name` | The name the runner registers itself under with the control plane | Console and `planton runner list` lookups |

The remaining outputs (`service_arn`, `service_name`, `cluster_arn`, `task_definition_arn`, `execution_role_arn`, `region`) echo the appliance's own AWS identifiers for inspection with AWS tooling; downstream Infra Components have no real use for them.

## Common Patterns

Browse the [Presets](#presets) tab for ready-to-deploy configurations.

**Private VPC worker** -- the standard appliance: a pull-based worker on two private subnets that makes a private-endpoint cluster deployable. Start from the **Private VPC Worker** preset.

**Public subnet worker** -- the same worker on public subnets with `assignPublicIp: true`, for VPCs with an internet gateway but no NAT; still zero inbound exposure. Start from the **Public Subnet Worker** preset.

**High-capacity deploy worker** -- pinned version and larger sizing for heavy Terraform/Pulumi stacks and concurrent operations. Start from the **High Capacity (Production Hardened)** preset.

## Works With

- [**AWS Subnet**](/infra-catalog/aws-subnet) -- where the runner's network interfaces live; the placement that defines what it can reach
- [**AWS Security Group**](/infra-catalog/aws-security-group) -- extra groups a private target trusts, attached beside the created outbound-only group
- [**AWS IAM Role**](/infra-catalog/aws-iam-role) -- the runner's runtime identity, composed first-class with exactly the permissions its workloads need
