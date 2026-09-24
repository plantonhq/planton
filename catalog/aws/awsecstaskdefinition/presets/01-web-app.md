# Web Application

This preset defines a Fargate web application task: one container with a
named HTTP port (ready for load-balancer and Service Connect wiring),
secrets injected by reference at task start, an in-container health check,
and a read-only root filesystem. Reference this task definition's
`task_definition_arn` output from an `AwsEcsService` -- each new revision
(a new image tag) changes the output and rolls the service.

## When to Use

- HTTP/API services running on Fargate behind an ALB target group
- Any single-container workload whose deploys should roll through the
  resource graph (new revision in, service rolls)

## Key Configuration Choices

- **`cpu: 512` / `memory: 1024`** -- a half-vCPU Fargate task size that
  fits most web workloads; valid pairings are constrained by AWS (512 CPU
  pairs with 1024-4096 MiB)
- **Named port with `appProtocol: http`** -- the name is the join key for
  Service Connect, and the protocol unlocks per-request telemetry
- **`secrets` by ARN** -- the ECS agent resolves Secrets Manager / SSM
  references at task start via the execution role; no secret material
  lives in the task definition
- **`secretEnvironment` for a Planton secret** -- the component stores the
  value in a Secrets Manager secret only the execution role can read and
  injects it by ARN; a `$secret/` reference in `environment` would be
  registered in plain text, so the platform refuses it
- **Two roles by design** -- the execution role pulls images and writes
  logs; the task role is the application's own AWS identity
- **`readonlyRootFilesystem: true`** -- a strong hardening default for
  stateless services; writable paths come from volumes when needed
- **Default logging left on** -- the module creates `/ecs/<family>` with
  30-day retention and streams each container under its own prefix

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<service-name>` | The task-definition family name | Your service's name (e.g., `api`) |
| `<aws-region>` | AWS region code (e.g., `us-east-1`) | Your deployment region |
| `<account-id>` | Your AWS account ID | AWS console |
| `<repository>:<tag>` | The ECR repository and image tag | Your ECR console or CI pipeline |
| `<execution-role-resource-name>` | Name of the AwsIamRole for the ECS agent | Your role manifest's `metadata.name` |
| `<task-role-resource-name>` | Name of the AwsIamRole for the application | Your role manifest's `metadata.name` |
| `<secret-name>` | The Secrets Manager secret holding the value | Secrets Manager console |
| `<payments-api-key-slug>` | The Planton secret holding the payments API key | `planton secret list -o json` |

## Common Additions

- `runtimePlatform.cpuArchitecture: ARM64` to run on Graviton (~20%
  cheaper per vCPU; images must be arm64-built)
- `ephemeralStorageGib` when the workload needs more than the free 20 GiB
- A FireLens sidecar (`firelensConfiguration`) to route logs to a
  third-party destination
