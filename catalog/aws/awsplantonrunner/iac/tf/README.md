# Terraform Module to Deploy AwsPlantonRunner

This module provisions a standing Planton runner appliance on ECS
Fargate: the runner token in the `<name>-token` Secrets Manager secret
(zero-day recovery window -- the token is re-mintable credential
material -- injected into the container by the ECS agent at task
start), the execution role (image pull, logs, and a read grant scoped
to exactly the one secret), the runtime IAM role (the runner's own AWS
identity -- created permissionless, or the referenced `task_role`
passed through), an explicit-retention CloudWatch log group, an
outbound-only security group in the VPC derived from the first
referenced subnet, a dedicated ECS cluster, the task definition, and a
Fargate service holding exactly one runner.

The container's environment contract: `PLANTON_RUNNER_TOKEN` (resolved
from the secret by the ECS agent), `PLANTON_RUNNER_NAME` (the
registration name, `<env>-<metadata.name>` or `metadata.name` outside
an environment), `PLANTON_RUNNER_ENDPOINT` only when a control-plane
endpoint is declared (omitted, the runner's built-in hosted default
applies), `PORT` 50051, and `LOG_LEVEL` info. No execution-mode
variable exists: the runner self-configures its mode from the join
response.

The service deliberately does not gate on task steady state: the
runner's readiness contract is its work queue (operations wait there
until the worker polls), never ECS task liveness.

`variables.tf` reflects the proto schema for `AwsPlantonRunner`.

## Usage

Use the Planton CLI (tofu) with the default local backend:

```shell
planton tofu init --manifest e2e/manifest.yaml
planton tofu plan --manifest e2e/manifest.yaml
planton tofu apply --manifest e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest e2e/manifest.yaml --auto-approve
```

**Note**: AWS provider credentials are provided via IaC input (CLI), not in the manifest `spec`.

See [`e2e/manifest.yaml`](../../e2e/manifest.yaml) for a minimal test manifest.
