# AwsEcsTaskDefinition

Registers an ECS task definition -- the immutable, versioned blueprint of
the containers a task runs: images, ports, environment, secrets, health
checks, sizing, volumes, and the IAM identities the task assumes.

## Purpose

A task definition is to ECS what a launch template is to EC2 fleets: the
reusable template that everything which RUNS tasks references -- ECS
services for steady-state workloads, scheduled tasks, and one-off RunTask
calls. AWS makes revisions immutable (every change registers the next
revision of the family), and this component leans into that: the
`task_definition_arn` output carries the revision, so a referencing
`AwsEcsService` rolls automatically whenever a new revision registers.
"Change the image tag, the service rolls" is the composed behavior of the
resource graph, not deploy tooling.

## Key Features

- **Multi-container tasks** -- an application container plus sidecars
  (log routers, telemetry collectors, proxies), with per-container
  sizing, startup ordering (`dependsOn`), health checks, and in-place
  restart policies.
- **Secrets by reference** -- container `secrets` are Secrets Manager /
  SSM Parameter Store ARNs the ECS agent resolves at task start via the
  execution role; secret material never enters the task definition.
- **Secrets the module stores** -- container `secret_environment` values
  are kept in one Secrets Manager secret each
  (`<family>/<container>/<name>`), readable only by the execution role
  through the secret's resource policy, and injected by ARN pinned to the
  stored version, so a changed value registers a new revision.
- **Zero-configuration logging** -- one CloudWatch log group per family
  (`/ecs/<family>`, 30-day retention) with per-container stream
  prefixes; reference an existing `AwsCloudwatchLogGroup` or disable the
  default entirely. FireLens routing is first-class for everything else.
- **Honest identity split** -- the execution role (the agent's setup
  permissions) and the task role (the application's runtime permissions)
  are separate referenced `AwsIamRole` nodes, never one accumulated role.
- **Graviton and Windows** -- `runtimePlatform` selects ARM64 (~20%
  cheaper per vCPU on Fargate) or a Windows Server family.
- **The full volume family** -- EFS (durable shared storage with
  access-point pinning and IAM-authorized, transit-encrypted mounts),
  S3 buckets mounted as file systems (`s3files`), EC2 Docker volumes,
  host paths, and launch-time volumes (`configureAtLaunch`) whose
  managed-EBS backing the running `AwsEcsService` supplies per task.
- **EC2 namespace and placement control** -- `ipcMode` / `pidMode`
  namespace sharing and task-level `placementConstraints` for EC2 tasks
  (CEL guards the Fargate rules); `pidMode: task` also works on Fargate
  for sidecar process observation.
- **Fault injection opt-in** -- `enableFaultInjection` lets AWS FIS chaos
  experiments target the task's containers; off by default.

## Example

```yaml
apiVersion: aws.planton.dev/v1alpha1
kind: AwsEcsTaskDefinition
metadata:
  name: api
spec:
  region: us-west-2
  cpu: 512
  memory: 1024
  executionRole:
    valueFrom:
      kind: AwsIamRole
      name: api-execution
      fieldPath: status.outputs.role_arn
  containers:
    - name: app
      image: 123456789012.dkr.ecr.us-west-2.amazonaws.com/api:1.4.2
      portMappings:
        - containerPort: 8080
          name: http
          appProtocol: http
      healthCheck:
        command: ["CMD-SHELL", "curl -f http://localhost:8080/healthz || exit 1"]
```

Deploy with:

```shell
planton apply -f task-definition.yaml
```

Both a Pulumi module and a Terraform/OpenTofu module implement this
component at full behavioral parity; the provisioner is an execution
detail.

---

© Planton. Licensed under [Apache-2.0](https://github.com/plantonhq/planton/blob/main/LICENSE).
